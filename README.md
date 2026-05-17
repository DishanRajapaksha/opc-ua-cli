# opc-ua-cli

A small, script-friendly OPC UA command-line client written in Go.

This repository is the OPC UA sibling of `opc-xml-da-cli`. The goal is a practical field tool for browsing, reading, writing, and monitoring OPC UA nodes without dragging a GUI into an SSH session like a grand piano through a window.

## Features

- Discover OPC UA server endpoints and supported security modes.
- Browse nodes from any root node.
- Read scalar node values.
- Write scalar node values with explicit types.
- Monitor one or more nodes using OPC UA subscriptions.
- Output as tables, plain text, or JSON.
- Support anonymous and username/password authentication.
- Support OPC UA security policy/mode selection with client certificate and key files.

## Install

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

## Usage

```bash
opc-ua-cli endpoints --endpoint opc.tcp://localhost:4840
```

```bash
opc-ua-cli browse --endpoint opc.tcp://localhost:4840 --node i=84 --depth 1
```

```bash
opc-ua-cli read --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

```bash
opc-ua-cli read --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --format json
```

```bash
opc-ua-cli write --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --type int32 --value 42
```

```bash
opc-ua-cli monitor --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s
```

Monitor for a fixed time:

```bash
opc-ua-cli monitor --endpoint opc.tcp://localhost:4840 --node 'ns=2;s=Demo.Static.Scalar.Int32' --interval 1s --duration 30s
```

## Security and authentication

Anonymous, no security:

```bash
opc-ua-cli read \
  --endpoint opc.tcp://localhost:4840 \
  --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

Username/password:

```bash
opc-ua-cli read \
  --endpoint opc.tcp://localhost:4840 \
  --username user \
  --password secret \
  --node 'ns=2;s=Demo.Static.Scalar.Int32'
```

Signed and encrypted endpoint:

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

## Project layout

```text
.
├── main.go
├── internal
│   ├── cli        # command parsing and command handlers
│   ├── config     # connection and security configuration
│   ├── domain     # plain application models
│   ├── output     # table, text, and JSON rendering
│   └── uaclient   # OPC UA session lifecycle and protocol operations
├── .github/workflows/ci.yml
├── Makefile
└── go.mod
```

## Development

```bash
make fmt
make test
make build
```

CI runs formatting, tests, and a build on pushes and pull requests.

## Design notes

This is intentionally dependency-light. The CLI uses the Go standard `flag` package instead of Cobra because the command surface is small. If the tool grows into profiles, config files, completions, or nested command groups, Cobra becomes worth its rent. Until then, fewer moving parts wins.

The OPC UA implementation is isolated under `internal/uaclient`, so command parsing, output formatting, and protocol handling do not melt into one regrettable soup.
