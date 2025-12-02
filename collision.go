package main

// --- Collision Logic ---

// checkAABBCollision performs Axis-Aligned Bounding Box collision detection.
func checkAABBCollision(r1, r2 collisionRect) bool {
	return r1.x < r2.x+r2.w &&
		r1.x+r1.w > r2.x &&
		r1.y < r2.y+r2.h &&
		r1.y+r1.h > r2.y
}

