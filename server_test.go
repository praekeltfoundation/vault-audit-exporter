package auditexporter

import (
	"encoding/json"
	"io"
	"net"

	"github.com/hashicorp/vault/audit"
)

type auditEntryCollector struct {
	requests  []*audit.AuditRequestEntry
	responses []*audit.AuditResponseEntry
}

func newAuditEntryCollector() *auditEntryCollector {
	return &auditEntryCollector{
		requests:  make([]*audit.AuditRequestEntry, 0),
		responses: make([]*audit.AuditResponseEntry, 0),
	}
}

func (col *auditEntryCollector) HandleRequest(req *audit.AuditRequestEntry) {
	col.requests = append(col.requests, req)
}

func (col *auditEntryCollector) HandleResponse(res *audit.AuditResponseEntry) {
	col.responses = append(col.responses, res)
}

func writeJSONLine(v interface{}, writer io.Writer) {
	b, _ := json.Marshal(v)
	writeLine(b, writer)
}

func writeStringLine(s string, writer io.Writer) {
	writeLine([]byte(s), writer)
}

func writeLine(b []byte, writer io.Writer) {
	_, _ = writer.Write(b)
	_, _ = writer.Write([]byte{'\n'})
}

func (ts *TestSuite) TestHandleRequest() {
	server, client := net.Pipe()

	req := dummyRequest()
	go func() {
		defer server.Close()
		writeJSONLine(req, server)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.requests, 1)
	ts.Equal(req, collector.requests[0])
}

func (ts *TestSuite) TestHandleResponse() {
	server, client := net.Pipe()

	res := dummyResponse()
	go func() {
		defer server.Close()
		writeJSONLine(res, server)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.responses, 1)
	ts.Equal(res, collector.responses[0])
}

func (ts *TestSuite) TestUnknownTypeIgnored() {
	server, client := net.Pipe()

	req := dummyRequest()
	res := dummyResponse()
	go func() {
		defer server.Close()
		writeJSONLine(req, server)
		writeStringLine("{\"type\": \"foo\"}", server)
		writeJSONLine(res, server)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.requests, 1)
	ts.Equal(req, collector.requests[0])
	ts.Len(collector.responses, 1)
	ts.Equal(res, collector.responses[0])
}

func (ts *TestSuite) TestUnknownJSON() {
	server, client := net.Pipe()
	defer server.Close()

	go func() {
		writeStringLine("{\"foo\": \"bar\"}", server)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Empty(collector.requests)
	ts.Empty(collector.responses)
}

func (ts *TestSuite) TestInvalidJSON() {
	server, client := net.Pipe()
	defer server.Close()

	go func() {
		writeStringLine("baz", server)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Empty(collector.requests)
	ts.Empty(collector.responses)
}
