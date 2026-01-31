package pluginapi

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/grpc"
)

// ServePlugin is a helper function that dramatically simplifies plugin main() functions.
// It reads the plugin config from embedded YAML, initializes the BasePlugin, and starts serving.
// This now serves direct gRPC on ORI_PLUGIN_GRPC_PORT (no go-plugin handshake).
//
// This function eliminates 30+ lines of boilerplate from every plugin's main() function,
// reducing it to a single line:
//
// Usage:
//
//	//go:embed plugin.yaml
//	var configYAML string
//
//	func main() {
//	    pluginapi.ServeGRPCPlugin(&myTool{}, configYAML)
//	}
//
// The function automatically:
// - Parses plugin.yaml configuration
// - Creates and initializes BasePlugin with all metadata
// - Injects BasePlugin into your tool struct (via reflection)
// - Starts the gRPC plugin server on ORI_PLUGIN_GRPC_PORT
//
// Requirements:
// - tool must be a pointer to a struct
// - tool must embed pluginapi.BasePlugin
// - configYAML must be a valid plugin.yaml string
func ServePlugin(tool PluginTool, configYAML string) {
	ServeGRPCPlugin(tool, configYAML)
}

// ServeGRPCPlugin starts a direct gRPC server (no go-plugin handshake).
// It listens on the port provided via ORI_PLUGIN_GRPC_PORT.
func ServeGRPCPlugin(tool PluginTool, configYAML string) {
	// Parse plugin config from embedded YAML
	config, err := readPluginConfig(configYAML)
	if err != nil {
		panic(fmt.Sprintf("ServeGRPCPlugin failed to parse config: %v", err))
	}

	// Get API version from config, default to "v1"
	apiVersion := config.Requirements.ApiVersion
	if apiVersion == "" {
		apiVersion = "v1"
	}

	// Create base plugin with all metadata from config
	base := newBasePlugin(
		config.Name,
		config.Version,
		config.Requirements.MinOriVersion,
		config.Requirements.MaxOriVersion, // May be empty (no max limit)
		apiVersion,
	)

	// Set plugin config for YAML-based features
	base.SetPluginConfig(&config)

	// Set metadata from config
	if metadata, err := config.ToMetadata(); err == nil {
		base.SetMetadata(metadata)
	}

	// Use reflection to inject BasePlugin into the tool struct
	if err := injectBasePlugin(tool, &base); err != nil {
		panic(fmt.Sprintf("ServeGRPCPlugin failed: %v", err))
	}

	portStr := strings.TrimSpace(os.Getenv("ORI_PLUGIN_GRPC_PORT"))
	if portStr == "" {
		panic("ServeGRPCPlugin requires ORI_PLUGIN_GRPC_PORT to be set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		panic(fmt.Sprintf("ServeGRPCPlugin invalid ORI_PLUGIN_GRPC_PORT: %q", portStr))
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("ServeGRPCPlugin failed to listen on %s: %v", addr, err))
	}

	server := grpc.NewServer()
	RegisterToolServiceServer(server, &grpcServer{Impl: tool})

	if err := server.Serve(lis); err != nil {
		panic(fmt.Sprintf("ServeGRPCPlugin gRPC server error: %v", err))
	}
}

// injectBasePlugin uses reflection to find and set the embedded BasePlugin field
func injectBasePlugin(tool PluginTool, base *BasePlugin) error {
	toolValue := reflect.ValueOf(tool)

	// Ensure tool is a pointer
	if toolValue.Kind() != reflect.Ptr {
		return fmt.Errorf("tool must be a pointer, got %T", tool)
	}

	// Get the struct value
	structValue := toolValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("tool must be a pointer to struct, got %T", tool)
	}

	// Find the embedded BasePlugin field
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Check if this is an embedded BasePlugin field
		if field.Type == reflect.TypeOf(BasePlugin{}) && field.Anonymous {
			fieldValue := structValue.Field(i)

			if !fieldValue.CanSet() {
				return fmt.Errorf("cannot set BasePlugin field in %T (field is unexported)", tool)
			}

			// Set the BasePlugin field by copying pointer's element
			baseValue := reflect.ValueOf(base).Elem()
			fieldValue.Set(baseValue)
			return nil
		}
	}

	return fmt.Errorf("tool %T does not embed pluginapi.BasePlugin", tool)
}
