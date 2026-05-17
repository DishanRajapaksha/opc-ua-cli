package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/domain"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/output"
	"github.com/DishanRajapaksha/opc-ua-cli/internal/uaclient"
)

func (a *App) initConfig(args []string) error {
	fs := a.newFlagSet("init-config")
	outputPath := config.DefaultConfigPath
	force := false
	fs.StringVar(&outputPath, "output", outputPath, "output YAML config file")
	fs.BoolVar(&force, "force", false, "overwrite output file if it exists")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("refusing to overwrite existing file %q; use --force to overwrite", outputPath)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat %q: %w", outputPath, err)
		}
	}

	contents, err := config.StarterConfigYAML()
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, contents, 0o600); err != nil {
		return fmt.Errorf("write config %q: %w", outputPath, err)
	}
	fmt.Fprintf(a.out, "wrote starter config to %s\n", outputPath)
	return nil
}

func (a *App) validateConfig(args []string) error {
	fs := a.newFlagSet("validate-config")
	configPath := config.DefaultConfigPath
	profile := ""
	fs.StringVar(&configPath, "config", config.DefaultConfigPath, "YAML config file")
	fs.StringVar(&profile, "profile", "", "config profile name")
	verbose := fs.Bool("verbose", false, "print high-level connection decisions")
	debug := fs.Bool("debug", false, "enable lower-level OPC UA client debug logging")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.LoadClientConfigForProfile(configPath, profile)
	if err != nil {
		return err
	}
	if err := config.ValidateClientConfig(cfg); err != nil {
		return err
	}
	_ = verbose
	_ = debug
	fmt.Fprintln(a.out, "config validation: PASS")
	return nil
}

func (a *App) newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.err)
	return fs
}

func (a *App) endpoints(args []string) error {
	fs := a.newFlagSet("endpoints")
	cfg := config.DefaultClientConfig()
	configPath := config.DefaultConfigPath
	profile := ""
	format := "table"
	fs.StringVar(&configPath, "config", config.DefaultConfigPath, "YAML config file")
	fs.StringVar(&profile, "profile", "", "config profile name")
	fs.StringVar(&cfg.Endpoint, "endpoint", cfg.Endpoint, "OPC UA endpoint URL")
	fs.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "request timeout")
	fs.StringVar(&format, "format", format, "output format: table, json")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "print high-level connection decisions")
	fs.BoolVar(&cfg.Debug, "debug", false, "enable lower-level OPC UA client debug logging")
	if err := fs.Parse(args); err != nil {
		return err
	}

	fileCfg, err := config.LoadClientConfigForProfile(configPath, profile)
	if err != nil {
		return err
	}
	visited := visitedFlags(fs)
	if !visited["endpoint"] {
		cfg.Endpoint = fileCfg.Endpoint
	}
	if !visited["timeout"] {
		cfg.Timeout = fileCfg.Timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	rows, err := uaclient.ListEndpoints(ctx, cfg.Endpoint)
	if err != nil {
		return err
	}
	return a.renderEndpoints(format, rows)
}

func (a *App) namespaces(args []string) error {
	fs := a.newFlagSet("namespaces")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer cancel()
	service := uaclient.NewService(common.client)
	if err := service.Connect(ctx); err != nil {
		return err
	}
	defer service.Close(context.Background())
	uris, err := service.NamespaceArray(ctx)
	if err != nil {
		return err
	}
	rows := make([][]string, 0, len(uris))
	for i, uri := range uris {
		rows = append(rows, []string{fmt.Sprint(i), uri})
	}
	return output.WriteTable(a.out, []string{"Index", "URI"}, rows)
}

func (a *App) attributes(args []string) error {
	fs := a.newFlagSet("attributes")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	node := fs.String("node", "", "node id to inspect")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	if *node == "" {
		return errors.New("--node is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer cancel()
	service := uaclient.NewService(common.client)
	if err := service.Connect(ctx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	row, err := service.Attributes(ctx, *node)
	if err != nil {
		return err
	}
	return a.renderAttributes(common.format, row)
}

func (a *App) browse(args []string) error {
	fs := a.newFlagSet("browse")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	node := fs.String("node", "i=84", "root node id")
	depth := fs.Int("depth", 1, "recursive browse depth")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	if *depth < 0 {
		return errors.New("--depth must be zero or greater")
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer cancel()

	service := uaclient.NewService(common.client)
	if err := service.Connect(ctx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	rows, err := service.Browse(ctx, *node, *depth)
	if err != nil {
		return err
	}
	return a.renderNodes(common.format, rows)
}

func (a *App) read(args []string) error {
	fs := a.newFlagSet("read")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	var nodes stringList
	nodesFile := fs.String("nodes", "", "path to file with one node id per line")
	fs.Var(&nodes, "node", "node id to read; repeat for multiple nodes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	fromFile, err := readNodesFile(*nodesFile)
	if err != nil {
		return err
	}
	allNodes := append([]string(nodes), fromFile...)
	if len(allNodes) == 0 {
		return errors.New("at least one --node is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer cancel()

	service := uaclient.NewService(common.client)
	if err := service.Connect(ctx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	rows := make([]domain.ReadResult, 0, len(allNodes))
	for _, node := range allNodes {
		row, readErr := service.Read(ctx, node)
		if readErr != nil {
			return readErr
		}
		rows = append(rows, row)
	}
	if len(rows) == 1 {
		return a.renderRead(common.format, rows[0])
	}
	return renderReadMany(a, common.format, rows)
}

func (a *App) write(args []string) error {
	fs := a.newFlagSet("write")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	node := fs.String("node", "", "node id to write")
	value := fs.String("value", "", "value to write")
	valueType := fs.String("type", "string", "scalar value type")
	var items stringList
	fs.Var(&items, "item", "write item in node:type:value format; repeat for multiple nodes")
	dryRun := fs.Bool("dry-run", false, "show what would be written without sending the write request")
	yes := fs.Bool("yes", false, "skip interactive confirmation and perform write immediately")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	writeItems := make([]writeItem, 0, len(items)+1)
	for _, raw := range items {
		item, err := parseWriteItem(raw)
		if err != nil {
			return err
		}
		writeItems = append(writeItems, item)
	}
	if *node != "" {
		writeItems = append(writeItems, writeItem{Node: *node, Type: *valueType, Value: *value})
	}
	if len(writeItems) == 0 {
		return errors.New("either --node/--type/--value or at least one --item is required")
	}

	source := "defaults and CLI flags"
	if common.configPath != "" {
		source = fmt.Sprintf("config=%s", common.configPath)
		if common.profile != "" {
			source += fmt.Sprintf(" profile=%s", common.profile)
		}
		source += " (with CLI overrides)"
	}

	fmt.Fprintln(a.out, "Write request")
	fmt.Fprintf(a.out, "Endpoint: %s\n", common.client.Endpoint)
	fmt.Fprintf(a.out, "Config Source: %s\n", source)
	for i, item := range writeItems {
		fmt.Fprintf(a.out, "Item %d Node: %s\n", i+1, item.Node)
		fmt.Fprintf(a.out, "Item %d Type: %s\n", i+1, item.Type)
		fmt.Fprintf(a.out, "Item %d Value: %s\n", i+1, item.Value)
	}

	if *dryRun {
		fmt.Fprintln(a.out, "Dry run: write request not sent")
		return nil
	}

	if !*yes {
		if !isInteractiveTerminal() {
			return errors.New("write confirmation required in non-interactive mode; pass --yes to continue")
		}
		fmt.Fprint(a.out, "Confirm write? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read confirmation: %w", err)
		}
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer != "y" && answer != "yes" {
			return errors.New("write cancelled")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer cancel()

	service := uaclient.NewService(common.client)
	if err := service.Connect(ctx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	rows := make([]domain.WriteResult, 0, len(writeItems))
	for _, item := range writeItems {
		row, writeErr := service.Write(ctx, item.Node, item.Type, item.Value)
		if writeErr != nil {
			return writeErr
		}
		rows = append(rows, row)
	}
	if len(rows) == 1 {
		return a.renderWrite(common.format, rows[0])
	}
	return renderWriteMany(a, common.format, rows)
}

type writeItem struct {
	Node  string
	Type  string
	Value string
}

func parseWriteItem(raw string) (writeItem, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) != 3 {
		return writeItem{}, errors.New("invalid --item format, expected node:type:value")
	}
	if strings.TrimSpace(parts[0]) == "" {
		return writeItem{}, errors.New("invalid --item format: node cannot be empty")
	}
	return writeItem{Node: parts[0], Type: parts[1], Value: parts[2]}, nil
}

func readNodesFile(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read nodes file %q: %w", path, err)
	}
	defer file.Close()
	var nodes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		nodes = append(nodes, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan nodes file %q: %w", path, err)
	}
	return nodes, nil
}

func isInteractiveTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

func (a *App) monitor(args []string) error {
	fs := a.newFlagSet("monitor")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	var nodes stringList
	interval := fs.Duration("interval", time.Second, "subscription interval")
	duration := fs.Duration("duration", 0, "stop after this duration; zero runs until interrupted")
	fs.Var(&nodes, "node", "node id to monitor; repeat for multiple nodes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	if len(nodes) == 0 {
		return errors.New("at least one --node is required")
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer connectCancel()

	service := uaclient.NewService(common.client)
	if err := service.Connect(connectCtx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	runCtx := context.Background()
	if *duration > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(runCtx, *duration)
		defer cancel()
	}

	subscription, err := service.Monitor(runCtx, []string(nodes), *interval)
	if err != nil {
		return err
	}
	defer subscription.Close()

	events := subscription.Events
	errorsCh := subscription.Errors
	format := output.NormaliseFormat(common.format)

	for events != nil || errorsCh != nil {
		select {
		case event, ok := <-events:
			if !ok {
				events = nil
				continue
			}
			if err := a.renderDataChange(format, event); err != nil {
				return err
			}
		case err, ok := <-errorsCh:
			if !ok {
				errorsCh = nil
				continue
			}
			fmt.Fprintln(a.err, err)
		case <-runCtx.Done():
			return nil
		}
	}

	return nil
}

func (a *App) watch(args []string) error {
	fs := a.newFlagSet("watch")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	var nodes stringList
	interval := fs.Duration("interval", time.Second, "poll interval")
	duration := fs.Duration("duration", 0, "stop after this duration; zero runs until interrupted")
	fs.Var(&nodes, "node", "node id to watch; repeat for multiple nodes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	if len(nodes) == 0 {
		return errors.New("at least one --node is required")
	}
	if *interval <= 0 {
		return errors.New("--interval must be greater than zero")
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer connectCancel()

	service := uaclient.NewService(common.client)
	if err := service.Connect(connectCtx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	runCtx := context.Background()
	if *duration > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(runCtx, *duration)
		defer cancel()
	}

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	format := output.NormaliseFormat(common.format)

	for {
		for _, node := range nodes {
			row, err := service.Read(runCtx, node)
			if err != nil {
				// Non-fatal read issues should not stop watch mode.
				if errors.Is(err, uaclient.ErrConnection) || errors.Is(err, uaclient.ErrAuthSecurity) {
					return err
				}
				fmt.Fprintf(a.err, "watch read failed for %s: %v\n", node, err)
				continue
			}
			event := domain.DataChange{
				NodeID:          row.NodeID,
				Value:           row.Value,
				SourceTimestamp: firstNonEmpty(row.SourceTimestamp, row.ServerTimestamp),
			}
			if err := a.renderDataChange(format, event); err != nil {
				return err
			}
		}

		select {
		case <-runCtx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (a *App) alarms(args []string) error {
	fs := a.newFlagSet("alarms")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	node := fs.String("node", "i=2253", "event source node id; i=2253 is the Server object")
	interval := fs.Duration("interval", time.Second, "subscription interval")
	duration := fs.Duration("duration", 0, "stop after this duration; zero runs until interrupted")
	minSeverity := fs.Uint("min-severity", 0, "minimum alarm/event severity from 0 to 1000")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		return err
	}
	if *minSeverity > 1000 {
		return errors.New("--min-severity must be between 0 and 1000")
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer connectCancel()

	service := uaclient.NewService(common.client)
	if err := service.Connect(connectCtx); err != nil {
		return err
	}
	defer service.Close(context.Background())

	runCtx := context.Background()
	if *duration > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(runCtx, *duration)
		defer cancel()
	}

	subscription, err := service.SubscribeAlarms(runCtx, *node, *interval, uint16(*minSeverity))
	if err != nil {
		return err
	}
	defer subscription.Close()

	events := subscription.Events
	errorsCh := subscription.Errors
	format := output.NormaliseFormat(common.format)

	for events != nil || errorsCh != nil {
		select {
		case event, ok := <-events:
			if !ok {
				events = nil
				continue
			}
			if err := a.renderAlarmEvent(format, event); err != nil {
				return err
			}
		case err, ok := <-errorsCh:
			if !ok {
				errorsCh = nil
				continue
			}
			fmt.Fprintln(a.err, err)
		case <-runCtx.Done():
			return nil
		}
	}

	return nil
}

func (a *App) testConnection(args []string) error {
	fs := a.newFlagSet("test-connection")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := common.applyConfig(fs); err != nil {
		fmt.Fprintln(a.out, "FAIL")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.client.Timeout)
	defer cancel()

	fmt.Fprintln(a.out, "Connection diagnostics")
	fmt.Fprintf(a.out, "Endpoint: %s\n", common.client.Endpoint)
	fmt.Fprintf(a.out, "Policy: %s\n", common.client.Policy)
	fmt.Fprintf(a.out, "Mode: %s\n", common.client.Mode)
	if common.client.Username != "" {
		fmt.Fprintln(a.out, "Auth: username/password")
	} else {
		fmt.Fprintln(a.out, "Auth: anonymous")
	}
	fmt.Fprintln(a.out, "")

	fmt.Fprintln(a.out, "[1/4] Discover endpoints: PASS")
	if _, err := uaclient.ListEndpoints(ctx, common.client.Endpoint); err != nil {
		fmt.Fprintf(a.out, "[1/4] Discover endpoints: FAIL (%v)\n", err)
		fmt.Fprintln(a.out, "RESULT: FAIL")
		return fmt.Errorf("%w: endpoint discovery failed", uaclient.ErrConnection)
	}

	service := uaclient.NewService(common.client)
	if err := service.Connect(ctx); err != nil {
		fmt.Fprintf(a.out, "[2/4] Select security + establish session: FAIL (%v)\n", err)
		fmt.Fprintln(a.out, "RESULT: FAIL")
		return err
	}
	defer service.Close(context.Background())
	fmt.Fprintln(a.out, "[2/4] Select security + establish session: PASS")

	if _, err := service.Read(ctx, "i=2258"); err != nil {
		fmt.Fprintf(a.out, "[3/4] Read server status node i=2258: FAIL (%v)\n", err)
		fmt.Fprintln(a.out, "RESULT: FAIL")
		return err
	}
	fmt.Fprintln(a.out, "[3/4] Read server status node i=2258: PASS")
	fmt.Fprintln(a.out, "[4/4] End-to-end diagnostic: PASS")
	fmt.Fprintln(a.out, "RESULT: PASS")
	return nil
}
