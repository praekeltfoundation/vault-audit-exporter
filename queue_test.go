package vaultAuditExporter

import (
	"github.com/hashicorp/vault/audit"
	"github.com/satori/go.uuid"
)

func dummyRequest() *audit.AuditRequestEntry {
	return &audit.AuditRequestEntry{
		Type: "request",
		Request: audit.AuditRequest{
			ID: uuid.NewV4().String(),
		},
	}
}

func dummyResponse() *audit.AuditResponseEntry {
	return &audit.AuditResponseEntry{
		Type: "response",
		Request: audit.AuditRequest{
			ID: uuid.NewV4().String(),
		},
	}
}

func (ts *TestSuite) TestSendRequest() {
	q := NewAuditEntryQueue()
	ts.AddCleanup(q.Close)

	req := dummyRequest()
	go func() {
		q.sendRequest(req)
	}()

	entry := <-q.Receive()
	ts.IsType(&audit.AuditRequestEntry{}, entry)
	ts.Equal(req, entry)
}

func (ts *TestSuite) TestSendResponse() {
	q := NewAuditEntryQueue()
	ts.AddCleanup(q.Close)

	res := dummyResponse()
	go func() {
		q.sendResponse(res)
	}()

	entry := <-q.Receive()
	ts.IsType(&audit.AuditResponseEntry{}, entry)
	ts.Equal(res, entry)
}

func (ts *TestSuite) TestClose() {
	q := NewAuditEntryQueue()

	req := dummyRequest()
	res := dummyResponse()
	go func() {
		q.sendRequest(req)
		q.sendResponse(res)
		q.Close()
	}()

	var entries []interface{}
	for entry := range q.Receive() {
		entries = append(entries, entry)
	}

	ts.Equal([]interface{}{req, res}, entries)
}
