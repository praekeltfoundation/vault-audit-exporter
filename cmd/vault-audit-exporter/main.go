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

type logHandler struct {}
func (lh logHandler) HandleRequest(req *audit.AuditRequestEntry) {
	log.WithFields(log.Fields{"entry": req}).Info("Request")
}
func (lh logHandler) HandleResponse(res *audit.AuditResponseEntry) {
	log.WithFields(log.Fields{"entry": res}).Info("Response")
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

	var handler logHandler
	if err := vaultAuditExporter.Listen(network, address, &handler); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error listening")
		os.Exit(1)
	}
}
