package auditexporter

import (
	"github.com/hashicorp/vault/audit"
)

// AuditEntryQueue is a q for audit entries.
type AuditEntryQueue struct {
	channel chan interface{}
}

// NewAuditEntryQueue creates an new AuditEntryQueue
// TODO: Add buffer size, send timeout, dropping on error?
func NewAuditEntryQueue() *AuditEntryQueue {
	return &AuditEntryQueue{channel: make(chan interface{})}
}

// Close closes the underlying channels
func (q *AuditEntryQueue) Close() {
	close(q.channel)
}

// Receive returns a channel to receive AuditEntry instances from.
func (q *AuditEntryQueue) Receive() <-chan interface{} {
	return q.channel
}

func (q *AuditEntryQueue) HandleRequest(req *audit.AuditRequestEntry) {
	q.channel <- req
}

func (q *AuditEntryQueue) HandleResponse(res *audit.AuditResponseEntry) {
	q.channel <- res
}
