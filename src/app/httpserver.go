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
	"go.uber.org/zap"
)

var signalNotificationCh = make(chan os.Signal, 1)

// Start starts a single web server
func Start(api CreateGinApi, addr string) {
	MultiStart(map[string]CreateGinApi{addr: api})
}

// MultiStart starts multiple web servers concurrently
func MultiStart(apis map[string]CreateGinApi) {
	// relay selected signals to channel
	signal.Notify(signalNotificationCh, os.Interrupt, syscall.SIGTERM)

	// create scheduler
	schedulerWg := sync.WaitGroup{}
	schedulerCtx := context.Background()
	schedulerCtx, schedulerCancel := context.WithCancel(schedulerCtx)
	zap.S().Debugln("Created scheduler context")

	schedulerWg.Add(1)
	go checkInterrupt(schedulerCtx, &schedulerWg, schedulerCancel)
	zap.S().Debugln("Started CheckInterrupt")

	for addr, api := range apis {
		schedulerWg.Add(1)
		go run(schedulerCtx, &schedulerWg, api, addr)
	}
	zap.S().Infoln("Started HTTP server(s)")

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

func run(ctx context.Context, wg *sync.WaitGroup, api CreateGinApi, addr string) {
	defer wg.Done()

	// setup gin
	if Build == BuildProd {
		zap.S().Debug("setting gin in ReleaseMode")
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// setup controllers
	if api != nil {
		zap.S().Debug("Registering Api")
		api.NewGinApi(router)
	} else {
		zap.S().Warn("No Api to register, api is nil")
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Panicf("Failed to start server at %s, err = %+v", addr, err)
		}
	}()
	zap.S().Debugf("Started HTTP listen, address = http://%v", srv.Addr)
	fmt.Printf("listening on http://%s\n", srv.Addr)

	// wait for context cancellation
	<-ctx.Done()

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeoutCancel()
	err := srv.Shutdown(timeoutCtx)
	if err != nil {
		zap.S().Errorf("Cannot shut down HTTP server at %s, err = %v", addr, err)
	}
	zap.S().Infof("HTTP server %s was shut down", srv.Addr)
}
