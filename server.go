package vaultAuditExporter

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/hashicorp/vault/audit"
	log "github.com/sirupsen/logrus"
)

// AuditHandler handles audit entries as they are received.
type AuditHandler interface {
  HandleRequest(*audit.AuditRequestEntry)
  HandleResponse(*audit.AuditResponseEntry)
}

// Dummy struct to extract the `type` field
type auditEntry struct {
	Type string `json:"type"`
}

// Listen listens on the network and address specified and sends entries to
// the AuditHandler instance.
func Listen(network, address string, handler AuditHandler) error {
	ln, err := net.Listen(network, address)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"network": network,
		"address": address,
	}).Info("Listening...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{"remote_addr": conn.RemoteAddr()}).Info("Accepted connection")

		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler AuditHandler) {
	defer closeConnection(conn)

	var entry auditEntry
	var req audit.AuditRequestEntry
	var res audit.AuditResponseEntry

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines) // The default but I'd rather be explicit
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		log.Debug("Line received")

		if err := json.Unmarshal(lineBytes, &entry); err != nil {
			log.WithFields(log.Fields{"line": scanner.Text()}).Error("Unable to unmarshal audit entry")
			return
		}
		log.WithFields(log.Fields{"type": entry.Type}).Debug("Received audit entry")

		switch entry.Type {
		case "request":
			if err := json.Unmarshal(lineBytes, &req); err != nil {
				log.Error("Unable to unmarshal request audit entry")
				return
			}
			log.WithFields(log.Fields{"request_id": req.Request.ID}).Debug("Received request audit entry")
			handler.HandleRequest(&req)

		case "response":
			if err := json.Unmarshal(lineBytes, &res); err != nil {
				log.Error("Unable to unmarshal response audit entry")
				return
			}
			log.WithFields(log.Fields{"request_id": res.Request.ID}).Debug("Received response audit entry")
			handler.HandleResponse(&res)

		default:
			log.WithFields(log.Fields{"type": entry.Type}).Warn("Received unknown audit entry type")
		}
	}
}

func closeConnection(conn net.Conn) {
	if err := conn.Close(); err != nil {
		log.Warn("Connection not closed cleanly")
	}
	log.WithFields(log.Fields{"remote_addr": conn.RemoteAddr()}).Info("Closed connection")
}
