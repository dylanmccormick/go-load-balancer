package balancer

import (
	"bufio"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"log/slog"
	"net"
	"os"
	"time"
)

type Algorithm int

const (
	ROUND_ROBIN Algorithm = iota
	IP_HASHING
)

func (lb *LoadBalancer) IPHashingNextBackend(conn net.Conn) *Backend {
	slog.Info("Determining backend with IP Hashing")
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		slog.Error("Unable to hash port number. Serving backend 0", "error", err, "address", conn.RemoteAddr().String())
		return lb.backends[0]
	}
	slog.Info("Hashing requets", "ip", conn.RemoteAddr().String())
	h := fnv.New32a()
	h.Write([]byte(host))
	hash := h.Sum32()

	serverNo := int(hash) % len(lb.backends)

	return lb.backends[serverNo]
}

func (lb *LoadBalancer) NextBackend(conn net.Conn) *Backend {
	switch lb.algo {
	case ROUND_ROBIN:
		return lb.RoundRobinNextBackend()
	case IP_HASHING:
		return lb.IPHashingNextBackend(conn)
	}
	return lb.backends[0]
}

func (lb *LoadBalancer) CreateAndServeBackends() {
	for n, server := range lb.backends {
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
					// Log the error, but don't kill the server
					continue
				}
				go handleConnection(conn, server.Port)
			}
		}()
		slog.Info("Mock Backend Server listening", "number", n, "port", server.Port)
	}
}

func (lb *LoadBalancer) HandleProxy(conn net.Conn) {
	backend := lb.NextBackend(conn)
	backendConn, err := net.Dial("tcp", backend.Host+":"+backend.Port)
	if err != nil {
		slog.Error("Error connecting to backend server", "error", err)
		return
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
		conn.Close()
		return
	}
	pData := data[:n]
	slog.Info(string(pData), "bytes_count", n)

	msg := fmt.Sprintf("What is up client. From port %s", port)

	conn.Write([]byte(msg))
}

func bidirectionalProxy(client, server net.Conn) {
	defer client.Close()
	defer server.Close()
	done := make(chan error, 2)
	// copy request to server
	go func() {
		n, err := io.Copy(client, server)
		slog.Debug("Message copied to server", "bytes", n)
		done <- err
	}()

	// copy response to client
	go func() {
		n, err := io.Copy(server, client)
		slog.Debug("Message copied to client", "bytes", n)
		done <- err
	}()

	<-done
	<-done
}

func (lb *LoadBalancer) StartHealthChecks(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				lb.checkAllBackends()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (lb *LoadBalancer) RoundRobinNextBackend() *Backend {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	i := (lb.current + 1) % len(lb.backends)
	for checked := 0; checked < len(lb.backends); checked++ {
		slog.Info("Parsing round robin")
		selected := lb.backends[i]
		if selected.Healthy {
			lb.current = i
			return selected
		}
		i = (i + 1) % len(lb.backends)
	}
	slog.Warn("Round robin did not select a backend")
	return lb.backends[i]
}

func checkBackend(b *Backend) bool {
	bConn, err := net.Dial("tcp", b.Host+":"+b.Port)
	if err != nil {
		slog.Warn("Error connecting to backend server", "error", err, "address", b.Host+":"+b.Port)
		return false
	}
	bConn.Close()
	return true
}

func (lb *LoadBalancer) checkAllBackends() {
	for _, b := range lb.backends {
		isHealthy := checkBackend(b)
		lb.mu.Lock()
		b.Healthy = isHealthy
		lb.mu.Unlock()
	}
}
