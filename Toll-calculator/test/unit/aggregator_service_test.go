package unit

import (
	"fmt"
	"testing"

	"github.com/0x0Glitch/toll-calculator/test/fixtures"
	"github.com/0x0Glitch/toll-calculator/test/helpers"
	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Interfaces for testing (copied from main package)
type Aggregator interface {
	AggregateDistance(*types.Distance) error
	CalculateInvoice(int32) (*types.Invoice, error)
}

type Storer interface {
	Insert(*types.Distance) error
	Get(int32) (float64, error)
}

// InvoiceAggregator implementation for testing
type InvoiceAggregator struct {
	store Storer
}

const basePrice = 315

func NewInvoiceAggregator(store Storer) Aggregator {
	return &InvoiceAggregator{
		store: store,
	}
}

func (i *InvoiceAggregator) AggregateDistance(distance *types.Distance) error {
	return i.store.Insert(distance)
}

func (i *InvoiceAggregator) CalculateInvoice(obuID int32) (*types.Invoice, error) {
	dist, err := i.store.Get(obuID)
	if err != nil {
		return nil, err
	}
	inv := &types.Invoice{
		OBUID:         obuID,
		TotalDistance: dist,
		Amount:        basePrice * dist,
	}
	return inv, nil
}

// MemoryStore implementation for testing
type MemoryStore struct {
	data map[int32]float64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[int32]float64),
	}
}

func (m *MemoryStore) Insert(d *types.Distance) error {
	m.data[d.OBUID] += d.Values
	return nil
}

func (m *MemoryStore) Get(id int32) (float64, error) {
	dist, ok := m.data[id]
	if !ok {
		return 0.0, fmt.Errorf("couldn't find distance for id: %d", id)
	}
	return dist, nil
}

// AggregatorServiceTestSuite is the test suite for aggregator service
type AggregatorServiceTestSuite struct {
	suite.Suite
	store      Storer
	aggregator Aggregator
}

// SetupTest runs before each test
func (suite *AggregatorServiceTestSuite) SetupTest() {
	suite.store = NewMemoryStore()
	suite.aggregator = NewInvoiceAggregator(suite.store)
}

// TestAggregateDistance_SingleDistance tests aggregating a single distance
func (suite *AggregatorServiceTestSuite) TestAggregateDistance_SingleDistance() {
	// Arrange
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5)

	// Act
	err := suite.aggregator.AggregateDistance(&distance)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify the distance was stored
	storedDistance, err := suite.store.Get(fixtures.TestOBUID1)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 10.5, storedDistance, 0.001)
}

// TestAggregateDistance_MultipleDistances tests aggregating multiple distances for same OBU
func (suite *AggregatorServiceTestSuite) TestAggregateDistance_MultipleDistances() {
	// Arrange
	distances := []types.Distance{
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5),
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 15.2),
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 8.3),
	}
	expectedTotal := 10.5 + 15.2 + 8.3

	// Act
	for _, distance := range distances {
		err := suite.aggregator.AggregateDistance(&distance)
		assert.NoError(suite.T(), err)
	}

	// Assert
	storedDistance, err := suite.store.Get(fixtures.TestOBUID1)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), expectedTotal, storedDistance, 0.001)
}

// TestAggregateDistance_MultipleDifferentOBUs tests aggregating distances for different OBUs
func (suite *AggregatorServiceTestSuite) TestAggregateDistance_MultipleDifferentOBUs() {
	// Arrange
	distances := []types.Distance{
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5),
		helpers.GenerateTestDistance(fixtures.TestOBUID2, 20.3),
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 5.2),
		helpers.GenerateTestDistance(fixtures.TestOBUID3, 15.7),
	}

	// Act
	for _, distance := range distances {
		err := suite.aggregator.AggregateDistance(&distance)
		assert.NoError(suite.T(), err)
	}

	// Assert
	distance1, err := suite.store.Get(fixtures.TestOBUID1)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 15.7, distance1, 0.001) // 10.5 + 5.2

	distance2, err := suite.store.Get(fixtures.TestOBUID2)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 20.3, distance2, 0.001)

	distance3, err := suite.store.Get(fixtures.TestOBUID3)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 15.7, distance3, 0.001)
}

// TestCalculateInvoice_ValidOBU tests calculating invoice for valid OBU
func (suite *AggregatorServiceTestSuite) TestCalculateInvoice_ValidOBU() {
	// Arrange
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 25.5)
	err := suite.aggregator.AggregateDistance(&distance)
	assert.NoError(suite.T(), err)

	// Act
	invoice, err := suite.aggregator.CalculateInvoice(fixtures.TestOBUID1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), invoice)
	assert.Equal(suite.T(), fixtures.TestOBUID1, invoice.OBUID)
	helpers.AssertFloatEquals(suite.T(), 25.5, invoice.TotalDistance, 0.001)
	helpers.AssertFloatEquals(suite.T(), 25.5*basePrice, invoice.Amount, 0.001)
}

// TestCalculateInvoice_NonExistentOBU tests calculating invoice for non-existent OBU
func (suite *AggregatorServiceTestSuite) TestCalculateInvoice_NonExistentOBU() {
	// Act
	invoice, err := suite.aggregator.CalculateInvoice(fixtures.TestOBUID1)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), invoice)
	assert.Contains(suite.T(), err.Error(), "couldn't find distance")
}

// TestCalculateInvoice_ZeroDistance tests calculating invoice for OBU with zero distance
func (suite *AggregatorServiceTestSuite) TestCalculateInvoice_ZeroDistance() {
	// Arrange
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 0.0)
	err := suite.aggregator.AggregateDistance(&distance)
	assert.NoError(suite.T(), err)

	// Act
	invoice, err := suite.aggregator.CalculateInvoice(fixtures.TestOBUID1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), invoice)
	assert.Equal(suite.T(), fixtures.TestOBUID1, invoice.OBUID)
	helpers.AssertFloatEquals(suite.T(), 0.0, invoice.TotalDistance, 0.001)
	helpers.AssertFloatEquals(suite.T(), 0.0, invoice.Amount, 0.001)
}

// TestCalculateInvoice_LargeDistance tests calculating invoice for large distance
func (suite *AggregatorServiceTestSuite) TestCalculateInvoice_LargeDistance() {
	// Arrange
	largeDistance := 1000000.5
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, largeDistance)
	err := suite.aggregator.AggregateDistance(&distance)
	assert.NoError(suite.T(), err)

	// Act
	invoice, err := suite.aggregator.CalculateInvoice(fixtures.TestOBUID1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), invoice)
	helpers.AssertFloatEquals(suite.T(), largeDistance, invoice.TotalDistance, 0.001)
	helpers.AssertFloatEquals(suite.T(), largeDistance*basePrice, invoice.Amount, 0.001)
}

// TestCalculateInvoice_AccumulatedDistances tests invoice calculation with accumulated distances
func (suite *AggregatorServiceTestSuite) TestCalculateInvoice_AccumulatedDistances() {
	// Arrange
	distances := []float64{10.5, 20.3, 15.7, 8.9}
	expectedTotal := 0.0

	for _, dist := range distances {
		distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, dist)
		err := suite.aggregator.AggregateDistance(&distance)
		assert.NoError(suite.T(), err)
		expectedTotal += dist
	}

	// Act
	invoice, err := suite.aggregator.CalculateInvoice(fixtures.TestOBUID1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), invoice)
	helpers.AssertFloatEquals(suite.T(), expectedTotal, invoice.TotalDistance, 0.001)
	helpers.AssertFloatEquals(suite.T(), expectedTotal*basePrice, invoice.Amount, 0.001)
}

// TestBasePrice tests that the base price constant is correct
func (suite *AggregatorServiceTestSuite) TestBasePrice() {
	assert.Equal(suite.T(), 315, basePrice, "Base price should be 315")
}

// Run the test suite
func TestAggregatorServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AggregatorServiceTestSuite))
}

// MemoryStoreTestSuite tests the memory store implementation
type MemoryStoreTestSuite struct {
	suite.Suite
	store *MemoryStore
}

func (suite *MemoryStoreTestSuite) SetupTest() {
	suite.store = NewMemoryStore()
}

// TestMemoryStore_Insert tests inserting data into memory store
func (suite *MemoryStoreTestSuite) TestMemoryStore_Insert() {
	// Arrange
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5)

	// Act
	err := suite.store.Insert(&distance)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify data was stored
	storedDistance, err := suite.store.Get(fixtures.TestOBUID1)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 10.5, storedDistance, 0.001)
}

// TestMemoryStore_InsertMultiple tests inserting multiple distances for same OBU
func (suite *MemoryStoreTestSuite) TestMemoryStore_InsertMultiple() {
	// Arrange
	distances := []types.Distance{
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5),
		helpers.GenerateTestDistance(fixtures.TestOBUID1, 15.2),
	}

	// Act
	for _, distance := range distances {
		err := suite.store.Insert(&distance)
		assert.NoError(suite.T(), err)
	}

	// Assert
	storedDistance, err := suite.store.Get(fixtures.TestOBUID1)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 25.7, storedDistance, 0.001) // 10.5 + 15.2
}

// TestMemoryStore_GetNonExistent tests getting non-existent OBU
func (suite *MemoryStoreTestSuite) TestMemoryStore_GetNonExistent() {
	// Act
	distance, err := suite.store.Get(fixtures.TestOBUID1)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), 0.0, distance)
	assert.Contains(suite.T(), err.Error(), "couldn't find distance")
}

// TestMemoryStore_MultipleDifferentOBUs tests storing data for different OBUs
func (suite *MemoryStoreTestSuite) TestMemoryStore_MultipleDifferentOBUs() {
	// Arrange
	distance1 := helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5)
	distance2 := helpers.GenerateTestDistance(fixtures.TestOBUID2, 20.3)

	// Act
	err := suite.store.Insert(&distance1)
	assert.NoError(suite.T(), err)
	err = suite.store.Insert(&distance2)
	assert.NoError(suite.T(), err)

	// Assert
	storedDistance1, err := suite.store.Get(fixtures.TestOBUID1)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 10.5, storedDistance1, 0.001)

	storedDistance2, err := suite.store.Get(fixtures.TestOBUID2)
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 20.3, storedDistance2, 0.001)
}

// Run the memory store test suite
func TestMemoryStoreTestSuite(t *testing.T) {
	suite.Run(t, new(MemoryStoreTestSuite))
}

// Benchmark tests
func BenchmarkAggregateDistance(b *testing.B) {
	store := NewMemoryStore()
	aggregator := NewInvoiceAggregator(store)
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = aggregator.AggregateDistance(&distance)
	}
}

func BenchmarkCalculateInvoice(b *testing.B) {
	store := NewMemoryStore()
	aggregator := NewInvoiceAggregator(store)
	distance := helpers.GenerateTestDistance(fixtures.TestOBUID1, 10.5)
	_ = aggregator.AggregateDistance(&distance)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aggregator.CalculateInvoice(fixtures.TestOBUID1)
	}
}
