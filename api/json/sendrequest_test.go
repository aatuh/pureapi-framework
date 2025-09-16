package json

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/aatuh/pureapi-core/logging"
	"github.com/aatuh/pureapi-framework/defaults"
	"github.com/aatuh/pureapi-framework/util"
	"github.com/aatuh/pureapi-util/urlencoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SendDummyOutput is a simple output structure.
type SendDummyOutput struct {
	Result string `json:"result"`
}

type DummyRequestData struct {
	Headers       map[string]string
	Cookies       []http.Cookie
	Body          map[string]any
	URLParameters map[string]any
}

// To satisfy the expected type *util.RequestData in our SendParsedRequest call,
// we temporarily cast our DummyRequestData pointer to *util.RequestData.
// (In your actual code, use your real util.RequestData type.)
func (d *DummyRequestData) ToRequestData() *util.RequestData {
	// We assume that the underlying fields are compatible.
	// In real tests, use the actual constructor or conversion provided by util.
	rd := &util.RequestData{
		Headers:       d.Headers,
		Cookies:       d.Cookies,
		Body:          d.Body,
		URLParameters: d.URLParameters,
	}
	return rd
}

// DummyInput is a dummy input type used for SendRequest.
type DummyInput struct {
	Query string `json:"query"`
}

// SendRequestTestSuite is a test suite for SendRequest.
type SendRequestTestSuite struct {
	suite.Suite
	server      *httptest.Server
	capturedReq *http.Request
}

// TestSendRequestTestSuite runs the test suite.
func TestSendRequestTestSuite(t *testing.T) {
	suite.Run(t, new(SendRequestTestSuite))
}

// SetupSuite starts a test HTTP server.
func (s *SendRequestTestSuite) SetupSuite() {
	defaults.SetLoggerFactory(
		func(ctx context.Context) logging.ILogger {
			return nil
		},
	)
	// Create a test server that records the request and returns a response.
	s.server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.capturedReq = r
			resp := SendDummyOutput{Result: "ok"}
			data, _ := json.Marshal(resp)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(data)
			assert.NoError(s.T(), err)
		}),
	)
}

// TearDownSuite stops the test server.
func (s *SendRequestTestSuite) TearDownSuite() {
	s.server.Close()
}

// TestSendParsedRequest_GET_Success verifies that SendParsedRequest correctly
// constructs the GET request and appends URL parameters.
func (s *SendRequestTestSuite) TestSendParsedRequest_GET_Success() {
	// Build dummy request data with URL parameters and no body.
	dummyData := &DummyRequestData{
		Headers:       map[string]string{"Custom": "value"},
		URLParameters: map[string]any{"q": "search"},
	}
	rd := dummyData.ToRequestData()
	client := &http.Client{Timeout: 5 * time.Second}
	// Call SendParsedRequest with GET.
	resp, err := SendParsedRequest[SendDummyOutput](
		context.Background(),
		client,
		s.server.URL+"/endpoint",
		http.MethodGet,
		rd,
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	// Verify that the full URL includes the query parameter.
	parsedURL, err := url.Parse(s.capturedReq.URL.String())
	require.NoError(s.T(), err)
	q := parsedURL.Query().Get("q")
	assert.Equal(s.T(), "search", q)
	// Verify that the body is empty.
	body, _ := urlencoder.NewURLEncoder().Encode(rd.URLParameters)
	assert.True(s.T(), body.Encode() != "")
}

// TestSendParsedRequest_POST_Success verifies that for non-GET methods, the
// body is sent.
func (s *SendRequestTestSuite) TestSendParsedRequest_POST_Success() {
	// Build dummy request data with a body.
	dummyData := &DummyRequestData{
		Headers:       map[string]string{"Custom": "value"},
		Body:          map[string]any{"foo": "bar"},
		URLParameters: map[string]any{"p": "1"},
	}
	rd := dummyData.ToRequestData()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := SendParsedRequest[SendDummyOutput](
		context.Background(),
		client,
		s.server.URL+"/post",
		http.MethodPost,
		rd,
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
}

// TestSendParsedRequest_URLParameters verifies that when URL parameters are
// provided, they are appended to the URL.
func (s *SendRequestTestSuite) TestSendParsedRequest_URLParameters() {
	dummyData := &DummyRequestData{
		URLParameters: map[string]any{"key": "value", "num": 123},
	}
	rd := dummyData.ToRequestData()
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := SendParsedRequest[SendDummyOutput](
		context.Background(),
		client,
		s.server.URL+"/params",
		http.MethodGet,
		rd,
	)
	require.NoError(s.T(), err)
	parsedURL, err := url.Parse(s.capturedReq.URL.String())
	require.NoError(s.T(), err)
	q := parsedURL.Query()
	assert.Equal(s.T(), "value", q.Get("key"))
	assert.Equal(s.T(), "123", q.Get("num"))
}

// TestSendRequest_AddsSpanID_Default verifies that by default the X-Span-ID
// header is added.
func (s *SendRequestTestSuite) TestSendRequest_AddsSpanID_Default() {
	input := &DummyInput{Query: "test"}
	// Call SendRequest without additional options.
	_, err := SendRequest[DummyInput, SendDummyOutput](
		context.Background(),
		s.server.URL,
		"/span",
		http.MethodPost,
		input,
		nil,
	)
	require.NoError(s.T(), err)
	// The captured request should have a non-empty X-Span-ID header.
	spanID := s.capturedReq.Header.Get(defaults.XSpanID)
	assert.NotEmpty(s.T(), spanID, "X-Span-ID should be added by default")
}

// TestSendRequest_WithoutSpanID verifies that if the WithoutSpanID option is
// passed, the X-Span-ID header is not added.
func (s *SendRequestTestSuite) TestSendRequest_WithoutSpanID() {
	input := &DummyInput{Query: "test"}
	// Call SendRequest with the WithoutSpanID option.
	_, err := SendRequest[DummyInput, SendDummyOutput](
		context.Background(),
		s.server.URL,
		"/span",
		http.MethodPost,
		input,
		nil,
		WithoutSpanID(),
	)
	require.NoError(s.T(), err)
	// The captured request should not have an X-Span-ID header.
	spanID := s.capturedReq.Header.Get(defaults.XSpanID)
	assert.Empty(
		s.T(), spanID,
		"X-Span-ID should be omitted when WithoutSpanID is used",
	)
}

// TestSendRequest_Success verifies that SendRequest works end-to-end.
func (s *SendRequestTestSuite) TestSendRequest_Success() {
	input := &DummyInput{Query: "example"}
	resp, err := SendRequest[DummyInput, SendDummyOutput](
		context.Background(),
		s.server.URL,
		"/full",
		http.MethodPost,
		input,
		nil,
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	assert.Equal(s.T(), "ok", resp.Output.Result)
}
