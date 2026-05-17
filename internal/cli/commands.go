package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/DishanRajapaksha/opc-ua-cli/internal/config"
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
	node := fs.String("node", "", "node id to read")
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

	row, err := service.Read(ctx, *node)
	if err != nil {
		return err
	}
	return a.renderRead(common.format, row)
}

func (a *App) write(args []string) error {
	fs := a.newFlagSet("write")
	common := commonOptions{}
	addCommonFlags(fs, &common)
	node := fs.String("node", "", "node id to write")
	value := fs.String("value", "", "value to write")
	valueType := fs.String("type", "string", "scalar value type")
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

	row, err := service.Write(ctx, *node, *valueType, *value)
	if err != nil {
		return err
	}
	return a.renderWrite(common.format, row)
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
