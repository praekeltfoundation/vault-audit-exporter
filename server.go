package auditexporter

import (
	"bufio"
	"encoding/json"
	"net"
	"time"

	"github.com/hashicorp/vault/audit"
	log "github.com/sirupsen/logrus"
)

// A Handler deals with incoming audit entries.
type Handler interface {
	HandleRequest(*audit.AuditRequestEntry)
	HandleResponse(*audit.AuditResponseEntry)
}

// ListenAndServe starts a TCP listener on the given address and serves
// requests using the given handler.
func ListenAndServe(addr string, handler Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"addr": listener.Addr()}).Info("Listening...")

	return Serve(listener, handler)
}

// Serve loops forever accepting and handling connections.
func Serve(listener net.Listener, handler Handler) error {
	// net/http Server also closes the listener in Serve
	defer closeListener(listener)

	for {
		conn, err := listener.Accept()
		if err != nil {
			// This logic is inspired by the stdlib net/http server:
			// https://go.googlesource.com/go/+/go1.10.3/src/net/http/server.go#2777
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				// net/http does a fancy exponential backoff... we just do a constant
				// time for simplicity.
				log.Warnf("Accept error: %v; retrying in 100ms", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return err
		}

		log.WithFields(log.Fields{"remote_addr": conn.RemoteAddr()}).Info("Accepted connection")

		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler Handler) {
	defer closeConnection(conn)

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines) // The default but I'd rather be explicit
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		log.Debug("Line received")

		entryType, err := getEntryType(lineBytes)
		if err != nil {
			log.WithFields(log.Fields{"line": scanner.Text()}).Error("Unable to determine audit entry type")
			return
		}
		log.WithFields(log.Fields{"type": entryType}).Debug("Received audit entry")

		switch entryType {
		case "request":
			handleRequest(lineBytes, handler)

		case "response":
			handleResponse(lineBytes, handler)

		default:
			log.WithFields(log.Fields{"type": entryType}).Warn("Received unknown audit entry type")
		}
	}
}

// Dummy struct to extract the `type` field
type auditEntry struct {
	Type string `json:"type"`
}

func getEntryType(lineBytes []byte) (string, error) {
	var entry auditEntry
	err := json.Unmarshal(lineBytes, &entry)
	return entry.Type, err
}

func handleRequest(lineBytes []byte, handler Handler) {
	var req audit.AuditRequestEntry
	if err := json.Unmarshal(lineBytes, &req); err != nil {
		log.Error("Unable to unmarshal request audit entry")
		return
	}
	log.WithFields(log.Fields{"request_id": req.Request.ID}).Debug("Received request audit entry")
	handler.HandleRequest(&req)
}

func handleResponse(lineBytes []byte, handler Handler) {
	var res audit.AuditResponseEntry
	if err := json.Unmarshal(lineBytes, &res); err != nil {
		log.Error("Unable to unmarshal response audit entry")
		return
	}
	log.WithFields(log.Fields{"request_id": res.Request.ID}).Debug("Received response audit entry")
	handler.HandleResponse(&res)
}

func closeListener(listener net.Listener) {
	if err := listener.Close(); err != nil {
		log.Warn("Listening not stopped cleanly", err)
	}
	log.WithFields(log.Fields{"addr": listener.Addr()}).Info("Stopped listening")
}

func closeConnection(conn net.Conn) {
	if err := conn.Close(); err != nil {
		log.Warn("Connection not closed cleanly", err)
	}
	log.WithFields(log.Fields{"remote_addr": conn.RemoteAddr()}).Info("Closed connection")
}
