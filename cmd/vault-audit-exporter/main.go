package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/hashicorp/vault/audit"
	"github.com/praekeltfoundation/vault-audit-exporter"
	"github.com/praekeltfoundation/vault-audit-exporter/version"
	log "github.com/sirupsen/logrus"
)

var (
	network string
	address string
)

func logAuditEntries(queue *auditexporter.AuditEntryQueue) {
	for entry := range queue.Receive() {
		switch entry.(type) {
		case *audit.AuditRequestEntry:
			log.Infof("Request: %+v", entry)
		case *audit.AuditResponseEntry:
			log.Infof("Response: %+v", entry)
		default:
			log.Warn("Unknown audit entry type")
		}
	}
}

func main() {
	flag.StringVar(&network, "network", "tcp", "network type to listen on (e.g. tcp, tcp6, unix, ...)")
	flag.StringVar(&address, "address", "127.0.0.1:9090", "address to listen on (e.g. 127.0.0.1:9090)")

	versionFlag := flag.Bool("version", false, "Print version information and exit.")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.HumanReadable())
		return
	}

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatal("Error listening for connections", err)
	}
	log.WithFields(log.Fields{"addr": listener.Addr()}).Info("Listening...")

	queue := auditexporter.NewAuditEntryQueue()
	defer queue.Close()
	go logAuditEntries(queue)

	if err := auditexporter.Serve(listener, queue); err != nil {
		log.Fatal("Error accepting connections", err)
	}
}
