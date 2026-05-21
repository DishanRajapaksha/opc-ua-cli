# TUI Node Browser Implementation Plan

## Summary

Add an interactive `tui` command while preserving the existing script-friendly commands.

The TUI provides:

- Address-space tree with lazy child loading.
- Selected node attributes and current value reads.
- Monitored values panel.
- Info/event log.
- Footer key hints.

## Command Shape

```bash
opc-ua-cli tui --node i=84 --interval 1s
```

The command supports the same connection, config, auth, timeout, verbose, and debug globals as existing runtime commands. It does not support `--format` or `--duration`.

## Behavior

- Arrows/Enter: navigate and expand tree.
- Tab: move focus across panes.
- `a`: refresh details for selected node.
- `r`: read selected node once.
- `m`: monitor selected node.
- `u`: unmonitor selected node.
- `R`: reload selected branch.
- `c`: clear info log.
- `?`: show help.
- `q` or Ctrl-C: exit.

## Protocol Notes

- Uses existing `uaclient.Service`.
- Browse uses existing `Browse` behavior.
- Details use existing `Attributes` behavior.
- Reads use existing `Read` behavior.
- Live values use existing subscription-based `Monitor`.

## Validation

- CLI tests cover help, invalid interval handling, global flag handling, and completions.
- Controller tests cover monitor restart ordering, read-error logging, and bounded logs.
- Verified with `go test ./...` and `make build`.
