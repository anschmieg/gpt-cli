package main

import (
	"fmt"
	"net"
	"time"
)

// TestServer wraps a mock server for testing
type TestServer struct {
	mockServer *MockServer
	URL        string
}

func (ts *TestServer) Close() {
	if ts.mockServer != nil {
		ts.mockServer.Stop()
	}
}

// startMockTestServer starts a mock server for testing
func startMockTestServer() *TestServer {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	
	// Create mock server
	mockServer := NewMockServer(fmt.Sprintf("%d", port))
	
	// Start server in background
	go func() {
		mockServer.Start()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	return &TestServer{
		mockServer: mockServer,
		URL:        url,
	}
}