# Bridgette - Golem Bridge Monitor

Bridgette is a tool for monitoring the Optimism Bridge used by the Golem Network. It tracks deposits between L1 and L2, providing insights into bridge performance and reliability.

## Recent Updates

- **Code Generation**: Integrated templ template generation with `go generate`
- **Simplified API**: Consolidated endpoints for a more streamlined architecture
- **UI Performance**: Optimized HTMX refresh patterns for better user experience
- **Real-time Updates**: Improved auto-refresh mechanism for dashboard components

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
- `--db-url`: SQLite database URL (default: `file:./store/bridgette.db?_txlock=immediate&_auto_vacuum=2&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=true`)
- `--addr`: Address for the API to listen on (default: `:8084`)
- `--l1-bridge-address`: Address of the L1 bridge (default: `0x54d6c1435ac7b90a5d46d01ee2f22ed6ff270ed3`)
- `--web-ui-addr`: Address for the web UI (default: `:8085`)
- `--l1-block-interval`: Interval for polling L1 blocks (default: `2s`)
- `--l2-block-interval`: Interval for polling L2 blocks (default: `2s`)
- `--backfilling-batch-size`: Number of blocks to process in each backfilling batch (default: `10000`)
- `--forwarding-batch-size`: Number of blocks to process in each forwarding batch (default: `100`)

## Web UI

The web UI provides a dashboard showing bridge statistics and a timeline of deposits with their confirmation times. Access it at `http://localhost:8085` (or the configured address).

### Dashboard Features

- **Real-time Metrics**: Shows total deposits, average confirmation time, and bridged ETH
- **Bridge Performance**: Displays min/avg/max confirmation times for deposits
- **Unmatched Deposits**: Lists deposits waiting for L2 confirmation with auto-refresh
- **Deposit Timeline**: Chronological view of matched deposits with confirmation details

The UI auto-refreshes data at regular intervals to provide near real-time monitoring capabilities.

## Development

### Requirements

- Go 1.22 or higher
- templ CLI tool (automatically installed via go generate)

### Building

```bash
# Install dependencies
go mod tidy

# Generate templ templates
go generate ./pkg/webui

# Build the application
go build -o bridgette cmd/bridgette/main.go
```

### Code Generation

The project uses Go's code generation to compile templ templates into Go code:

1. HTML templates are written using the [templ](https://templ.guide) syntax in `.templ` files
2. The `go generate ./pkg/webui` command processes these files to create type-safe Go code
3. The generated code integrates with the application's HTTP handlers

The go generate directive automatically installs the templ CLI tool if needed. 