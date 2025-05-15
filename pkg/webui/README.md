# Bridgette Web UI

This package provides a web UI for the Bridgette application. It uses:

- [templ](https://templ.guide) for HTML templates
- [htmx](https://htmx.org) for dynamic UI updates
- [Tailwind CSS](https://tailwindcss.com) for styling (via CDN)

## Features

- Dashboard with bridge statistics
- Timeline of deposits showing:
  - Time taken for deposits to appear on L2
  - Transaction details for both L1 and L2
  - Amount, sender, and receiver information

## Development

To generate the templ code after modifying templates, run:

```bash
./generate.sh
```

## Usage

The web UI is automatically started with the main application. You can access it at `http://localhost:8085` by default (configurable with the `--web-ui-addr` flag).

```bash
# Example start command
./bridgette --l1-execution-url="..." --l2-execution-url="..." --web-ui-addr=":8085"
``` 