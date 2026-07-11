# OPC UA CLI

A script-friendly OPC UA command-line client written in Go.

## At a Glance

| Task | Command |
|---|---|
| Validate local config | `opc-ua-cli validate-config` |
| Test connectivity/auth/security | `opc-ua-cli test-connection` |
| Read server status | `opc-ua-cli status` |
| List server endpoints | `opc-ua-cli endpoints` |
| List namespace indexes and URIs | `opc-ua-cli namespaces` |
| Browse from root node | `opc-ua-cli browse --node i=84 --depth 1` |
| Browse interactively | `opc-ua-cli tui --node i=84 --interval 1s` |
| Inspect node metadata/permissions | `opc-ua-cli attributes --node 'ns=2;s=Demo.Value'` |
| Read one node | `opc-ua-cli read --node 'ns=2;s=Demo.Static.Scalar.Int32'` |
| Read multiple nodes | `opc-ua-cli read --node 'ns=2;s=A' --node 'ns=2;s=B'` |
| Read nodes from file | `opc-ua-cli read --nodes nodes.txt` |
| Dry-run a write | `opc-ua-cli write --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42` |
| Execute a write | `opc-ua-cli write --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42 --yes` |
| Monitor (subscription) | `opc-ua-cli monitor --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s` |
| Watch (polling) | `opc-ua-cli watch --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s` |
| Troubleshoot with verbose logs | `opc-ua-cli read --node 'ns=2;s=Demo.Value' --verbose` |
| Troubleshoot with debug logs | `opc-ua-cli read --node 'ns=2;s=Demo.Value' --debug` |

## Install

Install with Go:

```bash
go install github.com/DishanRajapaksha/opc-ua-cli@latest
```

Build from source:

```bash
git clone https://github.com/DishanRajapaksha/opc-ua-cli.git
cd opc-ua-cli
make test
make build
```

Binary output: `bin/opc-ua-cli`

## First Run

1. Create `config.yaml`:

```bash
opc-ua-cli init-config
```

1. Validate config locally (no server connection):

```bash
opc-ua-cli validate-config
```

1. Verify endpoint/auth/security with a real server test:

```bash
opc-ua-cli test-connection
```

1. Inspect namespaces and read one value:

```bash
opc-ua-cli namespaces
opc-ua-cli read --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

## Config and Profiles

Use a custom config file:

```bash
opc-ua-cli read --config site-a.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

Use a profile from config:

```bash
opc-ua-cli read --config config.yaml --profile site-a --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

Override config values with CLI flags:

```bash
opc-ua-cli read --config config.yaml --endpoint opc.tcp://192.168.1.50:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

## Core Commands

### Discover and inspect

```bash
opc-ua-cli endpoints
opc-ua-cli status
opc-ua-cli namespaces
opc-ua-cli browse --node i=84 --depth 1
opc-ua-cli tui --node i=84 --interval 1s
opc-ua-cli attributes --node 'ns=2;s=Demo.Value'
```

### Interactive browser

```bash
opc-ua-cli tui --node i=84 --interval 1s
```

The TUI opens a tree browser with node attributes, current reads, monitored values, and an event log. Use arrows/Enter to expand nodes, Tab to move focus, `r` to read once, `m` to monitor, `u` to unmonitor, `R` to reload children, and `q` to exit.

### Read values

```bash
# Single node
opc-ua-cli read --node 'ns=2;s=Demo.Static.Scalar.Int32'

# Multiple nodes (output keeps request order)
opc-ua-cli read --node 'ns=2;s=A' --node 'ns=2;s=B'

# From file (one node per line, blank lines and # comments ignored)
opc-ua-cli read --nodes nodes.txt

# Namespace alias/URI form
opc-ua-cli read --node 'nsu=plant;s=Inverter01.ActivePower'
```

### Write values

```bash
# Preview only (no write request sent)
opc-ua-cli write --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
opc-ua-cli write --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42 --dry-run

# Execute write (recommended for scripts)
opc-ua-cli write --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42 --yes

# Multiple writes
opc-ua-cli write --item 'ns=2;s=A:int32:42' --item 'ns=2;s=B:bool:true' --yes
```

Supported write types:

- `string`
- `bool`
- `int8`, `int16`, `int32`, `int64`
- `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`

Write safety behavior:

- Prints endpoint, config/profile source, node, type, and value before write.
- Does not send by default.
- Requires `--yes` to send.
- Rejects `--dry-run` and `--yes` together.

### Stream changes

```bash
# Subscription-based
opc-ua-cli monitor --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s

# Polling-based alternative (useful when subscriptions are unreliable)
opc-ua-cli watch --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s

# Fixed duration + jsonl output
opc-ua-cli watch --node 'ns=2;s=Demo.Static.Scalar.Int32' --duration 1m --format jsonl
```

### Alarm/event subscriptions

```bash
opc-ua-cli alarms --node i=2253 --min-severity 500 --interval 1s
opc-ua-cli alarms --node i=2253 --min-severity 0 --format jsonl
```

Notes:

- Default alarm source node is `i=2253` (Server object).
- `--min-severity` range is `0` to `1000`.

## Troubleshooting and Diagnostics

Verbose and debug modes:

```bash
opc-ua-cli read --node 'ns=2;s=Demo.Value' --verbose
opc-ua-cli read --node 'ns=2;s=Demo.Value' --debug
```

- `--verbose`: high-level connection decisions.
- `--debug`: lower-level OPC UA client debug logging (where supported by the client library).

Sensitive values (passwords, inline cert/key material) are not printed by CLI verbose logs.

Field diagnostics command:

```bash
opc-ua-cli test-connection
opc-ua-cli test-connection --config site-a.yaml
opc-ua-cli test-connection --profile site-a
```

## Output Formats

Snapshot commands support:

- `table` (default)
- `text`
- `json`
- `csv`

Stream commands (`monitor`, `watch`, and `alarms`) support:

- `text` (default)
- `jsonl`
- `csv`

Streams reject `json` and `table`, because an unbounded stream is neither one complete JSON document nor a finite table.

Example:

```bash
opc-ua-cli read --node 'ns=2;s=Demo.Static.Scalar.Int32' --format json
```

## Security and Authentication Examples

Anonymous:

```yaml
endpoint: opc.tcp://localhost:4840
policy: None
mode: None
timeout: 10s
```

Username/password:

```yaml
endpoint: opc.tcp://localhost:4840
policy: None
mode: None
username: user
password: secret
timeout: 10s
```

Signed and encrypted:

```yaml
endpoint: opc.tcp://localhost:4840
policy: Basic256Sha256
mode: SignAndEncrypt
cert: client-cert.pem
key: client-key.pem
timeout: 10s
```

One-off command equivalent:

```bash
opc-ua-cli read \
  --endpoint opc.tcp://localhost:4840 \
  --policy Basic256Sha256 \
  --mode SignAndEncrypt \
  --cert client-cert.pem \
  --key client-key.pem \
  --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

## Shell Completions

Generate:

```bash
opc-ua-cli completions bash
opc-ua-cli completions zsh
```

Install examples:

```bash
# bash
opc-ua-cli completions bash > /etc/bash_completion.d/opc-ua-cli

# zsh
mkdir -p "${HOME}/.zsh/completions"
opc-ua-cli completions zsh > "${HOME}/.zsh/completions/_opc-ua-cli"
```

## Exit Codes

Stable exit codes for scripts:

- `0`: success
- `1`: general error
- `2`: usage or configuration error
- `3`: transport or connection error
- `4`: protocol or request error
- `5`: authentication or security error
- `6`: node not found
- `7`: write or control rejected
- `8`: operation timeout
- `9`: output or formatting error

## Command Help

```bash
opc-ua-cli help
```
