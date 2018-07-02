package vaultAuditExporter

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/hashicorp/vault/audit"
	log "github.com/sirupsen/logrus"
)

// Listen listens on the network and address specified and sends entries to
// the AuditHandler instance.
func Listen(network, address string, queue *AuditEntryQueue) error {
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

		go handleConnection(conn, queue)
	}
}

func handleConnection(conn net.Conn, queue *AuditEntryQueue) {
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
			handleRequest(lineBytes, queue)

		case "response":
			handleResponse(lineBytes, queue)

		default:
			log.WithFields(log.Fields{"type": entryType}).Warn("Received unknown audit entry type")
		}
	}
}

// Dummy struct to extract the `type` field
type auditEntry struct {
	Type string `json:"type"`
}

func getEntryType(lineBytes []byte) (entryType string, err error) {
	var entry auditEntry
	if err := json.Unmarshal(lineBytes, &entry); err != nil {
		return "", err
	}
	return entry.Type, nil
}

func handleRequest(lineBytes []byte, queue *AuditEntryQueue) {
	var req audit.AuditRequestEntry
	if err := json.Unmarshal(lineBytes, &req); err != nil {
		log.Error("Unable to unmarshal request audit entry")
		return
	}
	log.WithFields(log.Fields{"request_id": req.Request.ID}).Debug("Received request audit entry")
	queue.sendRequest(&req)
}

func handleResponse(lineBytes []byte, queue *AuditEntryQueue) {
	var res audit.AuditResponseEntry
	if err := json.Unmarshal(lineBytes, &res); err != nil {
		log.Error("Unable to unmarshal response audit entry")
		return
	}
	log.WithFields(log.Fields{"request_id": res.Request.ID}).Debug("Received response audit entry")
	queue.sendResponse(&res)
}

func closeConnection(conn net.Conn) {
	if err := conn.Close(); err != nil {
		log.Warn("Connection not closed cleanly", err)
	}
	log.WithFields(log.Fields{"remote_addr": conn.RemoteAddr()}).Info("Closed connection")
}
