package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/dylanmccormick/go-load-balancer/internal/balancer"
)

const (
	ConnHost = ""
	ConnPort = "8080"
	ConnType = "tcp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	ln, err := net.Listen(ConnType, ConnHost+":"+ConnPort)
	if err != nil {
		slog.Error("Error starting TCP server", "error", err)
		os.Exit(1)
	}

	lb := balancer.CreateLB()

	lb.CreateAndServeBackends()
	lb.StartHealthChecks(context.TODO())

	slog.Info("Server listening on port :" + ConnPort)
	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Error accepting connection", "error", err)
		}

		go lb.HandleProxy(conn)
	}
}
