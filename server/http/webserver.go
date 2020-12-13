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
	httplifecycle "github.com/saiya/dsps/server/http/lifecycle"
	"github.com/saiya/dsps/server/http/middleware"
	"github.com/saiya/dsps/server/logger"
)

// StartServer starts HTTP web server
func StartServer(mainContext context.Context, deps *ServerDependencies) {
	engine := CreateServer(mainContext, deps)
	runServer(mainContext, deps.Config, engine, deps.GetServerClose())
}

// CreateServer creates server (http.Handler) instance.
func CreateServer(mainContext context.Context, deps *ServerDependencies) http.Handler {
	if os.Getenv("GIN_MODE") == "" { // Use release mode by default
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(middleware.LoggingMiddleware())

	var router gin.IRouter
	if deps.Config.HTTPServer.PathPrefix == "/" {
		router = engine
	} else {
		router = engine.Group(deps.Config.HTTPServer.PathPrefix)
	}

	InitEndpoints(mainContext, router, deps)

	return engine
}

// see: https://github.com/gin-gonic/gin#manually
func runServer(mainContext context.Context, config *config.ServerConfig, engine http.Handler, serverClose httplifecycle.ServerClose) {
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
				logger.Of(mainContext).FatalExitProcess(fmt.Sprintf("HTTP server listen failed on %s", addr), err)
			} else {
				logger.Of(mainContext).Infof(logger.CatServer, "HTTP server listener closed")
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
	logger.Of(mainContext).Infof(logger.CatServer, "Shutting down server (%v)...", signal)
	serverClose.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.HTTPServer.GracefulShutdownTimeout.Duration)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Of(mainContext).Infof(logger.CatServer, "Stopping long-running requests (e.g. long pollings)")
		} else {
			logger.Of(mainContext).Warnf(logger.CatServer, "Server forced to shutdown: %v", err)
		}
	}
	logger.Of(mainContext).Infof(logger.CatServer, "Server exiting...")
}
