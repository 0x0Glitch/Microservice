package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer wraps httptest.Server with additional utilities
type TestServer struct {
	*httptest.Server
	t *testing.T
}

// NewTestServer creates a new test server
func NewTestServer(t *testing.T, handler http.Handler) *TestServer {
	server := httptest.NewServer(handler)
	return &TestServer{
		Server: server,
		t:      t,
	}
}

// PostJSON sends a POST request with JSON body
func (ts *TestServer) PostJSON(path string, body interface{}) *http.Response {
	jsonBody, err := json.Marshal(body)
	require.NoError(ts.t, err)

	resp, err := http.Post(ts.URL+path, "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(ts.t, err)
	return resp
}

// GetJSON sends a GET request and expects JSON response
func (ts *TestServer) GetJSON(path string, target interface{}) *http.Response {
	resp, err := http.Get(ts.URL + path)
	require.NoError(ts.t, err)

	if target != nil {
		err = json.NewDecoder(resp.Body).Decode(target)
		require.NoError(ts.t, err)
	}
	return resp
}

// AssertJSONResponse asserts the response status and decodes JSON body
func AssertJSONResponse(t *testing.T, resp *http.Response, expectedStatus int, target interface{}) {
	assert.Equal(t, expectedStatus, resp.StatusCode)

	if target != nil {
		err := json.NewDecoder(resp.Body).Decode(target)
		require.NoError(t, err)
	}
	resp.Body.Close()
}

// AssertErrorResponse asserts an error response
func AssertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedError string) {
	assert.Equal(t, expectedStatus, resp.StatusCode)

	var errorResp map[string]string
	err := json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["error"], expectedError)
	resp.Body.Close()
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutCh:
			t.Fatalf("Timeout waiting for condition: %s", message)
		}
	}
}

// GenerateTestOBUData creates test OBU data
func GenerateTestOBUData(obuID int32, lat, long float64) types.OBUData {
	return types.OBUData{
		OBUID:     obuID,
		Lat:       lat,
		Long:      long,
		RequestID: int(time.Now().UnixNano() % 1000000),
	}
}

// GenerateTestDistance creates test distance data
func GenerateTestDistance(obuID int32, value float64) types.Distance {
	return types.Distance{
		OBUID:  obuID,
		Values: value,
		Unix:   time.Now().UnixNano(),
	}
}

// GenerateTestInvoice creates test invoice data
func GenerateTestInvoice(obuID int32, totalDistance, amount float64) types.Invoice {
	return types.Invoice{
		OBUID:         obuID,
		TotalDistance: totalDistance,
		Amount:        amount,
	}
}

// FloatEquals compares floats with tolerance
func FloatEquals(a, b, tolerance float64) bool {
	return (a-b) < tolerance && (b-a) < tolerance
}

// AssertFloatEquals asserts float equality with tolerance
func AssertFloatEquals(t *testing.T, expected, actual, tolerance float64, msgAndArgs ...interface{}) {
	if !FloatEquals(expected, actual, tolerance) {
		t.Errorf("Expected %f, got %f (tolerance: %f)", expected, actual, tolerance)
		if len(msgAndArgs) > 0 {
			t.Errorf(fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...))
		}
	}
}
