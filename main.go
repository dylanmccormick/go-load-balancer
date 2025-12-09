package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"
)

const (
	CONN_HOST = ""
	CONN_PORT = "8080"
	CONN_TYPE = "tcp"
)

type MockServer struct {
	ConnType string
	Host     string
	Port     string
}

var SERVER_LIST = []string{"Server1", "Server2", "Server3", "Server4"}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	ln, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		slog.Error("Error starting TCP server", "error", err)
		os.Exit(1)
	}

	mockServers := make(map[string]MockServer)
	mockServers["Server1"] = MockServer{"tcp", "localhost", "8000"}
	mockServers["Server2"] = MockServer{"tcp", "localhost", "8001"}
	mockServers["Server3"] = MockServer{"tcp", "localhost", "8002"}
	mockServers["Server4"] = MockServer{"tcp", "localhost", "8003"}

	createAndServeMockServers(mockServers)

	slog.Info("Server listening on port :" + CONN_PORT)
	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Error accepting connection", "error", err)
		}

		go handleProxy(conn, mockServers)
	}
}

func handleProxy(conn net.Conn, servers map[string]MockServer) {
	backendHost, backendPort, err := selectBackend(servers)
	if err != nil {
		slog.Error("Error determining backend", "error", err)
	}
	backendConn, err := net.Dial("tcp", backendHost+":"+backendPort)
	if err != nil {
		slog.Error("Error connecting to backend server", "error", err)
	}
	bidirectionalProxy(conn, backendConn)
}

func handleConnection(conn net.Conn, port string) {
	reader := bufio.NewReader(conn)
	defer conn.Close()
	data := make([]byte, 256)
	n, err := reader.Read(data)
	if err != nil {
		slog.Error("Error reading from connection", "error", err)
	}
	pData := data[:n]
	slog.Info(string(pData), "bytes_count", n)

	msg := fmt.Sprintf("What is up client. From port %s", port)

	conn.Write([]byte(msg))
}

func selectBackend(servers map[string]MockServer) (string, string, error) {
	rInt := rand.IntN(len(SERVER_LIST))
	selectedServer := servers[SERVER_LIST[rInt]]
	return selectedServer.Host, selectedServer.Port, nil
}

// func processRequest() {
// }
func bidirectionalProxy(client, backend net.Conn) {
	defer client.Close()
	defer backend.Close()
	done := make(chan error, 2)
	// copy request to server
	go func() {
		n, err := io.Copy(client, backend)
		slog.Info("Message copied to backend", "bytes", n)
		done <- err
	}()

	// copy response to client
	go func() {
		n, err := io.Copy(backend, client)
		slog.Info("Message copied to client", "bytes", n)
		done <- err
	}()

	<-done
}

func createAndServeMockServers(servers map[string]MockServer) {
	for name, server := range servers {
		backend, err := net.Listen("tcp", server.Host+":"+server.Port)
		if err != nil {
			slog.Error("Error starting TCP server for fake backend", "error", err)
			os.Exit(1)
		}
		go func() {
			for {
				conn, err := backend.Accept()
				if err != nil {
					slog.Error("Error accepting connection", "error", err)
				}
				go handleConnection(conn, server.Port)
			}
		}()
		slog.Info("Mock Backend Server listening", "name", name, "port", server.Port)
	}
}
