package vaultAuditExporter

import (
	"github.com/hashicorp/vault/audit"
)

// AuditEntryQueue is a q for audit entries.
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
func (q AuditEntryQueue) Close() {
	close(q.reqChan)
	close(q.resChan)
}

// ReceiveRequest returns a channel to receive AuditRequestEntry instances from.
func (q AuditEntryQueue) ReceiveRequest() <-chan *audit.AuditRequestEntry {
	return q.reqChan
}

// ReceiveResponse returns a channel to receive AuditResponseEntry instances
// from.
func (q AuditEntryQueue) ReceiveResponse() <-chan *audit.AuditResponseEntry {
	return q.resChan
}

func (q AuditEntryQueue) sendRequest(req *audit.AuditRequestEntry) {
	q.reqChan <- req
}

func (q AuditEntryQueue) sendResponse(res *audit.AuditResponseEntry) {
	q.resChan <- res
}
