package vaultAuditExporter

import (
	"github.com/hashicorp/vault/audit"
)

// AuditEntryQueue is a queue for audit entries.
type AuditEntryQueue struct {
	reqChan chan *audit.AuditRequestEntry
	resChan chan *audit.AuditResponseEntry
}

// NewAuditEntryQueue creates an new AuditEntryQueue
// TODO: Add buffer size, send timeout, dropping on error?
func NewAuditEntryQueue() *AuditEntryQueue {
	return &AuditEntryQueue{
		reqChan: make(chan *audit.AuditRequestEntry),
		resChan: make(chan *audit.AuditResponseEntry),
	}
}

// Close closes the underlying channels
func (ah AuditEntryQueue) Close() {
	close(ah.reqChan)
	close(ah.resChan)
}

// ReceiveRequest returns a channel to receive AuditRequestEntry instances from.
func (ah AuditEntryQueue) ReceiveRequest() <-chan *audit.AuditRequestEntry {
	return ah.reqChan
}

// ReceiveResponse returns a channel to receive AuditResponseEntry instances
// from.
func (ah AuditEntryQueue) ReceiveResponse() <-chan *audit.AuditResponseEntry {
	return ah.resChan
}

func (ah AuditEntryQueue) sendRequest(req *audit.AuditRequestEntry) {
	ah.reqChan <- req
}

func (ah AuditEntryQueue) sendResponse(res *audit.AuditResponseEntry) {
	ah.resChan <- res
}
