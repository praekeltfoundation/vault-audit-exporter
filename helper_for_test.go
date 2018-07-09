package vaultAuditExporter

import (
	"testing"

	"github.com/hashicorp/vault/audit"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
)

// TestSuite is a testify test suite object that we can attach helper methods
// to.
type TestSuite struct {
	suite.Suite
	cleanups []func()
}

func Test_TestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// AddCleanup schedules the given cleanup function to be run after the test.
// Think of it like `defer`, except it applies to the whole test rather than
// the specific function it appears in.
func (ts *TestSuite) AddCleanup(f func()) {
	ts.cleanups = append([]func(){f}, ts.cleanups...)
}

// SetupTest clears all our TestSuite state at the start of each test, because
// the same object is shared across all tests.
func (ts *TestSuite) SetupTest() {
	ts.cleanups = []func(){}
}

// TearDownTest calls the registered cleanup functions.
func (ts *TestSuite) TearDownTest() {
	for _, f := range ts.cleanups {
		f()
	}
}

// WithoutError accepts a (result, error) pair, immediately fails the test if
// there is an error, and returns just the result if there is no error. It
// accepts and returns the result value as an `interface{}`, so it may need to
// be cast back to whatever type it should be afterwards.
func (ts *TestSuite) WithoutError(result interface{}, err error) interface{} {
	ts.T().Helper()
	ts.Require().NoError(err)
	return result
}

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
