# Toll Calculator Test Suite

This directory contains comprehensive unit and integration tests for the Toll Calculator microservice system. The test suite is designed following Go best practices and provides extensive coverage of all system components.

## 📁 Directory Structure

```
test/
├── unit/                    # Unit tests for individual components
│   ├── distance_calculator_test.go
│   ├── aggregator_service_test.go
│   └── http_handlers_test.go
├── integration/             # End-to-end integration tests
│   └── end_to_end_test.go
├── mocks/                   # Mock implementations for testing
│   └── mocks.go
├── fixtures/                # Test data and fixtures
│   └── test_data.go
├── helpers/                 # Test utilities and helpers
│   └── test_helpers.go
├── Makefile                 # Test automation commands
└── README.md               # This file
```

## 🧪 Test Categories

### Unit Tests

- **Distance Calculator Tests**: Test the distance calculation logic with various coordinate inputs
- **Aggregator Service Tests**: Test distance aggregation and invoice calculation
- **HTTP Handler Tests**: Test REST API endpoints and error handling
- **Memory Store Tests**: Test in-memory data storage functionality

### Integration Tests

- **End-to-End Tests**: Test complete system flow from OBU data to invoice generation
- **Multi-OBU Tests**: Test system behavior with multiple concurrent OBUs
- **High Volume Tests**: Test system performance under load
- **Error Handling Tests**: Test system resilience and error scenarios
- **Concurrent Request Tests**: Test system behavior under concurrent load

## 🚀 Running Tests

### Prerequisites

```bash
# Install test dependencies
go mod tidy
```

### Quick Start

```bash
# Run all unit tests
make test

# Run with verbose output
make test-unit

# Run integration tests (requires services running)
make test-integration

# Run all tests
make test-all
```

### Available Commands

| Command                 | Description                        |
| ----------------------- | ---------------------------------- |
| `make test`             | Run unit tests only                |
| `make test-unit`        | Run unit tests with verbose output |
| `make test-integration` | Run integration tests              |
| `make test-all`         | Run all tests (unit + integration) |
| `make test-coverage`    | Generate coverage report           |
| `make test-benchmark`   | Run benchmark tests                |
| `make test-race`        | Run tests with race detection      |
| `make test-short`       | Skip long-running tests            |
| `make clean`            | Clean test cache and artifacts     |

### Advanced Testing

```bash
# Run specific test
make test-specific TEST=TestCalculateDistance

# Run with coverage
make test-coverage

# Run benchmarks
make test-benchmark

# Run with race detection
make test-race

# Performance testing
make test-performance
```

## 📊 Test Coverage

The test suite aims for high coverage across all components:

- **Distance Calculator**: 100% function coverage
- **Aggregator Service**: 95%+ coverage including error paths
- **HTTP Handlers**: 100% endpoint coverage with error scenarios
- **Integration**: Complete end-to-end flow coverage

### Generating Coverage Reports

```bash
make test-coverage
# Opens coverage.html in your browser
```

## 🏗️ Test Architecture

### Test Patterns Used

1. **Table-Driven Tests**: For testing multiple input scenarios
2. **Test Suites**: Using testify/suite for organized test structure
3. **Mocks and Stubs**: For isolating components under test
4. **Test Fixtures**: Reusable test data and scenarios
5. **Helper Functions**: Common test utilities and assertions

### Example Test Structure

```go
func (suite *ServiceTestSuite) TestFeature_Scenario() {
    // Arrange
    input := fixtures.GetTestData()

    // Act
    result, err := suite.service.Method(input)

    // Assert
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), expected, result)
}
```

## 🔧 Test Configuration

### Environment Variables

```bash
# Integration test configuration
export AGGREGATOR_URL="http://localhost:3000"
export GATEWAY_URL="http://localhost:30000"
export TEST_TIMEOUT="5m"
```

### Test Flags

```bash
# Skip integration tests
go test -short ./test/...

# Run with race detection
go test -race ./test/...

# Run with coverage
go test -cover ./test/...

# Run specific test pattern
go test -run TestDistance ./test/...
```

## 🐛 Debugging Tests

### Verbose Output

```bash
go test -v ./test/unit/...
```

### Debug Specific Test

```bash
go test -v -run TestSpecificFunction ./test/unit/
```

### Memory and CPU Profiling

```bash
# Memory profiling
make test-memory

# CPU profiling
make test-cpu
```

## 🚀 Integration Test Setup

Integration tests require the following services to be running:

1. **Aggregator Service** (port 3000)
2. **Gateway Service** (port 30000)
3. **Kafka** (for message queue tests)

### Starting Services for Integration Tests

```bash
# Start all services
docker-compose up -d

# Or start individual services
cd aggregator && go run .
cd gateway && go run .
```

### Verifying Service Health

```bash
curl http://localhost:3000/health
curl http://localhost:30000/health
```

## 📈 Performance Testing

### Benchmark Tests

```bash
# Run all benchmarks
make test-benchmark

# Run specific benchmark
go test -bench=BenchmarkCalculateDistance ./test/unit/
```

### Load Testing

```bash
# High volume integration test
go test -run TestHighVolumeData ./test/integration/

# Concurrent request test
go test -run TestConcurrentRequests ./test/integration/
```

## 🔍 Test Data and Fixtures

### Sample Data

- **Test OBU IDs**: 12345, 67890, 11111
- **Test Coordinates**: New York, Los Angeles, Chicago
- **Test Distances**: Various distance calculations
- **Test Invoices**: Sample invoice data with different amounts

### Custom Test Data

```go
// Create custom test data
obuData := helpers.GenerateTestOBUData(12345, 40.7128, -74.0060)
distance := helpers.GenerateTestDistance(12345, 10.5)
invoice := helpers.GenerateTestInvoice(12345, 25.5, 8032.5)
```

## 🛠️ Extending Tests

### Adding New Unit Tests

1. Create test file in `test/unit/`
2. Follow naming convention: `*_test.go`
3. Use testify/suite for organization
4. Add test data to fixtures if needed

### Adding Integration Tests

1. Add test methods to `end_to_end_test.go`
2. Ensure proper service dependencies
3. Use realistic test scenarios
4. Include error handling tests

### Creating Mocks

1. Add mock implementations to `test/mocks/`
2. Implement all interface methods
3. Add helper methods for test setup
4. Include thread-safe operations if needed

## 📋 Best Practices

### Test Organization

- One test file per source file
- Group related tests in suites
- Use descriptive test names
- Follow Arrange-Act-Assert pattern

### Test Data

- Use fixtures for reusable data
- Generate dynamic test data when needed
- Clean up test data after tests
- Use realistic but simple test scenarios

### Assertions

- Use specific assertions (Equal vs True)
- Include helpful error messages
- Test both success and failure paths
- Verify all important properties

### Performance

- Use benchmarks for performance-critical code
- Set appropriate test timeouts
- Clean up resources in tests
- Use parallel tests when safe

## 🚨 Troubleshooting

### Common Issues

1. **Import Path Errors**: Ensure you're running tests from the correct directory
2. **Service Not Running**: Integration tests require services to be running
3. **Race Conditions**: Use proper synchronization in concurrent tests
4. **Timeout Issues**: Increase timeout for slow tests

### Debug Commands

```bash
# Check test compilation
go test -c ./test/unit/

# Run with detailed output
go test -v -x ./test/unit/

# Check for race conditions
go test -race ./test/unit/
```

## 📚 Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Test Best Practices](https://golang.org/doc/tutorial/add-a-test)
- [Testing Microservices](https://microservices.io/patterns/testing/)

## 🤝 Contributing

When adding new tests:

1. Follow existing patterns and conventions
2. Add appropriate documentation
3. Ensure tests are deterministic
4. Include both positive and negative test cases
5. Update this README if adding new test categories
