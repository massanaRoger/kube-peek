# kube-peek

A learning project implementing a simplified kubectl alternative for listing Kubernetes Pods. Built to understand Kubernetes client-go library and CLI development patterns.

## About

This project was created as a learning exercise to explore:
- Kubernetes client-go library usage
- Building CLI tools with Cobra
- Real-time data presentation with watch functionality
- Go project structure and interfaces

## Features

- **Pod Listing**: Query pods with namespace and selector filtering
- **Multiple Output Formats**: Table view and JSON output
- **Watch Mode**: Real-time updates with live table refreshing
- **CLI Interface**: Clean command structure built with Cobra

## Installation

### Build from Source

```bash
git clone https://github.com/massanaRoger/kube-peek.git
cd kube-peek
go build -o kubepeek .
```

## Prerequisites

- Go 1.21+
- Access to a Kubernetes cluster
- Valid kubeconfig file (typically at `~/.kube/config`)

## Usage

```bash
# List pods in default namespace
kubepeek get pods

# List pods in specific namespace
kubepeek get pods -n kube-system

# List pods across all namespaces
kubepeek get pods -A

# Output as JSON
kubepeek get pods -o json

# Watch pods with live updates
kubepeek get pods -w

# Filter by labels
kubepeek get pods -l app=nginx

# Filter by fields
kubepeek get pods --field-selector status.phase=Running
```

## Architecture

The project demonstrates clean separation of concerns:

```
cmd/
├── root.go           # CLI commands and flags
internal/kube/
├── client.go         # Kubernetes client setup
├── controller.go     # Main control logic
├── pods_interface.go # Pod source interface
├── print.go          # Output formatters
├── print_live.go     # Live table updates
└── source_clientgo.go # client-go implementation
```

## Learning Highlights

- **Interface Design**: Abstracted `PodSource` and `Printer` interfaces
- **Client-go Integration**: Proper Kubernetes API client usage
- **Watch API**: Implementing real-time updates with Kubernetes watch streams
- **CLI Patterns**: Cobra command structure and flag handling
- **Terminal Control**: ANSI escape sequences for live table updates

## Development

```bash
# Build
go build -o kubepeek .

# Format code
go fmt ./...

# Check for issues
go vet ./...
```

## License

MIT License