package mocks

import (
	"context"
	"sync"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/stretchr/testify/mock"
)

// MockStorer is a mock implementation of the Storer interface
type MockStorer struct {
	mock.Mock
	mu   sync.RWMutex
	data map[int32]float64
}

func NewMockStorer() *MockStorer {
	return &MockStorer{
		data: make(map[int32]float64),
	}
}

func (m *MockStorer) Insert(d *types.Distance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(d)
	if args.Get(0) != nil {
		return args.Error(0)
	}

	// Actually store the data for realistic testing
	m.data[d.OBUID] += d.Values
	return nil
}

func (m *MockStorer) Get(id int32) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	args := m.Called(id)
	if args.Get(1) != nil {
		return 0, args.Error(1)
	}

	// Return actual stored data if available
	if val, exists := m.data[id]; exists {
		return val, nil
	}

	return args.Get(0).(float64), args.Error(1)
}

func (m *MockStorer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[int32]float64)
}

func (m *MockStorer) GetStoredData() map[int32]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[int32]float64)
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

// MockAggregator is a mock implementation of the Aggregator interface
type MockAggregator struct {
	mock.Mock
}

func (m *MockAggregator) AggregateDistance(distance *types.Distance) error {
	args := m.Called(distance)
	return args.Error(0)
}

func (m *MockAggregator) CalculateInvoice(obuID int32) (*types.Invoice, error) {
	args := m.Called(obuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Invoice), args.Error(1)
}

// MockCalculatorService is a mock implementation of the CalculatorServicer interface
type MockCalculatorService struct {
	mock.Mock
	prevPoints map[int32][]float64 // Track previous points per OBU
	mu         sync.RWMutex
}

func NewMockCalculatorService() *MockCalculatorService {
	return &MockCalculatorService{
		prevPoints: make(map[int32][]float64),
	}
}

func (m *MockCalculatorService) CalculateDistance(data types.OBUData) (float64, error) {
	args := m.Called(data)
	if args.Get(1) != nil {
		return 0, args.Error(1)
	}

	// Store the point for realistic behavior
	m.mu.Lock()
	m.prevPoints[data.OBUID] = []float64{data.Lat, data.Long}
	m.mu.Unlock()

	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCalculatorService) GetPreviousPoint(obuID int32) []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.prevPoints[obuID]
}

func (m *MockCalculatorService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prevPoints = make(map[int32][]float64)
}

// MockDataProducer is a mock implementation of the DataProducer interface
type MockDataProducer struct {
	mock.Mock
	ProducedData []types.OBUData
	mu           sync.Mutex
}

func NewMockDataProducer() *MockDataProducer {
	return &MockDataProducer{
		ProducedData: make([]types.OBUData, 0),
	}
}

func (m *MockDataProducer) ProduceData(data types.OBUData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(data)
	if args.Get(0) == nil {
		m.ProducedData = append(m.ProducedData, data)
	}
	return args.Error(0)
}

func (m *MockDataProducer) GetProducedData() []types.OBUData {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]types.OBUData, len(m.ProducedData))
	copy(result, m.ProducedData)
	return result
}

func (m *MockDataProducer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ProducedData = make([]types.OBUData, 0)
}

// MockAggregatorClient is a mock implementation of the aggregator client
type MockAggregatorClient struct {
	mock.Mock
	AggregatedRequests []types.AggregatorRequest
	mu                 sync.Mutex
}

func NewMockAggregatorClient() *MockAggregatorClient {
	return &MockAggregatorClient{
		AggregatedRequests: make([]types.AggregatorRequest, 0),
	}
}

func (m *MockAggregatorClient) Aggregate(ctx context.Context, req *types.AggregatorRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		m.AggregatedRequests = append(m.AggregatedRequests, *req)
	}
	return args.Error(0)
}

func (m *MockAggregatorClient) GetInvoice(ctx context.Context, obuID int) (*types.Invoice, error) {
	args := m.Called(ctx, obuID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Invoice), args.Error(1)
}

func (m *MockAggregatorClient) GetAggregatedRequests() []types.AggregatorRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]types.AggregatorRequest, len(m.AggregatedRequests))
	copy(result, m.AggregatedRequests)
	return result
}

func (m *MockAggregatorClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AggregatedRequests = make([]types.AggregatorRequest, 0)
}
