package vaultAuditExporter

import (
	"encoding/json"
	"net"
)

func setupHander(ts *TestSuite) (net.Conn, *AuditEntryQueue) {
	server, client := net.Pipe()
	ts.AddCleanup(func() { ts.Nil(server.Close()) })

	queue := NewAuditEntryQueue()
	ts.AddCleanup(queue.Close)

	go handleConnection(client, queue)

	return server, queue
}

func writeJSONLine(ts *TestSuite, v interface{}, conn net.Conn) {
	b, _ := ts.WithoutError(json.Marshal(v)).([]byte)

	ts.WithoutError(conn.Write(b))
	ts.WithoutError(conn.Write([]byte("\n")))
}

func (ts *TestSuite) TestHandleRequest() {
	server, queue := setupHander(ts)

	req := dummyRequest()
	writeJSONLine(ts, req, server)

	ts.Equal(req, <-queue.Receive())
}

func (ts *TestSuite) TestHandleResponse() {
	server, queue := setupHander(ts)

	res := dummyResponse()
	writeJSONLine(ts, res, server)

	ts.Equal(res, <-queue.Receive())
}
