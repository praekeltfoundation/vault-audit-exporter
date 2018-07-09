package auditexporter

import (
	"encoding/json"
	"io"
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

func writeJSONLine(ts *TestSuite, v interface{}, writer io.Writer) {
	b, _ := ts.WithoutError(json.Marshal(v)).([]byte)

	ts.WithoutError(writer.Write(b))
	ts.WithoutError(writer.Write([]byte("\n")))
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
