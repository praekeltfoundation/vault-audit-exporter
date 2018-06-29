package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/praekeltfoundation/vault-audit-exporter"
	"github.com/praekeltfoundation/vault-audit-exporter/version"
	log "github.com/sirupsen/logrus"
)

var (
	network string
	address string
)

func logAuditEntries(queue *vaultAuditExporter.AuditEntryQueue) {
	select {
	case req := <-queue.ReceiveRequest():
		log.WithFields(log.Fields{"entry": fmt.Sprintf("%+v", req)}).Info("Request")
	case res := <-queue.ReceiveResponse():
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

	queue := vaultAuditExporter.NewAuditEntryQueue()
	defer queue.Close()

	go logAuditEntries(queue)

	if err := vaultAuditExporter.Listen(network, address, queue); err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error listening")
		os.Exit(1)
	}
}
