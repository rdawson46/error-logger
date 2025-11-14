package main

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    false,
	ReportTimestamp: true,
	TimeFormat:      time.Kitchen,
})

type Message struct {
	Code    uint8  `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Code uint8 `json:"code"`
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	logger.Info("New connection", "remote_addr", conn.RemoteAddr())

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection closed by client", "remote_addr", conn.RemoteAddr())
			} else {
				logger.Warn("Failed to read from connection", "remote_addr", conn.RemoteAddr(), "error", err)
			}
			return
		}

		var msg Message
		if json.Unmarshal(buf[:n], &msg) == nil {
			logger.Info("Received structured log:",
				"remote_addr", conn.RemoteAddr(),
				"code", msg.Code,
				"message", msg.Message,
			)

			response := Response{Code: 200}
			res, err := json.Marshal(response)
			if err != nil {
				logger.Error("Failed to marshal JSON response", "error", err)
				continue
			}
			conn.Write(res)
		} else {
			logger.Info("Received simple log:",
				"remote_addr", conn.RemoteAddr(),
				"mess", string(buf[:n]),
			)

			conn.Write([]byte("log received"))
		}
	}
}

func main() {
	server, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		logger.Fatal("Could not start server", "error", err)
	}

	defer server.Close()

	logger.Info("TCP server listening", "address", server.Addr())

	for {
		conn, err := server.Accept()

		if err != nil {
			logger.Warn("Failed to accept connection", "error", err)
			continue
		}

		go handleConnection(conn)
	}
}
