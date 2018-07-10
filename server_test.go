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

func writeJSONLine(ts *TestSuite, v interface{}, writer io.Writer) {
	b, _ := ts.WithoutError(json.Marshal(v)).([]byte)

	ts.WithoutError(writer.Write(b))
	ts.WithoutError(writer.Write([]byte("\n")))
}

func writeLine(ts *TestSuite, s string, writer io.Writer) {
	ts.WithoutError(writer.Write([]byte(s + "\n")))
}

func (ts *TestSuite) TestHandleRequest() {
	server, client := net.Pipe()

	req := dummyRequest()
	go func() {
		writeJSONLine(ts, req, server)
		server.Close()
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
		writeJSONLine(ts, res, server)
		server.Close()
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
		writeJSONLine(ts, req, server)
		writeLine(ts, "{\"type\": \"foo\"}", server)
		writeJSONLine(ts, res, server)
		server.Close()
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.requests, 1)
	ts.Equal(req, collector.requests[0])
	ts.Len(collector.responses, 1)
	ts.Equal(res, collector.responses[0])
}
