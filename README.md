# ori-pluginapi

Go SDK for building [Ori Agent](https://github.com/oriagent/ori-agent) plugins.

## Installation

```bash
go get github.com/oriagent/ori-pluginapi
```

## Quick Start

Create a minimal plugin:

**main.go:**
```go
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/oriagent/ori-pluginapi"
)

//go:embed plugin.yaml
var configYAML string

type MyPlugin struct {
	pluginapi.BasePlugin
}

type Params struct {
	Input string `json:"input"`
}

func (t *MyPlugin) Call(ctx context.Context, args string) (string, error) {
	var params Params
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	return fmt.Sprintf("Hello, %s!", params.Input), nil
}

func main() {
	pluginapi.ServePlugin(&MyPlugin{}, configYAML)
}
```

**plugin.yaml:**
```yaml
name: my-plugin
version: 1.0.0
description: My first plugin

requirements:
  min_ori_version: "0.0.9"
  api_version: "v1"

tool_definition:
  description: "Greets the user"
  parameters:
    - name: input
      type: string
      description: "Name to greet"
      required: true
```

**go.mod:**
```
module my-plugin

go 1.21

require github.com/oriagent/ori-pluginapi v1.0.0
```

Build:
```bash
go mod tidy
go build -o my-plugin .
```

## Features

- **BasePlugin**: Default implementations for common interfaces
- **ServePlugin**: One-line plugin server bootstrap
- **Settings API**: Persistent key-value storage per agent
- **Structured Results**: Tables, lists, cards for rich UI rendering
- **Web Pages**: Serve custom HTML dashboards
- **YAML Config**: Define tool parameters in plugin.yaml

## Optional Interfaces

Plugins can implement these for additional features:

| Interface | Purpose |
|-----------|---------|
| `VersionedTool` | Version information |
| `AgentAwareTool` | Access agent context |
| `WebPageProvider` | Serve web pages |
| `SettingsProvider` | Default configuration |
| `InitializationProvider` | Required config variables |
| `MetadataProvider` | Maintainer/license info |
| `HealthCheckProvider` | Custom health checks |
| `FileAttachmentHandler` | Accept file uploads |

## License

Apache 2.0 - See [LICENSE](LICENSE)
# ori-pluginapi
