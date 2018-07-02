package main

import (
	"flag"
	"fmt"

	"github.com/praekeltfoundation/vault-audit-exporter"
	"github.com/praekeltfoundation/vault-audit-exporter/version"
	log "github.com/sirupsen/logrus"
)

var (
	network string
	address string
)

func logAuditEntries(queue *vaultAuditExporter.AuditEntryQueue) {
	for {
		select {
		case req := <-queue.ReceiveRequest():
			log.Infof("Request: %+v", req)
		case res := <-queue.ReceiveResponse():
			log.Infof("Response: %+v", res)
		case <-queue.Done():
			log.Warn("Audit entry queue closed")
			return
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

	queue := vaultAuditExporter.NewAuditEntryQueue()
	defer queue.Close()

	go logAuditEntries(queue)

	if err := vaultAuditExporter.Listen(network, address, queue); err != nil {
		log.Fatal("Error listening for connections", err)
	}
}
