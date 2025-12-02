package main

import "github.com/hajimehoshi/ebiten/v2"

// --- Structs ---

// Obstacle defines a scrolling hazard
type Obstacle struct {
	x, y, width, height float64
	passed              bool // Track if this obstacle has been passed for scoring
}

// Coin represents a collectible coin
type Coin struct {
	x, y, size float64
	collected  bool
}

// Fish represents a follower fish that trails behind the leader
type Fish struct {
	x, y           float64 // Current position of the fish
	offsetX, offsetY float64 // Relative offset from the leader's position
}

// Game holds the entire game state
type Game struct {
	playerY    float64
	obstacles  []*Obstacle
	coins      []*Coin // Array of coins
	fish       []*Fish // Array of follower fish
	score      int     // Score based on obstacles passed
	coinsCollected int // Number of coins collected
	gameOver   bool
	spawnTimer int
	restartInput string // Input string for restart code
	// Sprites
	fishSprite    *ebiten.Image // Pixel art sprite for fish
	kelpSprite    *ebiten.Image // Pixel art sprite for kelp (will be scaled)
}

// A simple structure to represent a bounding box for collision checking
type collisionRect struct {
	x, y, w, h float64
}

