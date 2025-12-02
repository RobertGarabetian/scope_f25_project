package main

import "math"

// --- Collision Logic ---

// checkAABBCollision performs Axis-Aligned Bounding Box collision detection.
func checkAABBCollision(r1, r2 collisionRect) bool {
	return r1.x < r2.x+r2.w &&
		r1.x+r1.w > r2.x &&
		r1.y < r2.y+r2.h &&
		r1.y+r1.h > r2.y
}

// circleCollision represents a circle for collision detection
type circleCollision struct {
	x, y, radius float64
}

// checkCircleCollision checks if two circles overlap
func checkCircleCollision(c1, c2 circleCollision) bool {
	dx := c1.x - c2.x
	dy := c1.y - c2.y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance < (c1.radius + c2.radius)
}

// checkCircleRectCollision checks if a circle overlaps with a rectangle
func checkCircleRectCollision(circle circleCollision, rect collisionRect) bool {
	// Find the closest point on the rectangle to the circle center
	closestX := math.Max(rect.x, math.Min(circle.x, rect.x+rect.w))
	closestY := math.Max(rect.y, math.Min(circle.y, rect.y+rect.h))
	
	// Calculate distance from circle center to closest point
	dx := circle.x - closestX
	dy := circle.y - closestY
	distance := math.Sqrt(dx*dx + dy*dy)
	
	return distance < circle.radius
}

