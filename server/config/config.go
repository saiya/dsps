package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	goyaml "github.com/goccy/go-yaml"

	"github.com/saiya/dsps/server/domain"
)

// Overrides is to override configuration file.
type Overrides struct {
	BuildVersion string // Supplied by compiler
	BuildAt      string // UNIX epoch, supplied by compiler

	Port   int
	Listen string

	Debug bool
}

// ServerConfig represents parsed/post-processed configuration.
type ServerConfig struct {
	BuildInfo  *BuildInfo        `json:"__buildInfo"`
	Storages   StoragesConfig    `json:"storages"`
	HTTPServer *HTTPServerConfig `json:"http"`
	Logging    *LoggingConfig    `json:"logging"`
	Telemetry  *TelemetryConfig  `json:"telemetry"`
	Channels   ChannelsConfig    `json:"channels"`
	Admin      *AdminConfig      `json:"admin"`
}

// BuildInfo represents compile time metadata.
type BuildInfo struct {
	BuildVersion string       `json:"buildVersion"`
	BuildAt      *domain.Time `json:"buildAt"`
}

// LoadConfigFile load file.
// If configFile == "", generate default config.
// If configFile == "-", load from stdin.
func LoadConfigFile(ctx context.Context, configFile string, configOverrides Overrides) (ServerConfig, error) {
	var yamlBytes []byte
	var err error
	switch configFile {
	case "":
		yamlBytes = []byte("")
	case "-":
		yamlBytes, err = ioutil.ReadAll(os.Stdin)
	default:
		yamlBytes, err = ioutil.ReadFile(configFile) //nolint:gosec // Disables G304: Potential file inclusion via variable
	}
	if err != nil {
		return ServerConfig{}, err
	}
	return ParseConfig(ctx, configOverrides, string(yamlBytes))
}

// ParseConfig constructs post-processed configuration object.
func ParseConfig(ctx context.Context, overrides Overrides, yaml string) (ServerConfig, error) {
	config := ServerConfig{
		BuildInfo:  parseBuildInfo(overrides),
		Storages:   DefaultStoragesConfig(),
		Logging:    loggingConfigDefault(),
		Telemetry:  tracingConfigDefault(),
		HTTPServer: httpServerConfigDefault(),
		Admin:      adminConfigDefault(),
	}

	if strings.Contains(yaml, "\t") {
		// github.com/goccy/go-yaml silently ignore TAB (0x09) so that hard to debug it for users...
		return config, fmt.Errorf("Configuration file could not contain tab character (0x09) because YAML spec forbit it, use space to indent")
	}

	validate := validator.New()
	if err := goyaml.UnmarshalWithOptions([]byte(yaml), &config, goyaml.Strict(), goyaml.Validator(validate), goyaml.UseJSONUnmarshaler()); err != nil {
		return config, fmt.Errorf("Failed to parse configuration YAML file: %w", err)
	}

	if overrides.Debug {
		config.Logging.Category = map[string]string{"*": "DEBUG"}
	}
	if err := PostprocessStorageConfig(&config.Storages); err != nil {
		return config, fmt.Errorf("Storage configration problem: %w", err)
	}
	if err := PostprocessHTTPServerConfig(config.HTTPServer, overrides); err != nil {
		return config, fmt.Errorf("HTTP server configration problem: %w", err)
	}
	if err := PostprocessTelemetryConfig(config.Telemetry); err != nil {
		return config, fmt.Errorf("Tracing configration problem: %w", err)
	}
	if err := PostprocessChannelsConfig(&config.Channels); err != nil {
		return config, fmt.Errorf("Channel configration problem: %w", err)
	}
	if err := PostprocessAdminConfig(config.Admin); err != nil {
		return config, fmt.Errorf("Admin configration problem: %w", err)
	}

	return config, nil
}

func parseBuildInfo(overrides Overrides) *BuildInfo {
	buildInfo := BuildInfo{
		BuildVersion: overrides.BuildVersion,
	}
	if buildAt, err := strconv.ParseInt(overrides.BuildAt, 10, 64); err != nil {
		var wrapped domain.Time
		wrapped.Time = time.Unix(buildAt, 0)
		buildInfo.BuildAt = &wrapped
	}
	return &buildInfo
}

// DumpConfig output config content
func (config *ServerConfig) DumpConfig(w io.Writer) error {
	json, err := json.Marshal(config)
	if err != nil {
		return err
	}
	data, err := goyaml.JSONToYAML(json)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}
