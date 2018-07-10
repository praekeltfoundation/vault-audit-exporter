package auditexporter

import (
	"encoding/json"
	"io"
	"net"

	"github.com/hashicorp/vault/audit"
)

type auditEntryCollector struct {
	requests  chan *audit.AuditRequestEntry
	responses chan *audit.AuditResponseEntry
}

func newAuditEntryCollector() *auditEntryCollector {
	return &auditEntryCollector{
		requests:  make(chan *audit.AuditRequestEntry),
		responses: make(chan *audit.AuditResponseEntry),
	}
}

func (col *auditEntryCollector) HandleRequest(req *audit.AuditRequestEntry) {
	col.requests <- req
}

func (col *auditEntryCollector) HandleResponse(res *audit.AuditResponseEntry) {
	col.responses <- res
}

func (col *auditEntryCollector) close() {
	close(col.requests)
	close(col.responses)
}

func setupHander(ts *TestSuite) (net.Conn, *auditEntryCollector) {
	server, client := net.Pipe()
	ts.AddCleanup(func() { ts.Nil(server.Close()) })

	collector := newAuditEntryCollector()
	ts.AddCleanup(collector.close)

	go handleConnection(client, collector)

	return server, collector
}

func writeJSONLine(ts *TestSuite, v interface{}, writer io.Writer) {
	b, _ := ts.WithoutError(json.Marshal(v)).([]byte)

	ts.WithoutError(writer.Write(b))
	ts.WithoutError(writer.Write([]byte("\n")))
}

func (ts *TestSuite) TestHandleRequest() {
	server, collector := setupHander(ts)

	req := dummyRequest()
	writeJSONLine(ts, req, server)

	ts.Equal(req, <-collector.requests)
}

func (ts *TestSuite) TestHandleResponse() {
	server, collector := setupHander(ts)

	res := dummyResponse()
	writeJSONLine(ts, res, server)

	ts.Equal(res, <-collector.responses)
}
