package auditexporter

func (ts *TestSuite) TestSendRequest() {
	q := NewAuditEntryQueue()
	ts.AddCleanup(q.Close)

	req := dummyRequest()
	go func() {
		q.HandleRequest(req)
	}()

	ts.Equal(req, <-q.Receive())
}

func (ts *TestSuite) TestSendResponse() {
	q := NewAuditEntryQueue()
	ts.AddCleanup(q.Close)

	res := dummyResponse()
	go func() {
		q.HandleResponse(res)
	}()

	ts.Equal(res, <-q.Receive())
}

func (ts *TestSuite) TestClose() {
	q := NewAuditEntryQueue()

	req := dummyRequest()
	res := dummyResponse()
	go func() {
		q.HandleRequest(req)
		q.HandleResponse(res)
		q.Close()
	}()

	var entries []interface{} // nolint[prealloc]
	for entry := range q.Receive() {
		entries = append(entries, entry)
	}

	ts.Equal([]interface{}{req, res}, entries)
}
