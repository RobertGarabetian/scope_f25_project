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
	offsetX, offsetY float64 // Base relative offset from the leader's position (center of wander circle)
	targetOffsetX, targetOffsetY float64 // Random target offset for wandering
	wanderTimer    int     // Timer to change wander target
}

// BackgroundFish represents ambient fish swimming in the background
type BackgroundFish struct {
	x, y       float64 // Current position
	speed      float64 // Swimming speed
	direction  int     // 1 for right, -1 for left
	size       float64 // Size of the fish
	depth      float64 // Depth factor (0.0 to 1.0, lower = further back)
}

// Bubble represents a bubble floating upward
type Bubble struct {
	x, y       float64 // Current position
	speed      float64 // Rising speed
	size       float64 // Size of the bubble
	wobble     float64 // Horizontal wobble offset
	wobbleSpeed float64 // Speed of wobble animation
}

// Game holds the entire game state
type Game struct {
	playerY    float64
	obstacles  []*Obstacle
	coins      []*Coin // Array of coins
	fish       []*Fish // Array of follower fish
	backgroundFish []*BackgroundFish // Array of background ambient fish
	bubbles    []*Bubble // Array of floating bubbles
	score      int     // Score based on obstacles passed
	coinsCollected int // Number of coins collected
	gameOver   bool
	spawnTimer int
	restartInput string // Input string for restart code
	gameTime   int     // Total frames elapsed (for speed increase)
	speedMultiplier float64 // Current speed multiplier
	// Sprites
	fishSprite    *ebiten.Image // Pixel art sprite for fish
	kelpSprite    *ebiten.Image // Pixel art sprite for kelp (will be scaled)
}

// A simple structure to represent a bounding box for collision checking
type collisionRect struct {
	x, y, w, h float64
}

