package auditexporter

import (
	"encoding/json"
	"io"
	"net"
	"testing"

	"github.com/hashicorp/vault/audit"
	"github.com/stretchr/testify/suite"
)

// See helper_for_test.go for common infrastructure and tools.

// ServerTests is a testify test suite object that we can attach helper methods
// to.
type ServerTests struct{ TestSuite }

// TestServer is a standard Go test function that runs our test suite's tests.
func TestServer(t *testing.T) { suite.Run(t, new(ServerTests)) }

func (ts *ServerTests) writeJSONLine(writer io.Writer, v interface{}) {
	b, err := json.Marshal(v)
	ts.NoError(err)
	ts.writeLine(writer, b)
}

func (ts *ServerTests) writeStringLine(writer io.Writer, line string) {
	ts.write(writer, []byte(line), []byte{'\n'})
}

func (ts *ServerTests) writeLine(writer io.Writer, line []byte) {
	ts.write(writer, line, []byte{'\n'})
}

func (ts *ServerTests) write(writer io.Writer, bytes ...[]byte) {
	for _, b := range bytes {
		_, err := writer.Write(b)
		ts.NoError(err)
	}
}

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

func (ts *ServerTests) TestHandleRequest() {
	server, client := net.Pipe()

	req := dummyRequest()
	go func() {
		defer server.Close()
		ts.writeJSONLine(server, req)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.requests, 1)
	ts.Equal(req, collector.requests[0])
}

func (ts *ServerTests) TestHandleResponse() {
	server, client := net.Pipe()

	res := dummyResponse()
	go func() {
		defer server.Close()
		ts.writeJSONLine(server, res)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.responses, 1)
	ts.Equal(res, collector.responses[0])
}

func (ts *ServerTests) TestUnknownTypeIgnored() {
	server, client := net.Pipe()

	req := dummyRequest()
	res := dummyResponse()
	go func() {
		defer server.Close()
		ts.writeJSONLine(server, req)
		ts.writeStringLine(server, "{\"type\": \"foo\"}")
		ts.writeJSONLine(server, res)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.requests, 1)
	ts.Equal(req, collector.requests[0])
	ts.Len(collector.responses, 1)
	ts.Equal(res, collector.responses[0])
}

func (ts *ServerTests) TestUnknownJSON() {
	server, client := net.Pipe()
	defer server.Close()

	go func() {
		ts.writeStringLine(server, "{\"foo\": \"bar\"}")
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Empty(collector.requests)
	ts.Empty(collector.responses)
}

func (ts *ServerTests) TestInvalidJSON() {
	server, client := net.Pipe()
	defer server.Close()

	go func() {
		ts.writeStringLine(server, "baz")
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Empty(collector.requests)
	ts.Empty(collector.responses)
}

func (ts *ServerTests) TestServe() {
	listener, _ := ts.WithoutError(net.Listen("tcp", "127.0.0.1:0")).(net.Listener)
	ts.AddCleanup(func() { _ = listener.Close() })

	// We have to use the queue here because handling happens in a separate
	// goroutine
	queue := NewAuditEntryQueue()
	ts.AddCleanup(queue.Close)

	// Start serving
	go func() {
		_ = Serve(listener, queue)
	}()

	// Dial into the server
	addr := listener.Addr()
	conn, _ := ts.WithoutError(net.Dial(addr.Network(), addr.String())).(net.Conn)

	// Send some entries
	req := dummyRequest()
	res := dummyResponse()
	go func() {
		defer conn.Close()
		ts.writeJSONLine(conn, req)
		ts.writeJSONLine(conn, res)
	}()

	// Ensure they are received in the queue
	ts.Equal(req, <-queue.Receive())
	ts.Equal(res, <-queue.Receive())
}
