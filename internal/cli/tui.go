package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/uaclient"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const maxTUILogLines = 200

type tuiNode struct {
	ID       string
	Label    string
	Summary  string
	CanRead  bool
	Children bool
}

type tuiAttribute struct {
	Name  string
	Value string
}

type tuiValue struct {
	ID    string
	Label string
	Value string
}

type tuiBackend interface {
	Children(ctx context.Context, nodeID string) ([]tuiNode, error)
	Details(ctx context.Context, nodeID string) ([]tuiAttribute, error)
	Read(ctx context.Context, nodeID string) (tuiValue, error)
	Watch(ctx context.Context, nodeIDs []string, interval time.Duration) (<-chan tuiValue, <-chan error, func(), error)
}

type uaTUIBackend struct {
	service *uaclient.Service
	timeout time.Duration
}

func (b *uaTUIBackend) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if b.timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, b.timeout)
}

func (b *uaTUIBackend) Children(ctx context.Context, nodeID string) ([]tuiNode, error) {
	callCtx, cancel := b.withTimeout(ctx)
	defer cancel()

	rows, err := b.service.Browse(callCtx, nodeID, 0)
	if err != nil {
		return nil, err
	}
	nodes := make([]tuiNode, 0, len(rows))
	for _, row := range rows {
		nodes = append(nodes, uaNodeToTUI(row))
	}
	sort.Slice(nodes, func(i, j int) bool {
		return strings.ToLower(nodes[i].Label) < strings.ToLower(nodes[j].Label)
	})
	return nodes, nil
}

func (b *uaTUIBackend) Details(ctx context.Context, nodeID string) ([]tuiAttribute, error) {
	callCtx, cancel := b.withTimeout(ctx)
	defer cancel()

	result, err := b.service.Attributes(callCtx, nodeID)
	if err != nil {
		return nil, err
	}
	attrs := make([]tuiAttribute, 0, len(result.Attributes)+1)
	attrs = append(attrs, tuiAttribute{Name: "NodeID", Value: result.NodeID})
	for _, attr := range result.Attributes {
		attrs = append(attrs, tuiAttribute{
			Name:  attr.Name,
			Value: fmt.Sprintf("%v (%s)", attr.Value, attr.Status),
		})
	}
	return attrs, nil
}

func (b *uaTUIBackend) Read(ctx context.Context, nodeID string) (tuiValue, error) {
	callCtx, cancel := b.withTimeout(ctx)
	defer cancel()

	row, err := b.service.Read(callCtx, nodeID)
	if err != nil {
		return tuiValue{}, err
	}
	return tuiValue{ID: row.NodeID, Label: row.NodeID, Value: fmt.Sprint(row.Value)}, nil
}

func (b *uaTUIBackend) Watch(ctx context.Context, nodeIDs []string, interval time.Duration) (<-chan tuiValue, <-chan error, func(), error) {
	sub, err := b.service.Monitor(ctx, nodeIDs, interval)
	if err != nil {
		return nil, nil, nil, err
	}
	values := make(chan tuiValue, 32)
	errs := make(chan error, 8)
	go func() {
		defer close(values)
		defer close(errs)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-sub.Events:
				if !ok {
					return
				}
				values <- tuiValue{ID: event.NodeID, Label: event.NodeID, Value: fmt.Sprint(event.Value)}
			case err, ok := <-sub.Errors:
				if !ok {
					return
				}
				errs <- err
			}
		}
	}()
	return values, errs, sub.Close, nil
}

func uaNodeToTUI(row domain.Node) tuiNode {
	label := firstNonEmpty(row.DisplayName, row.BrowseName, row.NodeID)
	return tuiNode{
		ID:       row.NodeID,
		Label:    label,
		Summary:  strings.TrimSpace(strings.Join([]string{row.NodeClass, row.DataType}, " ")),
		CanRead:  row.NodeClass == "Variable",
		Children: row.NodeClass != "Variable",
	}
}

type tuiController struct {
	backend   tuiBackend
	interval  time.Duration
	monitored map[string]tuiValue
	logs      []string

	watchCancel context.CancelFunc
	watchStop   func()
	mu          sync.Mutex
}

func newTUIController(backend tuiBackend, interval time.Duration) *tuiController {
	return &tuiController{
		backend:   backend,
		interval:  interval,
		monitored: map[string]tuiValue{},
	}
}

func (c *tuiController) addLog(line string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logs = append(c.logs, line)
	if len(c.logs) > maxTUILogLines {
		c.logs = append([]string(nil), c.logs[len(c.logs)-maxTUILogLines:]...)
	}
}

func (c *tuiController) logText() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.logs, "\n")
}

func (c *tuiController) setMonitored(value tuiValue) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.monitored[value.ID] = value
}

func (c *tuiController) removeMonitored(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.monitored, id)
}

func (c *tuiController) monitoredIDs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	ids := make([]string, 0, len(c.monitored))
	for id := range c.monitored {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (c *tuiController) monitoredValues() []tuiValue {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]tuiValue, 0, len(c.monitored))
	for _, value := range c.monitored {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].Label < values[j].Label
	})
	return values
}

func (c *tuiController) stopWatch() {
	if c.watchCancel != nil {
		c.watchCancel()
		c.watchCancel = nil
	}
	if c.watchStop != nil {
		c.watchStop()
		c.watchStop = nil
	}
}

func (c *tuiController) restartWatch(ctx context.Context, onValue func(tuiValue), onError func(error)) error {
	c.stopWatch()
	ids := c.monitoredIDs()
	if len(ids) == 0 {
		return nil
	}
	watchCtx, cancel := context.WithCancel(ctx)
	values, errs, stop, err := c.backend.Watch(watchCtx, ids, c.interval)
	if err != nil {
		cancel()
		return err
	}
	c.watchCancel = cancel
	c.watchStop = stop
	go func() {
		for {
			select {
			case <-watchCtx.Done():
				return
			case value, ok := <-values:
				if !ok {
					return
				}
				c.setMonitored(value)
				onValue(value)
			case err, ok := <-errs:
				if !ok {
					return
				}
				onError(err)
			}
		}
	}()
	return nil
}

func (a *App) tui(args []string) error {
	fs := a.newFlagSet("tui")
	common := commonOptions{}
	addCommonFlagsWithoutFormat(fs, &common)
	root := fs.String("node", "i=84", "root node id")
	interval := fs.Duration("interval", time.Second, "monitor subscription interval")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *interval <= 0 {
		return errors.New("--interval must be greater than zero")
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer cancel()
	service := uaclient.NewService(common.client)
	if err := service.Connect(connectCtx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	backend := &uaTUIBackend{service: service, timeout: common.client.Timeout}
	return runTUI(context.Background(), "OPC UA Browser", tuiNode{ID: *root, Label: *root, Children: true}, backend, *interval)
}

func runTUI(ctx context.Context, title string, root tuiNode, backend tuiBackend, interval time.Duration) error {
	app := tview.NewApplication()
	controller := newTUIController(backend, interval)
	tree := tview.NewTreeView()
	details := tview.NewTable().SetBorders(false)
	monitored := tview.NewTable().SetBorders(false)
	logView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	footer := tview.NewTextView().SetDynamicColors(true).SetText("Enter:Expand  tab:Next  a:Attributes  r:Read  m:Monitor  u:Unmonitor  R:Reload  c:Clear  ?:Help  q:Exit")

	rootNode := tview.NewTreeNode(root.Label).SetReference(root).SetColor(tcell.ColorGreen)
	tree.SetRoot(rootNode).SetCurrentNode(rootNode)
	styleBox(tree.Box, "Address Space")
	styleBox(details.Box, "Attribute List")
	styleBox(monitored.Box, "Monitored Items")
	styleBox(logView.Box, "Info")

	controller.addLog(title + " ready")
	refreshLog := func() {
		logView.SetText(controller.logText())
		logView.ScrollToEnd()
	}
	refreshLog()

	refreshMonitored := func() {
		monitored.Clear()
		monitored.SetCell(0, 0, tview.NewTableCell("Item").SetTextColor(tcell.ColorAqua).SetSelectable(false))
		monitored.SetCell(0, 1, tview.NewTableCell("Value").SetTextColor(tcell.ColorAqua).SetSelectable(false))
		for i, value := range controller.monitoredValues() {
			monitored.SetCell(i+1, 0, tview.NewTableCell(value.Label))
			monitored.SetCell(i+1, 1, tview.NewTableCell(value.Value))
		}
	}
	refreshMonitored()

	showDetails := func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		ref, ok := node.GetReference().(tuiNode)
		if !ok {
			return
		}
		attrs, err := backend.Details(ctx, ref.ID)
		if err != nil {
			controller.addLog("attributes " + ref.ID + ": " + err.Error())
			refreshLog()
			return
		}
		details.Clear()
		for i, attr := range attrs {
			details.SetCell(i, 0, tview.NewTableCell(attr.Name).SetTextColor(tcell.ColorAqua))
			details.SetCell(i, 1, tview.NewTableCell(attr.Value))
		}
		controller.addLog("attributes refreshed for " + ref.ID)
		refreshLog()
	}

	loadChildren := func(node *tview.TreeNode, force bool) {
		if node == nil {
			return
		}
		ref, ok := node.GetReference().(tuiNode)
		if !ok {
			return
		}
		if !force && len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}
		children, err := backend.Children(ctx, ref.ID)
		if err != nil {
			controller.addLog("browse " + ref.ID + ": " + err.Error())
			refreshLog()
			return
		}
		node.ClearChildren()
		for _, child := range children {
			label := child.Label
			if child.Summary != "" {
				label += "  " + child.Summary
			}
			childNode := tview.NewTreeNode(label).SetReference(child)
			if child.Children {
				childNode.SetColor(tcell.ColorWhite)
			}
			node.AddChild(childNode)
		}
		node.SetExpanded(true)
		controller.addLog(fmt.Sprintf("loaded %d children for %s", len(children), ref.ID))
		refreshLog()
	}

	readSelected := func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		ref, ok := node.GetReference().(tuiNode)
		if !ok {
			return
		}
		value, err := backend.Read(ctx, ref.ID)
		if err != nil {
			controller.addLog("read " + ref.ID + ": " + err.Error())
			refreshLog()
			return
		}
		controller.addLog(fmt.Sprintf("value %s changed/read as %s", value.Label, value.Value))
		refreshLog()
	}

	monitorSelected := func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		ref, ok := node.GetReference().(tuiNode)
		if !ok {
			return
		}
		controller.setMonitored(tuiValue{ID: ref.ID, Label: firstNonEmpty(ref.Label, ref.ID), Value: "<waiting>"})
		refreshMonitored()
		if err := controller.restartWatch(ctx, func(value tuiValue) {
			app.QueueUpdateDraw(func() {
				refreshMonitored()
				controller.addLog("value " + value.Label + " changed to " + value.Value)
				refreshLog()
			})
		}, func(err error) {
			app.QueueUpdateDraw(func() {
				controller.addLog("monitor: " + err.Error())
				refreshLog()
			})
		}); err != nil {
			controller.addLog("monitor " + ref.ID + ": " + err.Error())
			refreshLog()
			return
		}
		controller.addLog("monitoring " + ref.ID)
		refreshLog()
	}

	unmonitorSelected := func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		ref, ok := node.GetReference().(tuiNode)
		if !ok {
			return
		}
		controller.removeMonitored(ref.ID)
		refreshMonitored()
		if err := controller.restartWatch(ctx, func(value tuiValue) {
			app.QueueUpdateDraw(func() {
				refreshMonitored()
				controller.addLog("value " + value.Label + " changed to " + value.Value)
				refreshLog()
			})
		}, func(err error) {
			app.QueueUpdateDraw(func() {
				controller.addLog("monitor: " + err.Error())
				refreshLog()
			})
		}); err != nil {
			controller.addLog("monitor restart: " + err.Error())
		}
		controller.addLog("unmonitored " + ref.ID)
		refreshLog()
	}

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		loadChildren(node, false)
		showDetails(node)
	})
	tree.SetChangedFunc(showDetails)

	right := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(details, 0, 1, false).
		AddItem(monitored, 0, 1, false)
	top := tview.NewFlex().
		AddItem(tree, 0, 2, true).
		AddItem(right, 0, 3, false)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(top, 0, 4, true).
		AddItem(logView, 0, 2, false).
		AddItem(footer, 1, 0, false)

	focusables := []tview.Primitive{tree, details, monitored, logView}
	focusIndex := 0
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		node := tree.GetCurrentNode()
		switch event.Key() {
		case tcell.KeyCtrlC:
			controller.stopWatch()
			app.Stop()
			return nil
		case tcell.KeyTab:
			focusIndex = (focusIndex + 1) % len(focusables)
			app.SetFocus(focusables[focusIndex])
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				controller.stopWatch()
				app.Stop()
				return nil
			case 'a':
				showDetails(node)
				return nil
			case 'r':
				readSelected(node)
				return nil
			case 'm':
				monitorSelected(node)
				return nil
			case 'u':
				unmonitorSelected(node)
				return nil
			case 'R':
				loadChildren(node, true)
				return nil
			case 'c':
				controller.logs = nil
				refreshLog()
				return nil
			case '?':
				controller.addLog("keys: arrows/Enter expand, tab focus, a attributes, r read, m monitor, u unmonitor, R reload, c clear, q exit")
				refreshLog()
				return nil
			}
		}
		return event
	})

	loadChildren(rootNode, true)
	showDetails(rootNode)
	if err := app.SetRoot(layout, true).SetFocus(tree).Run(); err != nil {
		controller.stopWatch()
		return err
	}
	controller.stopWatch()
	return nil
}

func styleBox(box *tview.Box, title string) {
	box.SetBorder(true).SetTitle(" " + title + " ").SetTitleColor(tcell.ColorAqua)
}
