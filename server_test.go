package auditexporter

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/vault/audit"
	"github.com/stretchr/testify/suite"
)

// ListenAndServe doesn't allow us to fetch the port of the listener, so we have
// to use a fixed one.
const TestPort = 12345

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

// In some of the tests it's necessary to wait for a goroutine to finish before
// the tests end in order to avoid a race condition in the test suite.
func waitableGoroutine(f func()) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
	return &wg
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

func (ts *ServerTests) TestInvalidRequest() {
	server, client := net.Pipe()

	req := dummyRequest()
	go func() {
		defer server.Close()
		// The 'time' field should have type string.
		ts.writeStringLine(server, "{\"type\": \"request\", \"time\": false}")

		// We can still send requests after an invalid one is ignored
		ts.writeJSONLine(server, req)
	}()

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	ts.Len(collector.requests, 1)
	ts.Equal(req, collector.requests[0])
}

func (ts *ServerTests) TestInvalidResponse() {
	server, client := net.Pipe()

	res := dummyResponse()
	go func() {
		defer server.Close()
		// The 'time' field should have type string.
		ts.writeStringLine(server, "{\"type\": \"response\", \"time\": false}")

		// We can still send responses after an invalid one is ignored
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

	// The error in the server that results in the connection closing happens
	// before the error is checked in ts.writeStringLine(), so it's necessary to
	// check that this goroutine completes before the end of the test.
	wg := waitableGoroutine(func() {
		ts.writeStringLine(server, "{\"foo\": \"bar\"}")
	})

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	// Ensure we can't write anymore because the connection was closed
	_, err := server.Write([]byte("foobar\n"))
	ts.Error(err)

	ts.Empty(collector.requests)
	ts.Empty(collector.responses)

	wg.Wait()
}

func (ts *ServerTests) TestInvalidJSON() {
	server, client := net.Pipe()
	defer server.Close()

	// The error in the server that results in the connection closing happens
	// before the error is checked in ts.writeStringLine(), so it's necessary to
	// check that this goroutine completes before the end of the test.
	wg := waitableGoroutine(func() {
		ts.writeStringLine(server, "baz")
	})

	collector := newAuditEntryCollector()
	handleConnection(client, collector)

	// Ensure we can't write anymore because the connection was closed
	_, err := server.Write([]byte("foobar\n"))
	ts.Error(err)

	ts.Empty(collector.requests)
	ts.Empty(collector.responses)

	wg.Wait()
}

func (ts *ServerTests) TestListenAndServe() {
	addr := fmt.Sprintf("127.0.0.1:%d", TestPort)

	// We have to use the queue here because handling happens in a separate
	// goroutine
	queue := NewAuditEntryQueue()
	ts.AddCleanup(queue.Close)

	// Start serving
	go func() {
		_ = ListenAndServe(addr, queue) // #nosec G104
	}()

	// Dial into the server - try a few times because we don't know when listening
	// has started
	var (
		conn net.Conn
		err  error
	)
	for i := 0; i < 5; i++ {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	ts.Require().NoError(err)

	// Send some entries
	req := dummyRequest()
	res := dummyResponse()
	// The responses from the server that may be received from the queue before
	// the error is checked in ts.writeJSONLine(), so it's necessary to check
	// that this goroutine completes before the end of the test.
	wg := waitableGoroutine(func() {
		defer conn.Close()
		ts.writeJSONLine(conn, req)
		ts.writeJSONLine(conn, res)
	})

	// Ensure they are received in the queue
	ts.Equal(req, <-queue.Receive())
	ts.Equal(res, <-queue.Receive())

	wg.Wait()
}

func (ts *ServerTests) TestListenAndServeInvalidAddr() {
	err := ListenAndServe("127.0.0.1:123456", newAuditEntryCollector())
	ts.EqualError(err, "listen tcp: address 123456: invalid port")
}

func (ts *ServerTests) TestServe() {
	listener, _ := ts.WithoutError(net.Listen("tcp", "127.0.0.1:0")).(net.Listener)
	ts.AddCleanup(func() { _ = listener.Close() }) // #nosec G104

	// We have to use the queue here because handling happens in a separate
	// goroutine
	queue := NewAuditEntryQueue()
	ts.AddCleanup(queue.Close)

	// Start serving
	go func() {
		_ = Serve(listener, queue) // #nosec G104
	}()

	// Dial into the server
	addr := listener.Addr()
	conn, _ := ts.WithoutError(net.Dial(addr.Network(), addr.String())).(net.Conn)

	// Send some entries
	req := dummyRequest()
	res := dummyResponse()
	// The responses from the server that may be received from the queue before
	// the error is checked in ts.writeJSONLine(), so it's necessary to check
	// that this goroutine completes before the end of the test.
	wg := waitableGoroutine(func() {
		defer conn.Close()
		ts.writeJSONLine(conn, req)
		ts.writeJSONLine(conn, res)
	})

	// Ensure they are received in the queue
	ts.Equal(req, <-queue.Receive())
	ts.Equal(res, <-queue.Receive())

	wg.Wait()
}
