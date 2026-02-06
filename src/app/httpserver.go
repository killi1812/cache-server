package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/killi1812/go-cache-server/config"
	"go.uber.org/zap"
)

var signalNotificationCh = make(chan os.Signal, 1)

// TODO: reimplement start to be a Background task mby

// Start will start the web server of the app
func Start() {
	// relay selected signals to channel
	// - os.Interrupt, ctrl-c
	// - syscall.SIGTERM, program termination
	signal.Notify(signalNotificationCh, os.Interrupt, syscall.SIGTERM)

	// create scheduler
	schedulerWg := sync.WaitGroup{}
	schedulerCtx := context.Background()
	schedulerCtx, schedulerCancel := context.WithCancel(schedulerCtx)
	zap.S().Debugln("Created scheduler context")

	schedulerWg.Add(1)
	go checkInterrupt(schedulerCtx, &schedulerWg, schedulerCancel)
	zap.S().Debugln("Started CheckInterrupt")

	schedulerWg.Add(1)
	go run(schedulerCtx, &schedulerWg)
	zap.S().Infoln("Started HTTP server")

	schedulerWg.Wait()

	zap.S().Debugln("Terminated program")
}

func checkInterrupt(ctx context.Context, wg *sync.WaitGroup, schedulerCancel context.CancelFunc) {
	defer wg.Done()

	for {
		select {

		case <-ctx.Done():
			zap.S().Debugln("Terminated CheckInterrupt")
			return

		case sig := <-signalNotificationCh:
			zap.S().Debugf("Received signal on notification channel, signal = %v", sig)
			schedulerCancel()
		}
	}
}

func run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// setup gin
	if Build == BuildProd {
		zap.S().Debugln("setting gin in ReleaseMode")
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// setup controllers
	basePath := router.Group("/api")
	for _, c := range controllers {
		zap.S().Debugln("Registering Controllers")
		c.RegisterEndpoints(basePath)
	}
	// cleanup
	controllers = nil

	addr := fmt.Sprintf(":%d", config.Config.CacheServer.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Panicf("Failes to start server err = %+v", err)
		}
	}()
	zap.S().Infof("Started HTTP listen, address = http://localhost%v", srv.Addr)

	// wait for context cancellation
	<-ctx.Done()

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeoutCancel()
	err := srv.Shutdown(timeoutCtx)
	if err != nil {
		zap.S().Errorf("Cannot shut down HTTP server, err = %v", err)
	}
	zap.S().Infoln("HTTP server was shut down")
}
