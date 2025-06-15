package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0x0Glitch/toll-calculator/test/fixtures"
	"github.com/0x0Glitch/toll-calculator/test/helpers"
	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// HTTPFunc type for testing
type HTTPFunc func(http.ResponseWriter, *http.Request) error

// APIError for testing
type APIError struct {
	code int
	Err  error
}

func (e APIError) Error() string {
	return e.Err.Error()
}

// Mock aggregator for HTTP handler testing
type MockHTTPAggregator struct {
	distances map[int32]float64
	invoices  map[int32]*types.Invoice
}

func NewMockHTTPAggregator() *MockHTTPAggregator {
	return &MockHTTPAggregator{
		distances: make(map[int32]float64),
		invoices:  make(map[int32]*types.Invoice),
	}
}

func (m *MockHTTPAggregator) AggregateDistance(distance *types.Distance) error {
	m.distances[distance.OBUID] += distance.Values
	return nil
}

func (m *MockHTTPAggregator) CalculateInvoice(obuID int32) (*types.Invoice, error) {
	if invoice, exists := m.invoices[obuID]; exists {
		return invoice, nil
	}
	if distance, exists := m.distances[obuID]; exists {
		invoice := &types.Invoice{
			OBUID:         obuID,
			TotalDistance: distance,
			Amount:        distance * 315,
		}
		return invoice, nil
	}
	return nil, fmt.Errorf("couldn't find distance for id: %d", obuID)
}

func (m *MockHTTPAggregator) SetInvoice(obuID int32, invoice *types.Invoice) {
	m.invoices[obuID] = invoice
}

// Helper functions for HTTP testing
func makeHTTPHandlerFunc(fn HTTPFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			if apiErr, ok := err.(APIError); ok {
				writeJSON(w, apiErr.code, map[string]string{"error": apiErr.Error()})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func handleGetInvoice(svc Aggregator) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return APIError{
				code: http.StatusMethodNotAllowed,
				Err:  fmt.Errorf("invalid HTTP method %v", r.Method),
			}
		}
		values, ok := r.URL.Query()["obu"]
		if !ok {
			return APIError{
				code: http.StatusBadRequest,
				Err:  fmt.Errorf("missing OBU ID"),
			}
		}

		obuID := values[0]
		var id int32
		if _, err := fmt.Sscanf(obuID, "%d", &id); err != nil {
			return APIError{
				code: http.StatusBadRequest,
				Err:  fmt.Errorf("invalid OBU ID %v", obuID),
			}
		}

		invoice, err := svc.CalculateInvoice(id)
		if err != nil {
			return APIError{
				code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to calculate invoice for OBU ID %v", id),
			}
		}
		return writeJSON(w, http.StatusOK, invoice)
	}
}

func handleAggregate(svc Aggregator) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return APIError{
				code: http.StatusMethodNotAllowed,
				Err:  fmt.Errorf("invalid HTTP method %v", r.Method),
			}
		}
		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			return APIError{
				code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode distance: %v", err),
			}
		}
		if err := svc.AggregateDistance(&distance); err != nil {
			return APIError{
				code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to aggregate distance: %v", err),
			}
		}
		return writeJSON(w, http.StatusOK, map[string]string{"message": "distance aggregated successfully"})
	}
}

// HTTPHandlersTestSuite tests HTTP handlers
type HTTPHandlersTestSuite struct {
	suite.Suite
	aggregator *MockHTTPAggregator
	server     *httptest.Server
}

func (suite *HTTPHandlersTestSuite) SetupTest() {
	suite.aggregator = NewMockHTTPAggregator()

	mux := http.NewServeMux()
	mux.HandleFunc("/invoice", makeHTTPHandlerFunc(handleGetInvoice(suite.aggregator)))
	mux.HandleFunc("/aggregate", makeHTTPHandlerFunc(handleAggregate(suite.aggregator)))

	suite.server = httptest.NewServer(mux)
}

func (suite *HTTPHandlersTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestGetInvoice_ValidOBU tests getting invoice for valid OBU
func (suite *HTTPHandlersTestSuite) TestGetInvoice_ValidOBU() {
	// Arrange
	expectedInvoice := helpers.GenerateTestInvoice(fixtures.TestOBUID1, 25.5, 25.5*315)
	suite.aggregator.SetInvoice(fixtures.TestOBUID1, &expectedInvoice)

	// Act
	resp, err := http.Get(fmt.Sprintf("%s/invoice?obu=%d", suite.server.URL, fixtures.TestOBUID1))

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var invoice types.Invoice
	err = json.NewDecoder(resp.Body).Decode(&invoice)
	assert.NoError(suite.T(), err)
	resp.Body.Close()

	assert.Equal(suite.T(), expectedInvoice.OBUID, invoice.OBUID)
	helpers.AssertFloatEquals(suite.T(), expectedInvoice.TotalDistance, invoice.TotalDistance, 0.001)
	helpers.AssertFloatEquals(suite.T(), expectedInvoice.Amount, invoice.Amount, 0.001)
}

// TestGetInvoice_MissingOBUParam tests getting invoice without OBU parameter
func (suite *HTTPHandlersTestSuite) TestGetInvoice_MissingOBUParam() {
	// Act
	resp, err := http.Get(fmt.Sprintf("%s/invoice", suite.server.URL))

	// Assert
	assert.NoError(suite.T(), err)
	helpers.AssertErrorResponse(suite.T(), resp, http.StatusBadRequest, "missing OBU ID")
}

// TestGetInvoice_InvalidOBUParam tests getting invoice with invalid OBU parameter
func (suite *HTTPHandlersTestSuite) TestGetInvoice_InvalidOBUParam() {
	// Act
	resp, err := http.Get(fmt.Sprintf("%s/invoice?obu=invalid", suite.server.URL))

	// Assert
	assert.NoError(suite.T(), err)
	helpers.AssertErrorResponse(suite.T(), resp, http.StatusBadRequest, "invalid OBU ID")
}

// TestGetInvoice_NonExistentOBU tests getting invoice for non-existent OBU
func (suite *HTTPHandlersTestSuite) TestGetInvoice_NonExistentOBU() {
	// Act
	resp, err := http.Get(fmt.Sprintf("%s/invoice?obu=%d", suite.server.URL, fixtures.TestOBUID1))

	// Assert
	assert.NoError(suite.T(), err)
	helpers.AssertErrorResponse(suite.T(), resp, http.StatusInternalServerError, "failed to calculate invoice")
}

// TestGetInvoice_WrongHTTPMethod tests getting invoice with wrong HTTP method
func (suite *HTTPHandlersTestSuite) TestGetInvoice_WrongHTTPMethod() {
	// Act
	resp, err := http.Post(fmt.Sprintf("%s/invoice?obu=%d", suite.server.URL, fixtures.TestOBUID1), "application/json", nil)

	// Assert
	assert.NoError(suite.T(), err)
	helpers.AssertErrorResponse(suite.T(), resp, http.StatusMethodNotAllowed, "invalid HTTP method")
}

// TestAggregate_ValidDistance tests aggregating valid distance
func (suite *HTTPHandlersTestSuite) TestAggregate_ValidDistance() {
	// Arrange
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5)
	jsonData, err := json.Marshal(distance)
	assert.NoError(suite.T(), err)

	// Act
	resp, err := http.Post(fmt.Sprintf("%s/aggregate", suite.server.URL), "application/json", bytes.NewBuffer(jsonData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)
	resp.Body.Close()

	assert.Equal(suite.T(), "distance aggregated successfully", response["message"])

	// Verify the distance was aggregated
	helpers.AssertFloatEquals(suite.T(), 10.5, suite.aggregator.distances[fixtures.TestOBUID1], 0.001)
}

// TestAggregate_InvalidJSON tests aggregating with invalid JSON
func (suite *HTTPHandlersTestSuite) TestAggregate_InvalidJSON() {
	// Act
	resp, err := http.Post(fmt.Sprintf("%s/aggregate", suite.server.URL), "application/json", bytes.NewBuffer([]byte("invalid json")))

	// Assert
	assert.NoError(suite.T(), err)
	helpers.AssertErrorResponse(suite.T(), resp, http.StatusBadRequest, "failed to decode distance")
}

// TestAggregate_WrongHTTPMethod tests aggregating with wrong HTTP method
func (suite *HTTPHandlersTestSuite) TestAggregate_WrongHTTPMethod() {
	// Act
	resp, err := http.Get(fmt.Sprintf("%s/aggregate", suite.server.URL))

	// Assert
	assert.NoError(suite.T(), err)
	helpers.AssertErrorResponse(suite.T(), resp, http.StatusMethodNotAllowed, "invalid HTTP method")
}

// TestAggregate_MultipleDistances tests aggregating multiple distances
func (suite *HTTPHandlersTestSuite) TestAggregate_MultipleDistances() {
	// Arrange
	distances := []types.Distance{
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5),
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 15.2),
		helpers.GenerateTestDistance(fixtures.TestOBUID2, 8.7),
	}

	// Act & Assert
	for _, distance := range distances {
		jsonData, err := json.Marshal(distance)
		assert.NoError(suite.T(), err)

		resp, err := http.Post(fmt.Sprintf("%s/aggregate", suite.server.URL), "application/json", bytes.NewBuffer(jsonData))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Verify aggregated distances
	helpers.AssertFloatEquals(suite.T(), 25.7, suite.aggregator.distances[fixtures.TestOBUID1], 0.001) // 10.5 + 15.2
	helpers.AssertFloatEquals(suite.T(), 8.7, suite.aggregator.distances[fixtures.TestOBUID2], 0.001)
}

// TestAggregate_ZeroDistance tests aggregating zero distance
func (suite *HTTPHandlersTestSuite) TestAggregate_ZeroDistance() {
	// Arrange
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 0.0)
	jsonData, err := json.Marshal(distance)
	assert.NoError(suite.T(), err)

	// Act
	resp, err := http.Post(fmt.Sprintf("%s/aggregate", suite.server.URL), "application/json", bytes.NewBuffer(jsonData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify zero distance was aggregated
	helpers.AssertFloatEquals(suite.T(), 0.0, suite.aggregator.distances[fixtures.TestOBUID1], 0.001)
}

// TestAggregate_NegativeDistance tests aggregating negative distance
func (suite *HTTPHandlersTestSuite) TestAggregate_NegativeDistance() {
	// Arrange
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, -5.5)
	jsonData, err := json.Marshal(distance)
	assert.NoError(suite.T(), err)

	// Act
	resp, err := http.Post(fmt.Sprintf("%s/aggregate", suite.server.URL), "application/json", bytes.NewBuffer(jsonData))

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify negative distance was aggregated (business logic allows it)
	helpers.AssertFloatEquals(suite.T(), -5.5, suite.aggregator.distances[fixtures.TestOBUID1], 0.001)
}

// TestIntegration_AggregateAndGetInvoice tests the full flow of aggregating and getting invoice
func (suite *HTTPHandlersTestSuite) TestIntegration_AggregateAndGetInvoice() {
	// Arrange - Aggregate some distances
	distances := []types.Distance{
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5),
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 15.2),
	}

	for _, distance := range distances {
		jsonData, err := json.Marshal(distance)
		assert.NoError(suite.T(), err)

		resp, err := http.Post(fmt.Sprintf("%s/aggregate", suite.server.URL), "application/json", bytes.NewBuffer(jsonData))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Act - Get invoice
	resp, err := http.Get(fmt.Sprintf("%s/invoice?obu=%d", suite.server.URL, fixtures.TestOBUID1))

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var invoice types.Invoice
	err = json.NewDecoder(resp.Body).Decode(&invoice)
	assert.NoError(suite.T(), err)
	resp.Body.Close()

	assert.Equal(suite.T(), fixtures.TestOBUID1, invoice.OBUID)
	helpers.AssertFloatEquals(suite.T(), 25.7, invoice.TotalDistance, 0.001) // 10.5 + 15.2
	helpers.AssertFloatEquals(suite.T(), 25.7*315, invoice.Amount, 0.001)
}

// Run the test suite
func TestHTTPHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPHandlersTestSuite))
}

// Benchmark tests
func BenchmarkHandleGetInvoice(b *testing.B) {
	aggregator := NewMockHTTPAggregator()
	invoice := helpers.GenerateTestInvoice(fixtures.TestOBUID1, 25.5, 25.5*315)
	aggregator.SetInvoice(fixtures.TestOBUID1, &invoice)

	handler := makeHTTPHandlerFunc(handleGetInvoice(aggregator))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/invoice?obu=%d", fixtures.TestOBUID1), nil)
		w := httptest.NewRecorder()
		handler(w, req)
	}
}

func BenchmarkHandleAggregate(b *testing.B) {
	aggregator := NewMockHTTPAggregator()
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5)
	jsonData, _ := json.Marshal(distance)

	handler := makeHTTPHandlerFunc(handleAggregate(aggregator))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/aggregate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
	}
}
