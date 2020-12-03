package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/config"
	httputil "github.com/saiya/dsps/server/http/util"
)

// StartServer starts HTTP web server
func StartServer(config *config.ServerConfig, deps *ServerDependencies) {
	engine := createServer(config, deps)
	runServer(config, engine, deps.GetServerClose())
}

func createServer(config *config.ServerConfig, deps *ServerDependencies) *gin.Engine {
	// TODO: Disable Gin debug mode if no GIN_MODE given
	engine := gin.New() // TODO: Use gin.New()
	// FIXME: Set logging & recovery (error capture) middlewares

	var router gin.IRoutes
	if config.HTTPServer.PathPrefix == "/" {
		router = engine
	} else {
		router = engine.Group(config.HTTPServer.PathPrefix)
	}

	InitEndpoints(router, deps)

	return engine
}

// see: https://github.com/gin-gonic/gin#manually
func runServer(config *config.ServerConfig, engine *gin.Engine, serverClose httputil.ServerClose) {
	srv := &http.Server{
		// FIXME: Make listen address configurable and document it
		// https://forum.eset.com/topic/22080-mac-firewall-issue-after-update-to-684000/
		Addr:           fmt.Sprintf("127.0.0.1:%d", config.HTTPServer.Port),
		Handler:        engine,
		ReadTimeout:    60 * time.Second, // FIXME: Fix hardcorded
		WriteTimeout:   60 * time.Second, // FIXME: Fix hardcorded
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("listen failed: %s\n", err) // FIXME: Use logger
			} else {
				log.Println("Server listener closed")
			}
		}
	}()
	log.Println(fmt.Sprintf("Server running on port %d", config.HTTPServer.Port)) // FIXME: Use logger

	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	serverClose.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.HTTPServer.GracefulShutdownTimeout.Duration)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Aborted long-running requests (e.g. awaiting long pollings)") // FIXME: Use logger
		} else {
			log.Fatal("Server forced to shutdown ", err) // FIXME: Use logger
		}
	}

	log.Println("Server exiting") // FIXME: Use logger
}
