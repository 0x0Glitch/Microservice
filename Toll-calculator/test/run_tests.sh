#!/bin/bash

# Toll Calculator Test Runner
# This script provides a comprehensive way to run tests for the toll calculator system

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
TEST_TYPE="unit"
COVERAGE=false
VERBOSE=false
RACE=false
BENCHMARK=false
INTEGRATION=false
CLEANUP=false
TIMEOUT="5m"
PARALLEL=4

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Toll Calculator Test Runner

Usage: $0 [OPTIONS]

OPTIONS:
    -t, --type TYPE         Test type: unit, integration, all (default: unit)
    -c, --coverage          Generate coverage report
    -v, --verbose           Verbose output
    -r, --race              Enable race detection
    -b, --benchmark         Run benchmark tests
    -i, --integration       Run integration tests (requires services)
    -p, --parallel N        Number of parallel test processes (default: 4)
    --timeout DURATION      Test timeout (default: 5m)
    --cleanup               Clean test cache and artifacts
    --help                  Show this help message

EXAMPLES:
    $0                      # Run unit tests
    $0 -t all -c -v         # Run all tests with coverage and verbose output
    $0 -i                   # Run integration tests only
    $0 -b                   # Run benchmark tests
    $0 --cleanup            # Clean test artifacts

ENVIRONMENT VARIABLES:
    AGGREGATOR_URL          Aggregator service URL (default: http://localhost:3000)
    GATEWAY_URL             Gateway service URL (default: http://localhost:30000)
    TEST_TIMEOUT            Test timeout duration (default: 5m)
    INTEGRATION_TESTS       Enable/disable integration tests (default: true)
    VERBOSE_TESTS           Enable verbose test output (default: false)

EOF
}

# Function to check if services are running
check_services() {
    local aggregator_url="${AGGREGATOR_URL:-http://localhost:3000}"
    local gateway_url="${GATEWAY_URL:-http://localhost:30000}"
    
    print_status "Checking if services are running..."
    
    if curl -s -f "${aggregator_url}/health" > /dev/null 2>&1; then
        print_success "Aggregator service is running at ${aggregator_url}"
    else
        print_warning "Aggregator service is not running at ${aggregator_url}"
        return 1
    fi
    
    if curl -s -f "${gateway_url}/health" > /dev/null 2>&1; then
        print_success "Gateway service is running at ${gateway_url}"
    else
        print_warning "Gateway service is not running at ${gateway_url}"
        return 1
    fi
    
    return 0
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    local cmd="go test"
    local flags=""
    
    if [ "$VERBOSE" = true ]; then
        flags="$flags -v"
    fi
    
    if [ "$RACE" = true ]; then
        flags="$flags -race"
    fi
    
    if [ "$COVERAGE" = true ]; then
        flags="$flags -coverprofile=coverage.out -covermode=atomic"
    fi
    
    flags="$flags -timeout=$TIMEOUT -parallel=$PARALLEL"
    
    # Change to parent directory to run tests
    cd "$(dirname "$0")/.."
    
    if eval "$cmd $flags ./test/unit/..."; then
        print_success "Unit tests passed!"
        
        if [ "$COVERAGE" = true ]; then
            print_status "Generating coverage report..."
            go tool cover -html=coverage.out -o coverage.html
            print_success "Coverage report generated: coverage.html"
            
            # Show coverage summary
            go tool cover -func=coverage.out | tail -1
        fi
    else
        print_error "Unit tests failed!"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Check if services are running
    if ! check_services; then
        print_error "Required services are not running. Please start them first."
        print_status "To start services:"
        print_status "  cd aggregator && go run . &"
        print_status "  cd gateway && go run . &"
        return 1
    fi
    
    local cmd="go test"
    local flags=""
    
    if [ "$VERBOSE" = true ]; then
        flags="$flags -v"
    fi
    
    flags="$flags -timeout=$TIMEOUT -parallel=$PARALLEL"
    
    # Change to parent directory to run tests
    cd "$(dirname "$0")/.."
    
    if eval "$cmd $flags ./test/integration/..."; then
        print_success "Integration tests passed!"
    else
        print_error "Integration tests failed!"
        return 1
    fi
}

# Function to run benchmark tests
run_benchmark_tests() {
    print_status "Running benchmark tests..."
    
    local cmd="go test"
    local flags="-bench=. -benchmem"
    
    if [ "$VERBOSE" = true ]; then
        flags="$flags -v"
    fi
    
    flags="$flags -timeout=$TIMEOUT"
    
    # Change to parent directory to run tests
    cd "$(dirname "$0")/.."
    
    if eval "$cmd $flags ./test/unit/..."; then
        print_success "Benchmark tests completed!"
    else
        print_error "Benchmark tests failed!"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    print_status "Running all tests..."
    
    # Run unit tests first
    if ! run_unit_tests; then
        return 1
    fi
    
    # Run integration tests if services are available
    if check_services; then
        if ! run_integration_tests; then
            return 1
        fi
    else
        print_warning "Skipping integration tests - services not running"
    fi
    
    print_success "All tests completed successfully!"
}

# Function to clean test artifacts
cleanup_tests() {
    print_status "Cleaning test artifacts..."
    
    cd "$(dirname "$0")/.."
    
    # Clean Go test cache
    go clean -testcache
    
    # Remove coverage files
    rm -f coverage.out coverage.html
    
    # Remove profile files
    rm -f *.prof
    
    print_success "Test artifacts cleaned!"
}

# Function to setup test environment
setup_environment() {
    print_status "Setting up test environment..."
    
    cd "$(dirname "$0")/.."
    
    # Ensure dependencies are up to date
    go mod tidy
    go mod download
    
    print_success "Test environment setup complete!"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -r|--race)
            RACE=true
            shift
            ;;
        -b|--benchmark)
            BENCHMARK=true
            shift
            ;;
        -i|--integration)
            INTEGRATION=true
            shift
            ;;
        -p|--parallel)
            PARALLEL="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --cleanup)
            CLEANUP=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_status "Toll Calculator Test Runner"
    print_status "=========================="
    
    # Setup environment
    setup_environment
    
    # Handle cleanup
    if [ "$CLEANUP" = true ]; then
        cleanup_tests
        exit 0
    fi
    
    # Handle benchmark tests
    if [ "$BENCHMARK" = true ]; then
        run_benchmark_tests
        exit $?
    fi
    
    # Handle integration tests
    if [ "$INTEGRATION" = true ]; then
        run_integration_tests
        exit $?
    fi
    
    # Handle test type
    case $TEST_TYPE in
        unit)
            run_unit_tests
            ;;
        integration)
            run_integration_tests
            ;;
        all)
            run_all_tests
            ;;
        *)
            print_error "Invalid test type: $TEST_TYPE"
            print_error "Valid types: unit, integration, all"
            exit 1
            ;;
    esac
    
    exit $?
}

# Run main function
main 