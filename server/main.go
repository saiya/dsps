package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/saiya/dsps/server/channel"
	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http"
	httplifecycle "github.com/saiya/dsps/server/http/lifecycle"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage"
	"github.com/saiya/dsps/server/unix"
)

var buildVersion string

// UNIX epoch (e.g. 1605633588)
var buildAt string

func main() {
	var (
		port   = flag.Int("port", 0, "Override http.port configuration item")
		listen = flag.String("listen", "", "Override http.listen configuration item")

		debug      = flag.Bool("debug", false, "Enable debug logs")
		dumpConfig = flag.Bool("dump-config", false, "Dump loaded configuration to stdout (for debug only)")
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
	config, err := config.LoadConfigFile(configFile, configOverrides)
	exitIfError(2, err)
	if *dumpConfig {
		exitIfError(2, config.DumpConfig(os.Stderr))
	}

	channelProvider, err := channel.NewChannelProvider(ctx, &config, clock)
	exitIfError(2, err)
	logFilter, err := logger.InitLogger(config.Logging)
	exitIfError(2, err)
	storage, err := storage.NewStorage(ctx, &config.Storages, clock, channelProvider)
	exitIfError(2, err)

	unix.NotifyUlimit(ctx, unix.UlimitRequirement{
		NoFiles: storage.GetNoFilePressure(),
	})

	http.StartServer(ctx, &http.ServerDependencies{
		Config:          &config,
		ChannelProvider: channelProvider,
		Storage:         storage,

		LogFilter:   logFilter,
		ServerClose: httplifecycle.NewServerClose(),
	})
}

func exitIfError(code int, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}
