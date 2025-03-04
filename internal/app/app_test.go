// app_test.go
package app_test

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/GlebRadaev/shlink/internal/app"
)

func resetFlagsAndArgs() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}
}

func resetEnv() {
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
}

func TestNewApplication(t *testing.T) {
	ctx := context.Background()
	application := app.NewApplication(ctx)
	assert.NotNil(t, application)
	assert.Equal(t, ctx, application.Ctx)
}

func TestApplicationInit(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv()

	ctx := context.Background()
	application := app.NewApplication(ctx)

	err := application.Init()
	assert.NoError(t, err)
	assert.NotNil(t, application.Config)
	assert.NotNil(t, application.Logger)
	assert.NotNil(t, application.Services)
	assert.NotNil(t, application.Server)
}

func TestApplicationStart(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := app.NewApplication(ctx)
	err := application.Init()
	assert.NoError(t, err)

	go func() {
		if err := application.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	resp, err := http.Get("http://" + application.Config.ServerAddress + "/ping")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = application.Shutdown()
	assert.NoError(t, err)
}

func TestApplicationShutdown(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := app.NewApplication(ctx)
	err := application.Init()
	assert.NoError(t, err)

	go func() {
		if err := application.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	err = application.Shutdown()
	assert.NoError(t, err)
}

func TestApplicationSetupRoutes(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv()

	ctx := context.Background()
	application := app.NewApplication(ctx)
	_ = application.Init()

	router := application.SetupRoutes()
	assert.NotNil(t, router)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestApplicationSignalHandling(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := app.NewApplication(ctx)
	err := application.Init()
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := application.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	resp, err := http.Get("http://" + application.Config.ServerAddress + "/ping")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)

	go func() {
		time.Sleep(500 * time.Millisecond)
		signalChan <- syscall.SIGTERM
	}()

	select {
	case <-signalChan:
		t.Log("Received SIGTERM, shutting down")
		cancel()
		err := application.Shutdown()
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for shutdown")
	}

	wg.Wait()
}
