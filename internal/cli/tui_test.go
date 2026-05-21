package cli

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeTUIBackend struct {
	childrenCalls int
	readErr       error
	watchNodes    []string
}

func (f *fakeTUIBackend) Children(context.Context, string) ([]tuiNode, error) {
	f.childrenCalls++
	return []tuiNode{{ID: "child", Label: "Child"}}, nil
}

func (f *fakeTUIBackend) Details(context.Context, string) ([]tuiAttribute, error) {
	return []tuiAttribute{{Name: "NodeID", Value: "node"}}, nil
}

func (f *fakeTUIBackend) Read(context.Context, string) (tuiValue, error) {
	if f.readErr != nil {
		return tuiValue{}, f.readErr
	}
	return tuiValue{ID: "node", Label: "Node", Value: "42"}, nil
}

func (f *fakeTUIBackend) Watch(ctx context.Context, nodeIDs []string, _ time.Duration) (<-chan tuiValue, <-chan error, func(), error) {
	f.watchNodes = append([]string(nil), nodeIDs...)
	values := make(chan tuiValue, 1)
	errs := make(chan error)
	values <- tuiValue{ID: nodeIDs[0], Label: nodeIDs[0], Value: "100"}
	close(values)
	close(errs)
	return values, errs, func() {}, nil
}

func TestTUIControllerLogIsBounded(t *testing.T) {
	c := newTUIController(&fakeTUIBackend{}, time.Second)
	for i := 0; i < maxTUILogLines+5; i++ {
		c.addLog("line")
	}
	if len(c.logs) != maxTUILogLines {
		t.Fatalf("len(logs) = %d, want %d", len(c.logs), maxTUILogLines)
	}
}

func TestTUIControllerMonitorRestartUsesSortedIDs(t *testing.T) {
	backend := &fakeTUIBackend{}
	c := newTUIController(backend, time.Second)
	c.setMonitored(tuiValue{ID: "b", Label: "B"})
	c.setMonitored(tuiValue{ID: "a", Label: "A"})
	if err := c.restartWatch(context.Background(), func(tuiValue) {}, func(error) {}); err != nil {
		t.Fatalf("restartWatch returned error: %v", err)
	}
	c.stopWatch()
	if strings.Join(backend.watchNodes, ",") != "a,b" {
		t.Fatalf("watchNodes = %#v, want sorted a,b", backend.watchNodes)
	}
}

func TestTUIControllerReadErrorCanBeLogged(t *testing.T) {
	backend := &fakeTUIBackend{readErr: errors.New("boom")}
	c := newTUIController(backend, time.Second)
	if _, err := backend.Read(context.Background(), "node"); err != nil {
		c.addLog("read node: " + err.Error())
	}
	if !strings.Contains(c.logText(), "read node: boom") {
		t.Fatalf("logText = %q", c.logText())
	}
}
