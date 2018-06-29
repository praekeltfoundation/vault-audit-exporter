package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/vault/audit"
	"github.com/praekeltfoundation/vault-audit-exporter"
	"github.com/praekeltfoundation/vault-audit-exporter/version"
	log "github.com/sirupsen/logrus"
)

var (
	network string
	address string
)

func logRequests(c chan *audit.AuditRequestEntry) {
	for req := range c {
		log.WithFields(log.Fields{"entry": fmt.Sprintf("%+v", req)}).Info("Request")
	}
}
func logResponses(c chan *audit.AuditResponseEntry) {
	for res := range c {
		log.WithFields(log.Fields{"entry": fmt.Sprintf("%+v", res)}).Info("Response")
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

	reqChan := make(chan *audit.AuditRequestEntry)
	resChan := make(chan *audit.AuditResponseEntry)

	go logRequests(reqChan)
	go logResponses(resChan)

	if err := vaultAuditExporter.Listen(network, address, reqChan, resChan); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error listening")
		os.Exit(1)
	}
}
