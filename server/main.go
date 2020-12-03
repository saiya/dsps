package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http"
	httputil "github.com/saiya/dsps/server/http/util"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage"
)

var buildVersion string

// UNIX epoch (e.g. 1605633588)
var buildAt string

func main() {
	var (
		port = flag.Int("port", 0, "Override http.port configuration item")
	)
	flag.Parse()
	configFile := flag.Arg(0)
	configOverrides := config.Overrides{
		BuildVersion: buildVersion,
		BuildAt:      buildAt,
		Port:         *port,
	}

	clock := domain.RealSystemClock
	config := loadConfig(configFile, configOverrides)
	channelProvider := newChannelProvider(&config)
	logger, dle, err := logger.NewLogger(&config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	storage, err := storage.NewStorage(&config.Storages, clock, channelProvider)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	http.StartServer(&config, &http.ServerDependencies{
		Logger:          logger,
		DebugLogEnabler: dle,

		ServerClose: httputil.NewServerClose(),

		Storage: storage,
	})
}

func loadConfig(configFile string, configOverrides config.Overrides) config.ServerConfig {
	configYaml := ""
	if configFile != "" {
		configYaml = loadConfigFile(configFile)
	}

	config, err := config.ParseConfig(configOverrides, configYaml)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	return config
}

func loadConfigFile(configFile string) string {
	yamlBytes, err := ioutil.ReadFile(configFile) //nolint:gosec // Disables G304: Potential file inclusion via variable
	if err != nil {
		log.Fatal(err) // FIXME: Improve error handling
	}
	return string(yamlBytes)
}

func newChannelProvider(config *config.ServerConfig) domain.ChannelProvider {
	// FIXME: Implement
	return func(id domain.ChannelID) *domain.Channel {
		return &domain.Channel{
			Expire: domain.Duration{Duration: 3 * time.Hour},
		}
	}
}
