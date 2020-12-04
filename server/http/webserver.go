package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/config"
	httputil "github.com/saiya/dsps/server/http/util"
	"github.com/saiya/dsps/server/logger"
)

// StartServer starts HTTP web server
func StartServer(mainContext context.Context, config *config.ServerConfig, deps *ServerDependencies) {
	engine := createServer(mainContext, config, deps)
	runServer(mainContext, config, engine, deps.GetServerClose())
}

func createServer(mainContext context.Context, config *config.ServerConfig, deps *ServerDependencies) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" { // Use release mode by default
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(logger.LoggingMiddleware())

	var router gin.IRouter
	if config.HTTPServer.PathPrefix == "/" {
		router = engine
	} else {
		router = engine.Group(config.HTTPServer.PathPrefix)
	}

	InitEndpoints(mainContext, router, deps)

	return engine
}

// see: https://github.com/gin-gonic/gin#manually
func runServer(mainContext context.Context, config *config.ServerConfig, engine *gin.Engine, serverClose httputil.ServerClose) {
	addr := config.HTTPServer.Listen
	srv := &http.Server{
		Addr:           addr,
		Handler:        engine,
		ReadTimeout:    60 * time.Second, // FIXME: Fix hardcorded
		WriteTimeout:   60 * time.Second, // FIXME: Fix hardcorded
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				logger.Of(mainContext).Fatal(fmt.Sprintf("HTTP server listen failed on %s", addr), err)
			} else {
				logger.Of(mainContext).Infof("HTTP server listener closed")
			}
		}
	}()
	logger.Of(mainContext).Infof("HTTP server running on %s", addr)

	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	signal := <-quit // Wait until signal...
	logger.Of(mainContext).Infof("Shutting down server (%v)...", signal)
	serverClose.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.HTTPServer.GracefulShutdownTimeout.Duration)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Of(mainContext).Infof("Stopping long-running requests (e.g. long pollings)")
		} else {
			logger.Of(mainContext).Warnf("Server forced to shutdown: %v", err)
		}
	}
	logger.Of(mainContext).Infof("Server exiting...")
}
