package fixtures

import (
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
)

// Sample OBU IDs for testing
var (
	TestOBUID1 int32 = 12345
	TestOBUID2 int32 = 67890
	TestOBUID3 int32 = 11111
)

// Sample coordinates for testing
var (
	TestCoord1 = struct{ Lat, Long float64 }{40.7128, -74.0060}  // New York
	TestCoord2 = struct{ Lat, Long float64 }{34.0522, -118.2437} // Los Angeles
	TestCoord3 = struct{ Lat, Long float64 }{41.8781, -87.6298}  // Chicago
)

// GetSampleOBUData returns sample OBU data for testing
func GetSampleOBUData() []types.OBUData {
	return []types.OBUData{
		{
			OBUID:     TestOBUID1,
			Lat:       TestCoord1.Lat,
			Long:      TestCoord1.Long,
			RequestID: 1001,
		},
		{
			OBUID:     TestOBUID1,
			Lat:       TestCoord2.Lat,
			Long:      TestCoord2.Long,
			RequestID: 1002,
		},
		{
			OBUID:     TestOBUID2,
			Lat:       TestCoord1.Lat,
			Long:      TestCoord1.Long,
			RequestID: 2001,
		},
		{
			OBUID:     TestOBUID2,
			Lat:       TestCoord3.Lat,
			Long:      TestCoord3.Long,
			RequestID: 2002,
		},
	}
}

// GetSampleDistances returns sample distance data for testing
func GetSampleDistances() []types.Distance {
	now := time.Now().UnixNano()
	return []types.Distance{
		{
			OBUID:  TestOBUID1,
			Values: 10.5,
			Unix:   now,
		},
		{
			OBUID:  TestOBUID1,
			Values: 15.2,
			Unix:   now + 1000000,
		},
		{
			OBUID:  TestOBUID2,
			Values: 8.7,
			Unix:   now + 2000000,
		},
		{
			OBUID:  TestOBUID2,
			Values: 12.3,
			Unix:   now + 3000000,
		},
	}
}

// GetSampleInvoices returns sample invoice data for testing
func GetSampleInvoices() []types.Invoice {
	return []types.Invoice{
		{
			OBUID:         TestOBUID1,
			TotalDistance: 25.7,
			Amount:        25.7 * 315, // basePrice = 315
		},
		{
			OBUID:         TestOBUID2,
			TotalDistance: 21.0,
			Amount:        21.0 * 315,
		},
	}
}

// GetSampleAggregatorRequests returns sample aggregator requests for testing
func GetSampleAggregatorRequests() []types.AggregatorRequest {
	now := time.Now().UnixNano()
	return []types.AggregatorRequest{
		{
			ObuID: TestOBUID1,
			Value: 10.5,
			Unix:  now,
		},
		{
			ObuID: TestOBUID1,
			Value: 15.2,
			Unix:  now + 1000000,
		},
		{
			ObuID: TestOBUID2,
			Value: 8.7,
			Unix:  now + 2000000,
		},
	}
}

// GetInvalidOBUData returns invalid OBU data for negative testing
func GetInvalidOBUData() []types.OBUData {
	return []types.OBUData{
		{
			OBUID:     -1, // Invalid negative ID
			Lat:       TestCoord1.Lat,
			Long:      TestCoord1.Long,
			RequestID: 9001,
		},
		{
			OBUID:     TestOBUID1,
			Lat:       200.0, // Invalid latitude (out of range)
			Long:      TestCoord1.Long,
			RequestID: 9002,
		},
		{
			OBUID:     TestOBUID1,
			Lat:       TestCoord1.Lat,
			Long:      200.0, // Invalid longitude (out of range)
			RequestID: 9003,
		},
	}
}

// GetTestCoordinatePairs returns coordinate pairs for distance calculation testing
func GetTestCoordinatePairs() []struct {
	From, To         struct{ Lat, Long float64 }
	ExpectedDistance float64
} {
	return []struct {
		From, To         struct{ Lat, Long float64 }
		ExpectedDistance float64
	}{
		{
			From:             TestCoord1,
			To:               TestCoord1,
			ExpectedDistance: 0.0, // Same point
		},
		{
			From:             struct{ Lat, Long float64 }{0.0, 0.0},
			To:               struct{ Lat, Long float64 }{3.0, 4.0},
			ExpectedDistance: 5.0, // 3-4-5 triangle
		},
		{
			From:             struct{ Lat, Long float64 }{1.0, 1.0},
			To:               struct{ Lat, Long float64 }{4.0, 5.0},
			ExpectedDistance: 5.0, // Another 3-4-5 triangle
		},
	}
}
