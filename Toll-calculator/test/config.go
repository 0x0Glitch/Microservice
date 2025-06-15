package test

import (
	"os"
	"strconv"
	"time"
)

// TestConfig holds configuration for tests
type TestConfig struct {
	AggregatorURL    string
	GatewayURL       string
	DataReceiverURL  string
	TestTimeout      time.Duration
	IntegrationTests bool
	ShortTests       bool
	Verbose          bool
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		AggregatorURL:    "http://localhost:3000",
		GatewayURL:       "http://localhost:30000",
		DataReceiverURL:  "http://localhost:30000",
		TestTimeout:      5 * time.Minute,
		IntegrationTests: true,
		ShortTests:       false,
		Verbose:          false,
	}
}

// LoadTestConfig loads test configuration from environment variables
func LoadTestConfig() *TestConfig {
	config := DefaultTestConfig()

	if url := os.Getenv("AGGREGATOR_URL"); url != "" {
		config.AggregatorURL = url
	}

	if url := os.Getenv("GATEWAY_URL"); url != "" {
		config.GatewayURL = url
	}

	if url := os.Getenv("DATA_RECEIVER_URL"); url != "" {
		config.DataReceiverURL = url
	}

	if timeout := os.Getenv("TEST_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.TestTimeout = d
		}
	}

	if integration := os.Getenv("INTEGRATION_TESTS"); integration != "" {
		if b, err := strconv.ParseBool(integration); err == nil {
			config.IntegrationTests = b
		}
	}

	if short := os.Getenv("SHORT_TESTS"); short != "" {
		if b, err := strconv.ParseBool(short); err == nil {
			config.ShortTests = b
		}
	}

	if verbose := os.Getenv("VERBOSE_TESTS"); verbose != "" {
		if b, err := strconv.ParseBool(verbose); err == nil {
			config.Verbose = b
		}
	}

	return config
}

// TestConstants holds common test constants
type TestConstants struct {
	BasePrice        float64
	DefaultOBUID     int32
	DefaultLatitude  float64
	DefaultLongitude float64
	FloatTolerance   float64
	DefaultTimeout   time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
}

// GetTestConstants returns common test constants
func GetTestConstants() *TestConstants {
	return &TestConstants{
		BasePrice:        315.0,
		DefaultOBUID:     12345,
		DefaultLatitude:  40.7128,
		DefaultLongitude: -74.0060,
		FloatTolerance:   0.001,
		DefaultTimeout:   30 * time.Second,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}
}

// TestEnvironment represents the test environment setup
type TestEnvironment struct {
	Config    *TestConfig
	Constants *TestConstants
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment() *TestEnvironment {
	return &TestEnvironment{
		Config:    LoadTestConfig(),
		Constants: GetTestConstants(),
	}
}

// IsIntegrationTest returns true if integration tests should run
func (te *TestEnvironment) IsIntegrationTest() bool {
	return te.Config.IntegrationTests
}

// IsShortTest returns true if running in short test mode
func (te *TestEnvironment) IsShortTest() bool {
	return te.Config.ShortTests
}

// GetAggregatorURL returns the aggregator service URL
func (te *TestEnvironment) GetAggregatorURL() string {
	return te.Config.AggregatorURL
}

// GetGatewayURL returns the gateway service URL
func (te *TestEnvironment) GetGatewayURL() string {
	return te.Config.GatewayURL
}

// GetDataReceiverURL returns the data receiver service URL
func (te *TestEnvironment) GetDataReceiverURL() string {
	return te.Config.DataReceiverURL
}

// GetTestTimeout returns the test timeout duration
func (te *TestEnvironment) GetTestTimeout() time.Duration {
	return te.Config.TestTimeout
}

// GetBasePrice returns the base price for toll calculation
func (te *TestEnvironment) GetBasePrice() float64 {
	return te.Constants.BasePrice
}

// GetFloatTolerance returns the tolerance for float comparisons
func (te *TestEnvironment) GetFloatTolerance() float64 {
	return te.Constants.FloatTolerance
}

// GetDefaultTimeout returns the default timeout for operations
func (te *TestEnvironment) GetDefaultTimeout() time.Duration {
	return te.Constants.DefaultTimeout
}

// GetMaxRetries returns the maximum number of retries for operations
func (te *TestEnvironment) GetMaxRetries() int {
	return te.Constants.MaxRetries
}

// GetRetryDelay returns the delay between retries
func (te *TestEnvironment) GetRetryDelay() time.Duration {
	return te.Constants.RetryDelay
}
