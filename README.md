# Bridgette - Golem Bridge Monitor

Bridgette is a tool for monitoring the Optimism Bridge used by the Golem Network. It tracks deposits between L1 (Ethereum) and L2 (Golem Network), providing insights into bridge performance and reliability.

## Features

- Monitors ETH deposits from L1 to L2
- Matches L1 deposits with their corresponding L2 confirmations
- Calculates time differences between deposits and confirmations
- Provides a clean, modern web UI to visualize the bridge activity
- Timeline view showing deposit history and confirmation times

## Technologies

- Go for backend processing
- SQLite for data storage
- [templ](https://templ.guide) for HTML templates
- [htmx](https://htmx.org) for dynamic UI updates
- [Tailwind CSS](https://tailwindcss.com) for styling

## Usage

```bash
# Start the bridge monitor
./bridgette --l1-execution-url="<L1-NODE-URL>" --l2-execution-url="<L2-NODE-URL>"
```

### Command-line Options

- `--l1-execution-url`: URL of the L1 execution layer (required)
- `--l2-execution-url`: URL of the L2 execution layer (required)
- `--db-url`: SQLite database URL (default: `./store/bridgette.db`)
- `--addr`: Address for the API to listen on (default: `:8084`)
- `--l1-bridge-address`: Address of the L1 bridge (default: `0x54d6c1435ac7b90a5d46d01ee2f22ed6ff270ed3`)
- `--web-ui-addr`: Address for the web UI (default: `:8085`)

## Web UI

The web UI provides a dashboard showing bridge statistics and a timeline of deposits with their confirmation times. Access it at `http://localhost:8085` (or the configured address).

## Development

### Requirements

- Go 1.22 or higher
- templ CLI tool (for template generation)

### Building

```bash
# Install dependencies
go mod tidy

# Generate templ templates
go generate ./pkg/webui

# Build the application
go build -o bridgette cmd/bridgette/main.go
``` 