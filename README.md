# opencligen

[![CI](https://github.com/crunchloop/opencligen/actions/workflows/ci.yml/badge.svg)](https://github.com/crunchloop/opencligen/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/crunchloop/opencligen)](https://goreportcard.com/report/github.com/crunchloop/opencligen)
[![Go Reference](https://pkg.go.dev/badge/github.com/crunchloop/opencligen.svg)](https://pkg.go.dev/github.com/crunchloop/opencligen)

Generate CLI tools from OpenAPI 3.0 specifications.

opencligen takes an OpenAPI spec and generates a complete Go CLI application with:
- One command per endpoint
- Commands grouped by tags
- Support for `x-cli` overrides for customizing names, flags, and configuration
- JSON and SSE (Server-Sent Events) response handling

## Installation

```bash
go install github.com/crunchloop/opencligen/cmd/opencligen@latest
```

Or build from source:

```bash
git clone https://github.com/crunchloop/opencligen
cd opencligen
go build -o opencligen ./cmd/opencligen
```

## Quick Start

Generate a CLI from your OpenAPI spec:

```bash
opencligen gen --spec api.json --out ./mycli --name mycli --build
```

This will:
1. Parse your OpenAPI spec
2. Generate a Go CLI application in `./mycli`
3. Build the binary at `./mycli/mycli`

## Usage

### Generator Command

```bash
opencligen gen [flags]

Flags:
      --spec string     Path to OpenAPI spec file (required)
      --out string      Output directory (required)
      --name string     Application name (required)
      --module string   Go module name (optional, defaults to app name)
      --build           Build the generated CLI after generation
      --dry-run         Print plan without generating files
```

### Example

```bash
# Generate and preview the plan
opencligen gen --spec api.json --out /tmp/mycli --name mycli --dry-run

# Generate and build
opencligen gen --spec api.json --out ./mycli --name mycli --build
```

## Generated CLI

The generated CLI follows these conventions:

### Command Structure

Commands are organized by OpenAPI tags:

```
mycli
├── tasks           # From tag: "tasks"
│   ├── list        # GET /v1/tasks
│   ├── create      # POST /v1/tasks
│   ├── get <id>    # GET /v1/tasks/{id}
│   └── cancel <id> # POST /v1/tasks/{id}/cancel
├── workspaces      # From tag: "workspaces"
│   ├── list
│   └── get <id>
└── stream          # From tag: "stream"
    └── subscribe   # SSE endpoint
```

### Base URL Configuration

The generated CLI requires a base URL. Configure it via:

1. **Command-line flag**: `--base-url https://api.example.com`
2. **Environment variable**: `MYAPP_BASE_URL=https://api.example.com`
3. **Config file**: `~/.config/myapp/config.yaml`

```yaml
# ~/.config/myapp/config.yaml
base_url: https://api.example.com
headers:
  Authorization: Bearer token123
```

### Request Body Input

For endpoints with request bodies, use the `--data` flag:

```bash
# Inline JSON
mycli tasks create --data '{"name": "Task 1"}'

# From file
mycli tasks create --data @task.json

# From stdin
echo '{"name": "Task 1"}' | mycli tasks create --data @-
```

### Global Flags

All generated CLIs include these global flags:

- `--base-url`: API base URL
- `--timeout`: Request timeout (default: 30s)
- `--header`: Extra headers (repeatable)

## x-cli Annotations

Customize the generated CLI using `x-cli` vendor extensions in your OpenAPI spec.

### Operation-level Overrides

```yaml
paths:
  /tasks/{taskId}/activities:
    get:
      operationId: listTaskActivities
      x-cli:
        name: "tasks activities"  # Custom command path
        aliases: ["act", "a"]     # Command aliases
        hidden: true              # Hide from help
        group: "admin"            # Override tag grouping
```

### Parameter-level Overrides

```yaml
parameters:
  - name: X-Org-Id
    in: header
    x-cli:
      flag: "org"           # Custom flag name
      shorthand: "o"        # Single-letter shorthand
      env: "ORG_ID"         # Environment variable
      config: "org_id"      # Config file key
      positional: false     # Force as flag (for path params)
```

### Supported x-cli Options

**Operation level:**
| Option | Type | Description |
|--------|------|-------------|
| `name` | string | Space-delimited command path (e.g., "tasks activities") |
| `aliases` | []string | Command aliases |
| `hidden` | bool | Hide command from help output |
| `group` | string | Override tag grouping |

**Parameter level:**
| Option | Type | Description |
|--------|------|-------------|
| `flag` | string | Override flag name |
| `shorthand` | string | Single-letter shorthand |
| `env` | string | Environment variable to read from |
| `config` | string | Config file key to read from |
| `positional` | bool | Whether path param is positional (default: true) |

## Command Naming

By default, command names are derived from `operationId`:

| operationId | Command |
|-------------|---------|
| `listTasks` | `list` |
| `getTask` | `get` |
| `createTask` | `create` |
| `updateTask` | `update` |
| `deleteTask` | `delete` |
| `startProcess` | `start` |
| `cancelTask` | `cancel` |
| `subscribeStream` | `subscribe` |
| `customOperation` | `custom-operation` |

Use `x-cli.name` to override this behavior.

## Flag Naming

Flags are derived from parameter names:

- Query/path params: converted to kebab-case
- Header params: `X-` prefix stripped, converted to kebab-case
  - `X-User-Id` → `--user-id`
  - `X-Request-ID` → `--request-id`

Use `x-cli.flag` to override.

## SSE (Server-Sent Events) Support

Endpoints returning `text/event-stream` are automatically handled:

```bash
mycli stream subscribe
# Outputs each SSE data chunk as pretty-printed JSON
```

## Development

### Running Tests

```bash
make test
# or
go test ./...
```

### Running Linting

```bash
make lint
```

### Building

```bash
make build
```

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Project Structure

```
opencligen/
├── cmd/opencligen/        # Generator CLI
├── internal/
│   ├── spec/              # OpenAPI spec loading
│   ├── plan/              # Command plan builder
│   ├── gen/               # Code generation
│   │   ├── templates/     # Go templates
│   │   └── runtime/       # Embedded runtime
│   ├── runtime_skel/      # Runtime implementation
│   └── testdata/          # Test fixtures
└── README.md
```

## License

MIT
