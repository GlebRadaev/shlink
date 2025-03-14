// app_test.go
package app_test

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/GlebRadaev/shlink/internal/app"
	"github.com/GlebRadaev/shlink/pkg/shlink"
)

func resetFlagsAndArgs() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}
}

func resetEnv() {
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("GRPC_SERVER_ADDRESS")
}

func waitForServer(addr string, timeout time.Duration) error {
	start := time.Now()
	for {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("server did not start within %v", timeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
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
	assert.NotNil(t, application.GRPCServer)
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

	err = waitForServer(application.Config.ServerAddress, 5*time.Second)
	assert.NoError(t, err)

	resp, err := http.Get("http://" + application.Config.ServerAddress + "/ping")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	conn, err := grpc.NewClient(application.Config.GRPCServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := shlink.NewShlinkServiceClient(conn)
	grpcResp, err := client.Ping(context.Background(), &shlink.PingRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, grpcResp)

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
	err = waitForServer(application.Config.ServerAddress, 5*time.Second)
	assert.NoError(t, err)

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
	err = waitForServer(application.Config.ServerAddress, 5*time.Second)
	assert.NoError(t, err)

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
