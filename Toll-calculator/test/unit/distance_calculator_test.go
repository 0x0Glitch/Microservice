package unit

import (
	"math"
	"testing"

	"github.com/0x0Glitch/toll-calculator/test/fixtures"
	"github.com/0x0Glitch/toll-calculator/test/helpers"
	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DistanceCalculatorTestSuite is the test suite for distance calculator
type DistanceCalculatorTestSuite struct {
	suite.Suite
	service CalculatorServicer
}

// CalculatorServicer interface for testing (copied from main package)
type CalculatorServicer interface {
	CalculateDistance(types.OBUData) (float64, error)
}

// CalculatorService implementation for testing
type CalculatorService struct {
	prevPoint []float64
}

func NewCalculatorService() CalculatorServicer {
	return &CalculatorService{}
}

func (s *CalculatorService) CalculateDistance(data types.OBUData) (float64, error) {
	distance := 0.0

	if len(s.prevPoint) > 0 {
		distance = calculateDistancer(s.prevPoint[0], s.prevPoint[1], data.Lat, data.Long)
	}
	s.prevPoint = []float64{data.Lat, data.Long}
	return distance, nil
}

func calculateDistancer(x1, x2, y1, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

// SetupTest runs before each test
func (suite *DistanceCalculatorTestSuite) SetupTest() {
	suite.service = NewCalculatorService()
}

// TestCalculateDistance_FirstPoint tests distance calculation for the first point
func (suite *DistanceCalculatorTestSuite) TestCalculateDistance_FirstPoint() {
	// Arrange
	obuData := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7128, -74.0060)

	// Act
	distance, err := suite.service.CalculateDistance(obuData)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0.0, distance, "First point should have zero distance")
}

// TestCalculateDistance_SecondPoint tests distance calculation for the second point
func (suite *DistanceCalculatorTestSuite) TestCalculateDistance_SecondPoint() {
	// Arrange
	firstPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 0.0, 0.0)
	secondPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 3.0, 4.0)

	// Act
	_, err := suite.service.CalculateDistance(firstPoint)
	assert.NoError(suite.T(), err)

	distance, err := suite.service.CalculateDistance(secondPoint)

	// Assert
	assert.NoError(suite.T(), err)
	helpers.AssertFloatEquals(suite.T(), 5.0, distance, 0.001, "Distance should be 5.0 for 3-4-5 triangle")
}

// TestCalculateDistance_MultiplePoints tests distance calculation for multiple points
func (suite *DistanceCalculatorTestSuite) TestCalculateDistance_MultiplePoints() {
	// Arrange
	points := []types.OBUData{
		helpers.GenerateTestOBUData(fixtures.TestOBUID1, 0.0, 0.0),
		helpers.GenerateTestOBUData(fixtures.TestOBUID1, 1.0, 0.0),
		helpers.GenerateTestOBUData(fixtures.TestOBUID1, 1.0, 1.0),
		helpers.GenerateTestOBUData(fixtures.TestOBUID1, 0.0, 1.0),
	}
	expectedDistances := []float64{0.0, 1.0, 1.0, 1.0}

	// Act & Assert
	for i, point := range points {
		distance, err := suite.service.CalculateDistance(point)
		assert.NoError(suite.T(), err)
		helpers.AssertFloatEquals(suite.T(), expectedDistances[i], distance, 0.001,
			"Distance for point %d should be %f", i, expectedDistances[i])
	}
}

// TestCalculateDistance_SamePoint tests distance calculation for the same point
func (suite *DistanceCalculatorTestSuite) TestCalculateDistance_SamePoint() {
	// Arrange
	point := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7128, -74.0060)

	// Act
	_, err := suite.service.CalculateDistance(point)
	assert.NoError(suite.T(), err)

	distance, err := suite.service.CalculateDistance(point)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0.0, distance, "Same point should have zero distance")
}

// TestCalculateDistance_NegativeCoordinates tests distance calculation with negative coordinates
func (suite *DistanceCalculatorTestSuite) TestCalculateDistance_NegativeCoordinates() {
	// Arrange
	firstPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, -1.0, -1.0)
	secondPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 2.0, 3.0)

	// Act
	_, err := suite.service.CalculateDistance(firstPoint)
	assert.NoError(suite.T(), err)

	distance, err := suite.service.CalculateDistance(secondPoint)

	// Assert
	assert.NoError(suite.T(), err)
	expectedDistance := math.Sqrt(math.Pow(3.0, 2) + math.Pow(4.0, 2)) // 5.0
	helpers.AssertFloatEquals(suite.T(), expectedDistance, distance, 0.001)
}

// TestCalculateDistance_LargeCoordinates tests distance calculation with large coordinates
func (suite *DistanceCalculatorTestSuite) TestCalculateDistance_LargeCoordinates() {
	// Arrange
	firstPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 1000.0, 2000.0)
	secondPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 1003.0, 2004.0)

	// Act
	_, err := suite.service.CalculateDistance(firstPoint)
	assert.NoError(suite.T(), err)

	distance, err := suite.service.CalculateDistance(secondPoint)

	// Assert
	assert.NoError(suite.T(), err)
	expectedDistance := math.Sqrt(math.Pow(3.0, 2) + math.Pow(4.0, 2)) // 5.0
	helpers.AssertFloatEquals(suite.T(), expectedDistance, distance, 0.001)
}

// TestCalculateDistance_ZeroCoordinates tests distance calculation with zero coordinates
func (suite *DistanceCalculatorTestSuite) TestCalculateDistance_ZeroCoordinates() {
	// Arrange
	firstPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 0.0, 0.0)
	secondPoint := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 0.0, 0.0)

	// Act
	_, err := suite.service.CalculateDistance(firstPoint)
	assert.NoError(suite.T(), err)

	distance, err := suite.service.CalculateDistance(secondPoint)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0.0, distance, "Zero coordinates should have zero distance")
}

// TestCalculateDistancer_DirectFunction tests the distance calculation function directly
func (suite *DistanceCalculatorTestSuite) TestCalculateDistancer_DirectFunction() {
	testCases := fixtures.GetTestCoordinatePairs()

	for i, tc := range testCases {
		// Act
		distance := calculateDistancer(tc.From.Lat, tc.From.Long, tc.To.Lat, tc.To.Long)

		// Assert
		helpers.AssertFloatEquals(suite.T(), tc.ExpectedDistance, distance, 0.001,
			"Test case %d: expected %f, got %f", i, tc.ExpectedDistance, distance)
	}
}

// TestCalculateDistancer_Symmetry tests that distance calculation is symmetric
func (suite *DistanceCalculatorTestSuite) TestCalculateDistancer_Symmetry() {
	// Arrange
	x1, y1 := 1.0, 2.0
	x2, y2 := 4.0, 6.0

	// Act
	distance1 := calculateDistancer(x1, y1, x2, y2)
	distance2 := calculateDistancer(x2, y2, x1, y1)

	// Assert
	helpers.AssertFloatEquals(suite.T(), distance1, distance2, 0.001,
		"Distance calculation should be symmetric")
}

// TestCalculateDistancer_TriangleInequality tests triangle inequality
func (suite *DistanceCalculatorTestSuite) TestCalculateDistancer_TriangleInequality() {
	// Arrange
	x1, y1 := 0.0, 0.0
	x2, y2 := 3.0, 0.0
	x3, y3 := 0.0, 4.0

	// Act
	d12 := calculateDistancer(x1, y1, x2, y2)
	d23 := calculateDistancer(x2, y2, x3, y3)
	d13 := calculateDistancer(x1, y1, x3, y3)

	// Assert - Triangle inequality: d13 <= d12 + d23
	assert.True(suite.T(), d13 <= d12+d23+0.001,
		"Triangle inequality should hold: %f <= %f + %f", d13, d12, d23)
}

// Run the test suite
func TestDistanceCalculatorTestSuite(t *testing.T) {
	suite.Run(t, new(DistanceCalculatorTestSuite))
}

// Benchmark tests
func BenchmarkCalculateDistance(b *testing.B) {
	service := NewCalculatorService()
	obuData := helpers.GenerateTestOBUData(fixtures.TestOBUID1, 40.7128, -74.0060)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CalculateDistance(obuData)
	}
}

func BenchmarkCalculateDistancer(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateDistancer(40.7128, -74.0060, 34.0522, -118.2437)
	}
}
