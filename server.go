package vaultAuditExporter

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/hashicorp/vault/audit"
	log "github.com/sirupsen/logrus"
)

type reqChan chan *audit.AuditRequestEntry
type resChan chan *audit.AuditResponseEntry

// Listen listens on the network and address specified and sends entries to
// the AuditHandler instance.
func Listen(network, address string, reqChan reqChan, resChan resChan) error {
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

		go handleConnection(conn, reqChan, resChan)
	}
}

func handleConnection(conn net.Conn, reqChan reqChan, resChan resChan) {
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
			handleRequest(lineBytes, reqChan)

		case "response":
			handleResponse(lineBytes, resChan)

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

func handleRequest(lineBytes []byte, reqChan reqChan) {
	var req audit.AuditRequestEntry
	if err := json.Unmarshal(lineBytes, &req); err != nil {
		log.Error("Unable to unmarshal request audit entry")
		return
	}
	log.WithFields(log.Fields{"request_id": req.Request.ID}).Debug("Received request audit entry")
	reqChan <- &req
}

func handleResponse(lineBytes []byte, resChan resChan) {
	var res audit.AuditResponseEntry
	if err := json.Unmarshal(lineBytes, &res); err != nil {
		log.Error("Unable to unmarshal response audit entry")
		return
	}
	log.WithFields(log.Fields{"request_id": res.Request.ID}).Debug("Received response audit entry")
	resChan <- &res
}

func closeConnection(conn net.Conn) {
	if err := conn.Close(); err != nil {
		log.Warn("Connection not closed cleanly")
	}
	log.WithFields(log.Fields{"remote_addr": conn.RemoteAddr()}).Info("Closed connection")
}
