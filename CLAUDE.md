# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Bridgette is a Go-based Optimism Bridge monitor that tracks ETH deposits between L1 and L2 networks for the Golem Network. It monitors deposit events, matches them between layers, and provides a real-time web dashboard.

## Development Environment

This project uses Nix flakes with a development shell. To enter the development environment:

```bash
# Enter the Nix development shell (includes Go, SQLite, sqlc)
nix develop

# Or with direnv (auto-loads when entering directory)
direnv allow
```

## Essential Commands

### Build and Development
```bash
# Generate templ templates (required before building)
go generate ./pkg/webui

# Build the application
go build -o bridgette .

# Run tests
go test ./...

# Install dependencies
go mod tidy

# Run the application
./bridgette --l1-execution-url="<L1-NODE-URL>" --l2-execution-url="<L2-NODE-URL>"
```

### Database Operations
```bash
# Generate sqlc queries after modifying SQL
cd pkg/sqlitestore && sqlc generate

# Run database migrations (handled automatically by the app)
# Migrations are in pkg/sqlitestore/migrations/
```

## Architecture

### Core Components

1. **Event Processing Pipeline** (main.go)
   - Monitors both L1 and L2 chains concurrently
   - Backfills historical data from specified block ranges
   - Forward-fills real-time events
   - Matches L1 deposits with L2 confirmations using hash correlation

2. **Database Layer** (pkg/sqlitestore/)
   - SQLite with WAL mode and optimized settings
   - Uses sqlc for type-safe query generation
   - Schema defined in migrations/ directory
   - Query definitions in queries.sql

3. **Web UI** (pkg/webui/)
   - Uses templ for type-safe HTML templating
   - htmx for dynamic updates without full page reloads
   - Auto-refreshing components for real-time monitoring
   - Routes defined in webui.go

4. **Log Parser** (pkg/logparser/)
   - Parses Ethereum event logs
   - Handles L1StandardBridgeETHDepositInitiated events
   - Handles L2 deposit finalized events
   - Generates matching hashes for event correlation

### Key Technical Decisions

- **templ Templates**: All HTML is generated using templ. Run `go generate ./pkg/webui` after modifying .templ files
- **sqlc**: Database queries are generated from SQL. Modify queries.sql then run `sqlc generate` in pkg/sqlitestore/
- **Event Matching**: L1 and L2 events are matched using a hash of (from, to, amount, extraData)
- **Concurrent Processing**: Uses goroutines for parallel L1/L2 monitoring
- **Auto-refresh**: Web UI components refresh automatically using htmx polling

### Important Contract Addresses

- L1 Standard Bridge: `0xF6080D9fbEEbcd44D89aFfBFd42F098cbFf92816`
- L2 Standard Bridge: `0x4200000000000000000000000000000000000010`

These addresses are configurable via command-line flags.

## Development Workflow

When modifying the web UI:
1. Edit .templ files in pkg/webui/
2. Run `go generate ./pkg/webui`
3. Rebuild the application

When modifying database queries:
1. Edit pkg/sqlitestore/queries.sql
2. Run `cd pkg/sqlitestore && sqlc generate`
3. Update Go code to use new queries

When adding new event types:
1. Add parser logic in pkg/logparser/
2. Add database schema migration in pkg/sqlitestore/migrations/
3. Add queries in pkg/sqlitestore/queries.sql
4. Generate sqlc code
5. Update main.go processing logic

## Production Deployment Notes

### Database Locking Issues

If you encounter "database is locked" errors in production:

1. **Single Instance Only**: SQLite requires only ONE instance to write to the database. Ensure only one pod/container is running.

2. **Local Storage Required**: SQLite must use local filesystem storage, NOT network storage (NFS, shared volumes). Network filesystems don't properly support SQLite's locking mechanisms.

3. **Increase Timeout**: The default 5-second busy timeout can be increased via the database URL:
   ```bash
   --db-url="file:/store/bridgette.db?_txlock=immediate&_busy_timeout=30000"  # 30 seconds
   ```

4. **Volume Configuration**: For Kubernetes, use a local PersistentVolume or hostPath, not shared storage:
   ```yaml
   # Use local storage, not networked storage
   volumeMounts:
   - name: store
     mountPath: /store
   volumes:
   - name: store
     hostPath:
       path: /var/lib/bridgette
       type: DirectoryOrCreate
   ```