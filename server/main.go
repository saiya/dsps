package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/domain/channel"
	"github.com/saiya/dsps/server/http"
	httplifecycle "github.com/saiya/dsps/server/http/lifecycle"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/storage"
	"github.com/saiya/dsps/server/telemetry"
	"github.com/saiya/dsps/server/unix"
)

var buildVersion string

// UNIX epoch (e.g. 1605633588)
var buildAt string

func main() {
	err := mainImpl(context.Background(), os.Args[1:], domain.RealSystemClock)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func mainImpl(ctx context.Context, args []string, clock domain.SystemClock) error {
	defer func() { logger.Of(ctx).Debugf(logger.CatServer, "Sever closed.") }()

	var (
		port   = flag.Int("port", 0, "Override http.port configuration item")
		listen = flag.String("listen", "", "Override http.listen configuration item")

		debug      = flag.Bool("debug", false, "Enable debug logs")
		dumpConfig = flag.Bool("dump-config", false, "Dump loaded configuration to stdout (for debug only)")
	)
	if err := flag.CommandLine.Parse(args); err != nil {
		return err
	}
	configFile := flag.Arg(0)
	configOverrides := config.Overrides{
		BuildVersion: buildVersion,
		BuildAt:      buildAt,
		Port:         *port,
		Listen:       *listen,
		Debug:        *debug,
	}

	config, err := config.LoadConfigFile(ctx, configFile, configOverrides)
	if err != nil {
		return err
	}
	if *dumpConfig {
		if err := config.DumpConfig(os.Stderr); err != nil {
			return err
		}
	}

	logFilter, err := logger.InitLogger(config.Logging)
	if err != nil {
		return err
	}

	telemetry, err := telemetry.InitTelemetry(config.Telemetry)
	if err != nil {
		return err
	}
	defer telemetry.Shutdown(ctx)

	channelProvider, err := channel.NewChannelProvider(ctx, &config, clock, telemetry)
	if err != nil {
		return err
	}
	defer channelProvider.Shutdown(ctx)

	storage, err := storage.NewStorage(ctx, &config.Storages, clock, channelProvider, telemetry)
	if err != nil {
		return err
	}
	defer func() {
		if err := storage.Shutdown(ctx); err != nil {
			logger.Of(ctx).WarnError(logger.CatStorage, "Failed to shutdown storage: %w", err)
		}
	}()

	unix.NotifyUlimit(ctx, unix.UlimitRequirement{
		NoFiles: channelProvider.GetFileDescriptorPressure() + storage.GetFileDescriptorPressure(),
	})

	http.StartServer(ctx, &http.ServerDependencies{
		Config:          &config,
		ChannelProvider: channelProvider,
		Storage:         storage,

		Telemetry:   telemetry,
		LogFilter:   logFilter,
		ServerClose: httplifecycle.NewServerClose(),
	})
	return nil
}
