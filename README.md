# OPC UA CLI

A small, script-friendly OPC UA command-line client written in Go.

## Features

- Discover OPC UA server endpoints and supported security modes.
- Browse nodes from any root node.
- Read scalar node values.
- Write scalar node values with explicit types.
- Monitor one or more nodes using OPC UA data-change subscriptions.
- Subscribe to OPC UA alarm/event notifications with severity filtering.
- Output as tables, plain text, or JSON.
- Support anonymous and username/password authentication.
- Support OPC UA security policy/mode selection with client certificate and key files.
- Load repeated connection settings from a YAML config file.

## Install

Download a binary from the GitHub Releases page, or install with Go:

```bash
go install github.com/DishanRajapaksha/opc-ua-cli@latest
```

Or build from source:

```bash
git clone https://github.com/DishanRajapaksha/opc-ua-cli.git
cd opc-ua-cli
make test
make build
```

The binary is written to `bin/opc-ua-cli`.

## YAML config

Create a local config file:

```yaml
endpoint: opc.tcp://localhost:4840
policy: None
mode: None
timeout: 10s

# Optional authentication.
# username: user
# password: secret

# Optional client certificate settings for secured endpoints.
# cert: client-cert.pem
# key: client-key.pem
```

There is also a `config.example.yaml` in the repo.

Use it like this:

```bash
opc-ua-cli read --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

CLI flags override the config file, so this is valid when you want a one-off endpoint change:

```bash
opc-ua-cli read --config config.yaml --endpoint opc.tcp://192.168.1.50:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

## Usage

List endpoints:

```bash
opc-ua-cli endpoints --config config.yaml
```

Browse nodes:

```bash
opc-ua-cli browse --config config.yaml --node i=84 --depth 1
```

Read a node:

```bash
opc-ua-cli read --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

Read as JSON:

```bash
opc-ua-cli read --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32' --format json
```

Write a scalar value:

```bash
opc-ua-cli write --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
```

Monitor a node value:

```bash
opc-ua-cli monitor --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s
```

Monitor for a fixed time:

```bash
opc-ua-cli monitor --config config.yaml --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s --duration 30s
```

Subscribe to alarms/events from the Server object:

```bash
opc-ua-cli alarms --config config.yaml --node i=2253 --min-severity 500 --interval 1s
```

Emit alarms/events as JSON lines:

```bash
opc-ua-cli alarms --config config.yaml --node i=2253 --min-severity 0 --format json
```

Run an alarm/event subscription for a fixed time:

```bash
opc-ua-cli alarms --config config.yaml --node i=2253 --min-severity 500 --interval 1s --duration 30s
```

You can still skip the config file and pass connection flags directly:

```bash
opc-ua-cli read --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

Run field diagnostics:

```bash
opc-ua-cli test-connection
opc-ua-cli test-connection --config site-a.yaml
opc-ua-cli test-connection --profile site-a
```

## Exit codes

The CLI uses stable exit codes so automation can branch on failure category:

- `0`: success
- `1`: general error
- `2`: config error
- `3`: connection error
- `4`: authentication/security error
- `5`: node not found
- `6`: bad OPC UA status code
- `7`: write rejected

## Alarm and event subscriptions

The `alarms` command uses an OPC UA event subscription. By default it subscribes to `i=2253`, the standard Server object. Some servers expose alarm/event notifications on a different object, so pass that node with `--node` when needed.

Selected event fields include:

- `EventId`
- `EventType`
- `SourceNode`
- `SourceName`
- `Time`
- `ReceiveTime`
- `Message`
- `Severity`
- `ConditionName`
- `ActiveState`
- `AckedState`
- `Retain`

`--min-severity` accepts values from `0` to `1000`.

## Security and authentication

Anonymous, no security:

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

Signed and encrypted endpoint:

```yaml
endpoint: opc.tcp://localhost:4840
policy: Basic256Sha256
mode: SignAndEncrypt
cert: client-cert.pem
key: client-key.pem
timeout: 10s
```

Equivalent one-off command without YAML:

```bash
opc-ua-cli read \
  --endpoint opc.tcp://localhost:4840 \
  --policy Basic256Sha256 \
  --mode SignAndEncrypt \
  --cert client-cert.pem \
  --key client-key.pem \
  --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

## Write types

Supported scalar write types:

- `string`
- `bool`
- `int8`, `int16`, `int32`, `int64`
- `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`

Aliases:

- `int` maps to `int32`
- `uint` maps to `uint32`
- `float` maps to `float32`
- `double` maps to `float64`
- `byte` maps to `uint8`
- `boolean` maps to `bool`

## Development

```bash
make fmt
make test
make build
```
