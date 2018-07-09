package auditexporter

func (ts *TestSuite) TestSendRequest() {
	q := NewAuditEntryQueue()
	ts.AddCleanup(q.Close)

	req := dummyRequest()
	go func() {
		q.sendRequest(req)
	}()

	ts.Equal(req, <-q.Receive())
}

func (ts *TestSuite) TestSendResponse() {
	q := NewAuditEntryQueue()
	ts.AddCleanup(q.Close)

	res := dummyResponse()
	go func() {
		q.sendResponse(res)
	}()

	ts.Equal(res, <-q.Receive())
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

	var entries []interface{} // nolint[prealloc]
	for entry := range q.Receive() {
		entries = append(entries, entry)
	}

	ts.Equal([]interface{}{req, res}, entries)
}
