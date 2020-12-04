package main

import (
	"context"
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
		port   = flag.Int("port", 0, "Override http.port configuration item")
		listen = flag.String("listen", "", "Override http.listen configuration item")

		debug = flag.Bool("debug", false, "Enable debug logs")
	)
	flag.Parse()
	configFile := flag.Arg(0)
	configOverrides := config.Overrides{
		BuildVersion: buildVersion,
		BuildAt:      buildAt,
		Port:         *port,
		Listen:       *listen,
		Debug:        *debug,
	}

	ctx := context.Background()
	clock := domain.RealSystemClock
	config := loadConfig(configFile, configOverrides)
	channelProvider := newChannelProvider(&config)
	if err := logger.InitLogger(&config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	storage, err := storage.NewStorage(ctx, &config.Storages, clock, channelProvider)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	http.StartServer(ctx, &config, &http.ServerDependencies{
		ServerClose:           httputil.NewServerClose(),
		Storage:               storage,
		LongPollingMaxTimeout: config.HTTPServer.LongPollingMaxTimeout,
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
