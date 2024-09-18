package zephyrix

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewApplication(t *testing.T) {
	app := NewApplication()
	require.NotNil(t, app, "NewApplication should return a non-nil Zephyrix instance")
}

func TestZephyrixStartStop(t *testing.T) {
	app := NewApplication()

	// Set up a test route
	app.Router().GET("/test", func(c Context) {
		c.JSON(200, "Test route")
	})

	// Start the application in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Start(ctx)
	}()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Test if the server is running
	resp, err := http.Get("http://localhost:8000/test")
	require.NoError(t, err, "Server should be accessible")
	require.Equal(t, 200, resp.StatusCode, "Server should respond with 200 OK")

	// Stop the application
	err = app.Stop()
	require.NoError(t, err, "Stop should not return an error")

	// Check if the application has stopped
	select {
	case err := <-errChan:
		require.NoError(t, err, "Start should not return an error when stopped")
	case <-time.After(2 * time.Second):
		t.Fatal("Application did not stop within the expected timeframe")
	}

	// Verify that the server is no longer accessible
	_, err = http.Get("http://localhost:8000/test")
	require.Error(t, err, "Server should not be accessible after stopping")
}

func TestZephyrixStartStopWithContext(t *testing.T) {
	app := NewApplication()
	app.Router().GET("/test", func(c Context) {
		c.JSON(200, "Test route")
	})

	// Start the application with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Start(ctx)
	}()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Test if the server is running
	resp, err := http.Get("http://localhost:8000/test")
	require.NoError(t, err, "Server should be accessible")
	require.Equal(t, 200, resp.StatusCode, "Server should respond with 200 OK")

	// Cancel the context to stop the application
	cancel()

	// Check if the application has stopped
	select {
	case err := <-errChan:
		require.NoError(t, err, "Start should not return an error when stopped via context cancellation")
	case <-time.After(2 * time.Second):
		t.Fatal("Application did not stop within the expected timeframe")
	}

	// Verify that the server is no longer accessible
	_, err = http.Get("http://localhost:8000/test")
	require.Error(t, err, "Server should not be accessible after stopping")
}
