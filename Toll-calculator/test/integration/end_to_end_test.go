package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/0x0Glitch/toll-calculator/test/fixtures"
	"github.com/0x0Glitch/toll-calculator/test/helpers"
	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// EndToEndTestSuite tests the complete toll calculator system
type EndToEndTestSuite struct {
	suite.Suite
	aggregatorURL string
	gatewayURL    string
}

// SetupSuite runs once before all tests in the suite
func (suite *EndToEndTestSuite) SetupSuite() {
	// These would typically be set from environment variables or test configuration
	suite.aggregatorURL = "http://localhost:3000" // Default aggregator port
	suite.gatewayURL = "http://localhost:30000"   // Default gateway port
}

// TestCompleteFlow tests the complete flow from OBU data to invoice generation
func (suite *EndToEndTestSuite) TestCompleteFlow() {
	// Skip if services are not running
	if !suite.isServiceRunning(suite.aggregatorURL) {
		suite.T().Skip("Aggregator service not running, skipping integration test")
	}

	// Arrange - Prepare test data
	obuData := []types.OBUData{
		helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7128, -74.0060), // New York
		helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7589, -73.9851), // Times Square
		helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7831, -73.9712), // Central Park
	}

	// Act - Send OBU data through the system
	for _, data := range obuData {
		err := suite.sendOBUData(data)
		assert.NoError(suite.T(), err, "Failed to send OBU data")

		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for all data to be processed
	time.Sleep(2 * time.Second)

	// Assert - Get invoice and verify
	invoice, err := suite.getInvoice(fixtures.TestOBUID1)
	assert.NoError(suite.T(), err, "Failed to get invoice")
	assert.NotNil(suite.T(), invoice, "Invoice should not be nil")
	assert.Equal(suite.T(), fixtures.TestOBUID1, invoice.OBUID)
	assert.Greater(suite.T(), invoice.TotalDistance, 0.0, "Total distance should be greater than 0")
	assert.Greater(suite.T(), invoice.Amount, 0.0, "Amount should be greater than 0")

	// Verify the amount calculation
	expectedAmount := invoice.TotalDistance * 315 // basePrice = 315
	helpers.AssertFloatEquals(suite.T(), expectedAmount, invoice.Amount, 0.001)
}

// TestMultipleOBUs tests the system with multiple OBUs
func (suite *EndToEndTestSuite) TestMultipleOBUs() {
	if !suite.isServiceRunning(suite.aggregatorURL) {
		suite.T().Skip("Aggregator service not running, skipping integration test")
	}

	// Arrange - Prepare test data for multiple OBUs
	obuDataSets := map[int32][]types.OBUData{
		fixtures.TestOBUID1: {
			helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7128, -74.0060),
			helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7589, -73.9851),
		},
		fixtures.TestOBUID2: {
			helpers.GenerateTestOBUData(fixtures.TestOBUID2, 34.0522, -118.2437),
			helpers.GenerateTestOBUData(fixtures.TestOBUID2, 34.0928, -118.3287),
		},
		fixtures.TestOBUID3: {
			helpers.GenerateTestOBUData(fixtures.TestOBUID3, 41.8781, -87.6298),
			helpers.GenerateTestOBUData(fixtures.TestOBUID3, 41.8819, -87.6278),
		},
	}

	// Act - Send data for all OBUs
	for obuID, dataSet := range obuDataSets {
		for _, data := range dataSet {
			err := suite.sendOBUData(data)
			assert.NoError(suite.T(), err, "Failed to send OBU data for OBU %d", obuID)
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Assert - Verify invoices for all OBUs
	for obuID := range obuDataSets {
		invoice, err := suite.getInvoice(obuID)
		assert.NoError(suite.T(), err, "Failed to get invoice for OBU %d", obuID)
		assert.NotNil(suite.T(), invoice, "Invoice should not be nil for OBU %d", obuID)
		assert.Equal(suite.T(), obuID, invoice.OBUID)
		assert.Greater(suite.T(), invoice.TotalDistance, 0.0, "Total distance should be greater than 0 for OBU %d", obuID)
		assert.Greater(suite.T(), invoice.Amount, 0.0, "Amount should be greater than 0 for OBU %d", obuID)
	}
}

// TestHighVolumeData tests the system with high volume of data
func (suite *EndToEndTestSuite) TestHighVolumeData() {
	if !suite.isServiceRunning(suite.aggregatorURL) {
		suite.T().Skip("Aggregator service not running, skipping integration test")
	}

	// Arrange - Generate high volume test data
	const numDataPoints = 100
	obuID := fixtures.TestOBUID1

	// Generate a route with many points
	baseLatitude := 40.7128
	baseLongitude := -74.0060

	var obuDataPoints []types.OBUData
	for i := 0; i < numDataPoints; i++ {
		// Create a path by slightly incrementing coordinates
		lat := baseLatitude + float64(i)*0.001
		long := baseLongitude + float64(i)*0.001
		obuDataPoints = append(obuDataPoints, helpers.GenerateTestOBUData(obuID, lat, long))
	}

	// Act - Send all data points
	start := time.Now()
	for _, data := range obuDataPoints {
		err := suite.sendOBUData(data)
		assert.NoError(suite.T(), err, "Failed to send OBU data point")
		time.Sleep(10 * time.Millisecond) // Small delay to avoid overwhelming the system
	}
	sendDuration := time.Since(start)

	// Wait for processing
	time.Sleep(5 * time.Second)

	// Assert - Verify the final invoice
	invoice, err := suite.getInvoice(obuID)
	assert.NoError(suite.T(), err, "Failed to get invoice after high volume test")
	assert.NotNil(suite.T(), invoice, "Invoice should not be nil")
	assert.Equal(suite.T(), obuID, invoice.OBUID)
	assert.Greater(suite.T(), invoice.TotalDistance, 0.0, "Total distance should be greater than 0")

	suite.T().Logf("Processed %d data points in %v", numDataPoints, sendDuration)
	suite.T().Logf("Final invoice: OBU=%d, Distance=%.2f, Amount=%.2f",
		invoice.OBUID, invoice.TotalDistance, invoice.Amount)
}

// TestErrorHandling tests error handling scenarios
func (suite *EndToEndTestSuite) TestErrorHandling() {
	if !suite.isServiceRunning(suite.aggregatorURL) {
		suite.T().Skip("Aggregator service not running, skipping integration test")
	}

	// Test getting invoice for non-existent OBU
	nonExistentOBUID := int32(999999)
	invoice, err := suite.getInvoice(nonExistentOBUID)
	assert.Error(suite.T(), err, "Should get error for non-existent OBU")
	assert.Nil(suite.T(), invoice, "Invoice should be nil for non-existent OBU")

	// Test invalid OBU ID in request
	resp, err := http.Get(fmt.Sprintf("%s/invoice?obu=invalid", suite.aggregatorURL))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// Test missing OBU ID parameter
	resp, err = http.Get(fmt.Sprintf("%s/invoice", suite.aggregatorURL))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

// TestConcurrentRequests tests the system under concurrent load
func (suite *EndToEndTestSuite) TestConcurrentRequests() {
	if !suite.isServiceRunning(suite.aggregatorURL) {
		suite.T().Skip("Aggregator service not running, skipping integration test")
	}

	const numGoroutines = 10
	const dataPointsPerGoroutine = 20

	// Channel to collect results
	results := make(chan error, numGoroutines)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			obuID := int32(fixtures.TestOBUID1 + goroutineID)

			// Send data points for this OBU
			for j := 0; j < dataPointsPerGoroutine; j++ {
				lat := 40.0 + float64(goroutineID)*0.1 + float64(j)*0.001
				long := -74.0 + float64(goroutineID)*0.1 + float64(j)*0.001
				data := helpers.GenerateTestOBUData(obuID, lat, long)

				if err := suite.sendOBUData(data); err != nil {
					results <- fmt.Errorf("goroutine %d failed to send data: %v", goroutineID, err)
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
			results <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(suite.T(), err, "Concurrent request failed")
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify invoices for all OBUs
	for i := 0; i < numGoroutines; i++ {
		obuID := int32(fixtures.TestOBUID1 + i)
		invoice, err := suite.getInvoice(obuID)
		assert.NoError(suite.T(), err, "Failed to get invoice for concurrent OBU %d", obuID)
		if invoice != nil {
			assert.Equal(suite.T(), obuID, invoice.OBUID)
			assert.Greater(suite.T(), invoice.TotalDistance, 0.0, "Distance should be > 0 for OBU %d", obuID)
		}
	}
}

// Helper methods

// isServiceRunning checks if a service is running at the given URL
func (suite *EndToEndTestSuite) isServiceRunning(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// sendOBUData sends OBU data to the system (simulating the data receiver)
func (suite *EndToEndTestSuite) sendOBUData(data types.OBUData) error {
	// In a real integration test, this would send data through the WebSocket
	// or directly to the data receiver service. For now, we'll simulate by
	// directly calling the distance calculator and aggregator services.

	// Calculate distance (simulating distance calculator service)
	distance := suite.calculateDistance(data)

	// Send to aggregator
	return suite.sendDistanceToAggregator(distance)
}

// calculateDistance simulates the distance calculation service
func (suite *EndToEndTestSuite) calculateDistance(data types.OBUData) types.Distance {
	// Simple distance calculation for testing
	// In reality, this would be done by the distance calculator service
	distance := 1.0 + (float64(data.OBUID%100) * 0.1) // Simulate some distance

	return types.Distance{
		OBUID:  data.OBUID,
		Values: distance,
		Unix:   time.Now().UnixNano(),
	}
}

// sendDistanceToAggregator sends distance data to the aggregator service
func (suite *EndToEndTestSuite) sendDistanceToAggregator(distance types.Distance) error {
	jsonData, err := json.Marshal(distance)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/aggregate", suite.aggregatorURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("aggregator returned status %d", resp.StatusCode)
	}

	return nil
}

// getInvoice retrieves an invoice from the aggregator service
func (suite *EndToEndTestSuite) getInvoice(obuID int32) (*types.Invoice, error) {
	resp, err := http.Get(fmt.Sprintf("%s/invoice?obu=%d", suite.aggregatorURL, obuID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get invoice, status: %d", resp.StatusCode)
	}

	var invoice types.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		return nil, err
	}

	return &invoice, nil
}

// Run the test suite
func TestEndToEndTestSuite(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}

// Performance test
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// This test would measure system performance under load
	// It's separate from the main suite to allow selective running

	suite := &EndToEndTestSuite{
		aggregatorURL: "http://localhost:3000",
		gatewayURL:    "http://localhost:30000",
	}

	if !suite.isServiceRunning(suite.aggregatorURL) {
		t.Skip("Aggregator service not running, skipping performance test")
	}

	const numRequests = 1000
	const concurrency = 50

	start := time.Now()

	// Channel to control concurrency
	semaphore := make(chan struct{}, concurrency)
	results := make(chan error, numRequests)

	// Launch requests
	for i := 0; i < numRequests; i++ {
		go func(requestID int) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			obuID := int32(fixtures.TestOBUID1 + (requestID % 10))
			data := helpers.GenerateTestOBUData(obuID, 40.0+float64(requestID)*0.001, -74.0+float64(requestID)*0.001)

			err := suite.sendOBUData(data)
			results <- err
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	duration := time.Since(start)

	t.Logf("Performance test completed:")
	t.Logf("  Requests: %d", numRequests)
	t.Logf("  Concurrency: %d", concurrency)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Requests/sec: %.2f", float64(numRequests)/duration.Seconds())
	t.Logf("  Errors: %d", len(errors))

	// Assert performance criteria
	assert.Less(t, len(errors), numRequests/10, "Error rate should be less than 10%")
	assert.Less(t, duration, 30*time.Second, "Should complete within 30 seconds")
}
