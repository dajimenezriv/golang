package main

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"
)

// ErrInvalidDimension is returned when a shape is built with impossible dimensions
var ErrInvalidDimension = errors.New("invalid dimension")

// positiveFinite reports whether every value is > 0 and not +Inf.
// NaN fails the > 0 comparison, so it is rejected too.
func positiveFinite(vs ...float64) bool {
	for _, v := range vs {
		if !(v > 0) || math.IsInf(v, 1) {
			return false
		}
	}
	return true
}

// Shape interface defines methods that all shapes must implement
type Shape interface {
	Area() float64
	Perimeter() float64
	fmt.Stringer
}

// Rectangle represents a four-sided shape with perpendicular sides
type Rectangle struct {
	Width  float64
	Height float64
}

// NewRectangle creates a new Rectangle with validation
func NewRectangle(width, height float64) (*Rectangle, error) {
	if !positiveFinite(width, height) {
		return nil, fmt.Errorf("%w: width %g, height %g must be positive and finite", ErrInvalidDimension, width, height)
	}
	return &Rectangle{Width: width, Height: height}, nil
}

// Area calculates the area of the rectangle
func (r *Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter calculates the perimeter of the rectangle
func (r *Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// String returns a string representation of the rectangle
func (r *Rectangle) String() string {
	return fmt.Sprintf("Rectangle(width: %g, height: %g)", r.Width, r.Height)
}

// Circle represents a perfectly round shape
type Circle struct {
	Radius float64
}

// NewCircle creates a new Circle with validation
func NewCircle(radius float64) (*Circle, error) {
	if !positiveFinite(radius) {
		return nil, fmt.Errorf("%w: radius %g must be positive and finite", ErrInvalidDimension, radius)
	}
	return &Circle{Radius: radius}, nil
}

// Area calculates the area of the circle
func (c *Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Perimeter calculates the circumference of the circle
func (c *Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// String returns a string representation of the circle
func (c *Circle) String() string {
	return fmt.Sprintf("Circle(radius: %g)", c.Radius)
}

// Triangle represents a three-sided polygon
type Triangle struct {
	SideA float64
	SideB float64
	SideC float64
}

// NewTriangle creates a new Triangle with validation
func NewTriangle(a, b, c float64) (*Triangle, error) {
	if !positiveFinite(a, b, c) {
		return nil, fmt.Errorf("%w: sides %g, %g, %g must be positive and finite", ErrInvalidDimension, a, b, c)
	}
	if a+b <= c || a+c <= b || b+c <= a {
		return nil, fmt.Errorf("%w: sides %g, %g, %g violate the triangle inequality", ErrInvalidDimension, a, b, c)
	}
	return &Triangle{SideA: a, SideB: b, SideC: c}, nil
}

// Area calculates the area of the triangle using Heron's formula
func (t *Triangle) Area() float64 {
	s := (t.SideA + t.SideB + t.SideC) / 2
	return math.Sqrt(s * (s - t.SideA) * (s - t.SideB) * (s - t.SideC))
}

// Perimeter calculates the perimeter of the triangle
func (t *Triangle) Perimeter() float64 {
	return t.SideA + t.SideB + t.SideC
}

// String returns a string representation of the triangle
func (t *Triangle) String() string {
	return fmt.Sprintf("Triangle(sides: %g, %g, %g)", t.SideA, t.SideB, t.SideC)
}

// ShapeCalculator provides utility functions for shapes
type ShapeCalculator struct{}

// NewShapeCalculator creates a new ShapeCalculator
func NewShapeCalculator() *ShapeCalculator {
	return &ShapeCalculator{}
}

// PrintProperties prints the properties of a shape
func (sc *ShapeCalculator) PrintProperties(s Shape) {
	fmt.Printf("%v area=%.2f perimeter=%.2f\n", s, s.Area(), s.Perimeter())
}

// TotalArea calculates the sum of areas of all shapes
func (sc *ShapeCalculator) TotalArea(shapes []Shape) float64 {
	var sum float64
	for _, s := range shapes {
		sum += s.Area()
	}
	return sum
}

// LargestShape finds the shape with the largest area
func (sc *ShapeCalculator) LargestShape(shapes []Shape) Shape {
	if len(shapes) == 0 {
		return nil
	}
	return slices.MaxFunc(shapes, func(a Shape, b Shape) int {
		return cmp.Compare(a.Area(), b.Area())
	})
}

// SortByArea sorts shapes by area in ascending or descending order
func (sc *ShapeCalculator) SortByArea(shapes []Shape, ascending bool) []Shape {
	s := slices.Clone(shapes)
	slices.SortFunc(s, func(a Shape, b Shape) int {
		if ascending {
			return cmp.Compare(a.Area(), b.Area())
		}
		return cmp.Compare(b.Area(), a.Area())
	})
	return s
}
