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

	reqReceived := <-q.ReceiveRequest()
	ts.Equal(req, reqReceived)
}

func (ts *TestSuite) TestSendResponse() {
	q := NewAuditEntryQueue()
	ts.AddCleanup(q.Close)

	res := dummyResponse()
	go func() {
		q.sendResponse(res)
	}()

	resReceived := <-q.ReceiveResponse()
	ts.Equal(res, resReceived)
}

func (ts *TestSuite) TestDone() {
	q := NewAuditEntryQueue()

	q.Close()

	ts.NotNil(<-q.Done())
}
