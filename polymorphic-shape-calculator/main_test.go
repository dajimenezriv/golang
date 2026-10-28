package main

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const (
	epsilon = 1e-9 // Small value for floating point comparisons
)

// Helper function to check if two float64 values are approximately equal
func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// TestRectangleConstructor tests the NewRectangle constructor
func TestRectangleConstructor(t *testing.T) {
	tests := []struct {
		name        string
		width       float64
		height      float64
		shouldError bool
	}{
		{"Valid rectangle", 5.0, 3.0, false},
		{"Zero width", 0, 3.0, true},
		{"Negative width", -2.0, 3.0, true},
		{"Zero height", 5.0, 0, true},
		{"Negative height", 5.0, -2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rect, err := NewRectangle(tt.width, tt.height)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected an error for width=%.2f, height=%.2f, but got none", tt.width, tt.height)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for width=%.2f, height=%.2f, but got: %v", tt.width, tt.height, err)
				}
				if rect == nil {
					t.Fatal("Rectangle should not be nil when no error")
				}
				if rect.Width != tt.width {
					t.Errorf("Expected width=%.2f, got %.2f", tt.width, rect.Width)
				}
				if rect.Height != tt.height {
					t.Errorf("Expected height=%.2f, got %.2f", tt.height, rect.Height)
				}
			}
		})
	}
}

// TestCircleConstructor tests the NewCircle constructor
func TestCircleConstructor(t *testing.T) {
	tests := []struct {
		name        string
		radius      float64
		shouldError bool
	}{
		{"Valid circle", 5.0, false},
		{"Zero radius", 0, true},
		{"Negative radius", -2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circle, err := NewCircle(tt.radius)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected an error for radius=%.2f, but got none", tt.radius)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for radius=%.2f, but got: %v", tt.radius, err)
				}
				if circle == nil {
					t.Fatal("Circle should not be nil when no error")
				}
				if circle.Radius != tt.radius {
					t.Errorf("Expected radius=%.2f, got %.2f", tt.radius, circle.Radius)
				}
			}
		})
	}
}

// TestTriangleConstructor tests the NewTriangle constructor
func TestTriangleConstructor(t *testing.T) {
	tests := []struct {
		name        string
		a, b, c     float64
		shouldError bool
	}{
		{"Valid triangle", 3.0, 4.0, 5.0, false},
		{"Equilateral triangle", 5.0, 5.0, 5.0, false},
		{"Invalid triangle (a + b = c)", 3.0, 4.0, 7.0, true},
		{"Invalid triangle (a + b < c)", 3.0, 4.0, 8.0, true},
		{"Zero side a", 0, 4.0, 5.0, true},
		{"Zero side b", 3.0, 0, 5.0, true},
		{"Zero side c", 3.0, 4.0, 0, true},
		{"Negative side a", -3.0, 4.0, 5.0, true},
		{"Negative side b", 3.0, -4.0, 5.0, true},
		{"Negative side c", 3.0, 4.0, -5.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triangle, err := NewTriangle(tt.a, tt.b, tt.c)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected an error for sides (%.2f, %.2f, %.2f), but got none", tt.a, tt.b, tt.c)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for sides (%.2f, %.2f, %.2f), but got: %v", tt.a, tt.b, tt.c, err)
				}
				if triangle == nil {
					t.Fatal("Triangle should not be nil when no error")
				}
				if triangle.SideA != tt.a {
					t.Errorf("Expected SideA=%.2f, got %.2f", tt.a, triangle.SideA)
				}
				if triangle.SideB != tt.b {
					t.Errorf("Expected SideB=%.2f, got %.2f", tt.b, triangle.SideB)
				}
				if triangle.SideC != tt.c {
					t.Errorf("Expected SideC=%.2f, got %.2f", tt.c, triangle.SideC)
				}
			}
		})
	}
}

// TestRectangleArea tests the Area method of Rectangle
func TestRectangleArea(t *testing.T) {
	tests := []struct {
		width  float64
		height float64
		area   float64
	}{
		{5.0, 3.0, 15.0},
		{10.0, 10.0, 100.0},
		{2.5, 6.0, 15.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2fx%.2f", tt.width, tt.height), func(t *testing.T) {
			rect, err := NewRectangle(tt.width, tt.height)
			if err != nil {
				t.Fatalf("Failed to create rectangle: %v", err)
			}

			area := rect.Area()
			if !floatEquals(area, tt.area) {
				t.Errorf("Expected area=%.2f, got %.2f", tt.area, area)
			}
		})
	}
}

// TestRectanglePerimeter tests the Perimeter method of Rectangle
func TestRectanglePerimeter(t *testing.T) {
	tests := []struct {
		width     float64
		height    float64
		perimeter float64
	}{
		{5.0, 3.0, 16.0},
		{10.0, 10.0, 40.0},
		{2.5, 6.0, 17.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2fx%.2f", tt.width, tt.height), func(t *testing.T) {
			rect, err := NewRectangle(tt.width, tt.height)
			if err != nil {
				t.Fatalf("Failed to create rectangle: %v", err)
			}

			perimeter := rect.Perimeter()
			if !floatEquals(perimeter, tt.perimeter) {
				t.Errorf("Expected perimeter=%.2f, got %.2f", tt.perimeter, perimeter)
			}
		})
	}
}

// TestCircleArea tests the Area method of Circle
func TestCircleArea(t *testing.T) {
	tests := []struct {
		radius float64
		area   float64
	}{
		{5.0, math.Pi * 25.0},
		{1.0, math.Pi},
		{0.5, math.Pi * 0.25},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("r=%.2f", tt.radius), func(t *testing.T) {
			circle, err := NewCircle(tt.radius)
			if err != nil {
				t.Fatalf("Failed to create circle: %v", err)
			}

			area := circle.Area()
			if !floatEquals(area, tt.area) {
				t.Errorf("Expected area=%.2f, got %.2f", tt.area, area)
			}
		})
	}
}

// TestCirclePerimeter tests the Perimeter method of Circle
func TestCirclePerimeter(t *testing.T) {
	tests := []struct {
		radius    float64
		perimeter float64
	}{
		{5.0, 2 * math.Pi * 5.0},
		{1.0, 2 * math.Pi},
		{0.5, math.Pi},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("r=%.2f", tt.radius), func(t *testing.T) {
			circle, err := NewCircle(tt.radius)
			if err != nil {
				t.Fatalf("Failed to create circle: %v", err)
			}

			perimeter := circle.Perimeter()
			if !floatEquals(perimeter, tt.perimeter) {
				t.Errorf("Expected perimeter=%.2f, got %.2f", tt.perimeter, perimeter)
			}
		})
	}
}

// TestTriangleArea tests the Area method of Triangle using Heron's formula
func TestTriangleArea(t *testing.T) {
	tests := []struct {
		a, b, c float64
		area    float64
	}{
		{3.0, 4.0, 5.0, 6.0},                // Right triangle
		{5.0, 5.0, 6.0, 12.0},               // Isosceles triangle
		{5.0, 5.0, 5.0, 10.825317547305483}, // Equilateral triangle
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("sides=%.2f,%.2f,%.2f", tt.a, tt.b, tt.c), func(t *testing.T) {
			triangle, err := NewTriangle(tt.a, tt.b, tt.c)
			if err != nil {
				t.Fatalf("Failed to create triangle: %v", err)
			}

			area := triangle.Area()
			if !floatEquals(area, tt.area) {
				t.Errorf("Expected area=%.8f, got %.8f", tt.area, area)
			}
		})
	}
}

// TestTrianglePerimeter tests the Perimeter method of Triangle
func TestTrianglePerimeter(t *testing.T) {
	tests := []struct {
		a, b, c   float64
		perimeter float64
	}{
		{3.0, 4.0, 5.0, 12.0},
		{5.0, 5.0, 6.0, 16.0},
		{5.0, 5.0, 5.0, 15.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("sides=%.2f,%.2f,%.2f", tt.a, tt.b, tt.c), func(t *testing.T) {
			triangle, err := NewTriangle(tt.a, tt.b, tt.c)
			if err != nil {
				t.Fatalf("Failed to create triangle: %v", err)
			}

			perimeter := triangle.Perimeter()
			if !floatEquals(perimeter, tt.perimeter) {
				t.Errorf("Expected perimeter=%.2f, got %.2f", tt.perimeter, perimeter)
			}
		})
	}
}

// TestShapeStringMethod tests the String() method of all shapes
func TestShapeStringMethod(t *testing.T) {
	tests := []struct {
		name          string
		shape         Shape
		expectedParts []string
	}{
		{
			"Rectangle",
			&Rectangle{Width: 5.0, Height: 3.0},
			[]string{"Rectangle", "5", "3", "width", "height"},
		},
		{
			"Circle",
			&Circle{Radius: 4.0},
			[]string{"Circle", "4", "radius"},
		},
		{
			"Triangle",
			&Triangle{SideA: 3.0, SideB: 4.0, SideC: 5.0},
			[]string{"Triangle", "3", "4", "5", "sides"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.shape.String()

			for _, part := range tt.expectedParts {
				if !strings.Contains(strings.ToLower(str), strings.ToLower(part)) {
					t.Errorf("Expected string representation to contain '%s', but got: %s", part, str)
				}
			}
		})
	}
}

// TestShapeCalculatorTotalArea tests the TotalArea method of ShapeCalculator
func TestShapeCalculatorTotalArea(t *testing.T) {
	rect, _ := NewRectangle(5.0, 3.0)         // Area = 15
	circle, _ := NewCircle(2.0)               // Area = 12.57
	triangle, _ := NewTriangle(3.0, 4.0, 5.0) // Area = 6

	calculator := NewShapeCalculator()
	shapes := []Shape{rect, circle, triangle}

	totalArea := calculator.TotalArea(shapes)
	expectedArea := 15.0 + math.Pi*4.0 + 6.0

	if !floatEquals(totalArea, expectedArea) {
		t.Errorf("Expected total area = %.2f, got %.2f", expectedArea, totalArea)
	}
}

// TestShapeCalculatorLargestShape tests the LargestShape method of ShapeCalculator
func TestShapeCalculatorLargestShape(t *testing.T) {
	rect, _ := NewRectangle(5.0, 3.0)         // Area = 15
	circle, _ := NewCircle(2.0)               // Area = 12.57
	triangle, _ := NewTriangle(3.0, 4.0, 5.0) // Area = 6

	calculator := NewShapeCalculator()
	shapes := []Shape{rect, circle, triangle}

	largest := calculator.LargestShape(shapes)

	if largest != rect {
		t.Errorf("Expected largest shape to be Rectangle, got %T with area %.2f", largest, largest.Area())
	}
}

// TestShapeCalculatorSortByArea tests the SortByArea method of ShapeCalculator
func TestShapeCalculatorSortByArea(t *testing.T) {
	rect, _ := NewRectangle(5.0, 3.0)         // Area = 15
	circle, _ := NewCircle(2.0)               // Area = 12.57
	triangle, _ := NewTriangle(3.0, 4.0, 5.0) // Area = 6

	calculator := NewShapeCalculator()
	shapes := []Shape{rect, circle, triangle}

	// Test ascending order
	sorted := calculator.SortByArea(shapes, true)

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 shapes after sorting, got %d", len(sorted))
	}

	expectedOrder := []Shape{triangle, circle, rect} // Smallest to largest
	for i, shape := range sorted {
		if shape.Area() != expectedOrder[i].Area() {
			t.Errorf("Wrong order in ascending sort at index %d: expected area %.2f, got %.2f",
				i, expectedOrder[i].Area(), shape.Area())
		}
	}

	// Test descending order
	sorted = calculator.SortByArea(shapes, false)

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 shapes after sorting, got %d", len(sorted))
	}

	expectedOrder = []Shape{rect, circle, triangle} // Largest to smallest
	for i, shape := range sorted {
		if shape.Area() != expectedOrder[i].Area() {
			t.Errorf("Wrong order in descending sort at index %d: expected area %.2f, got %.2f",
				i, expectedOrder[i].Area(), shape.Area())
		}
	}
}

// TestShapeInterfaceCompliance ensures all shapes implement the Shape interface
func TestShapeInterfaceCompliance(t *testing.T) {
	// This is more of a compile-time check, but let's verify at runtime too
	var shapes []Shape

	rect, _ := NewRectangle(5.0, 3.0)
	shapes = append(shapes, rect)

	circle, _ := NewCircle(2.0)
	shapes = append(shapes, circle)

	triangle, _ := NewTriangle(3.0, 4.0, 5.0)
	shapes = append(shapes, triangle)

	// No panic means they all implement the interface correctly
	for i, shape := range shapes {
		if shape == nil {
			t.Errorf("Shape at index %d is nil", i)
		} else {
			t.Logf("Shape %d: %T, Area: %.2f, Perimeter: %.2f, String: %s",
				i, shape, shape.Area(), shape.Perimeter(), shape.String())
		}
	}
}

// ByArea implements sort.Interface for []Shape based on Area
type ByArea []Shape

func (a ByArea) Len() int           { return len(a) }
func (a ByArea) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByArea) Less(i, j int) bool { return a[i].Area() < a[j].Area() }

// TestSortingShapes tests that shapes can be sorted using Sort package
func TestSortingShapes(t *testing.T) {
	rect, _ := NewRectangle(5.0, 3.0)         // Area = 15
	circle, _ := NewCircle(2.0)               // Area = 12.57
	triangle, _ := NewTriangle(3.0, 4.0, 5.0) // Area = 6

	shapes := []Shape{rect, circle, triangle}
	sort.Sort(ByArea(shapes))

	expected := []Shape{triangle, circle, rect} // Smallest to largest
	if !reflect.DeepEqual(shapes, expected) {
		t.Errorf("Shapes not sorted correctly by area")
		for i, shape := range shapes {
			t.Logf("Shape %d: %T, Area: %.2f", i, shape, shape.Area())
		}
	}
}
