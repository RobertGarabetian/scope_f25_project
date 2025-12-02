package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// --- Constants ---
const (
	ScreenWidth      = 1280
	ScreenHeight     = 720
	PlayerSize       = 32
	PlayerSpeed      = 5.0
	ScrollSpeed      = 3.0  // Speed at which the environment scrolls
	ObstacleMinGap   = 300   // Minimum vertical space for the path
	ObstacleMaxGap   = 400   // Maximum vertical space for the path
	NumFish          = 6     // Number of fish following the leader
	FishSize         = 24    // Size of each following fish
	FishFollowSpeed  = 4.0   // Speed at which fish follow
	CircleRadius     = 120.0 // Radius of the circle behind the leader
	CircleOffsetX    = -80.0 // X offset of the circle center behind the leader
	PlayerX          = 80.0  // Fixed X position of the leader
)

// --- Structs ---

// Obstacle defines a scrolling hazard
type Obstacle struct {
	x, y, width, height float64
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
	fish       []*Fish // Array of follower fish
	score      int
	gameOver   bool
	spawnTimer int
}

// --- Initialization ---

func init() {
	// Seed the random number generator for obstacle generation
	rand.Seed(time.Now().UnixNano())
}

// NewGame initializes the game state
func NewGame() *Game {
	centerY := float64(ScreenHeight)/2 - PlayerSize/2
	
	// Initialize fish array - place them randomly in a circle behind the leader
	fish := make([]*Fish, NumFish)
	circleCenterX := PlayerX + CircleOffsetX
	circleCenterY := centerY
	
	// Place fish randomly within circle, avoiding overlaps
	maxAttempts := 100
	for i := 0; i < NumFish; i++ {
		var fx, fy float64
		placed := false
		
		for attempt := 0; attempt < maxAttempts && !placed; attempt++ {
			// Random angle and distance within the circle
			angle := rand.Float64() * 2 * math.Pi
			// Use square root to get uniform distribution within circle
			radius := CircleRadius * math.Sqrt(rand.Float64())
			
			fx = circleCenterX + radius*math.Cos(angle)
			fy = circleCenterY + radius*math.Sin(angle)
			
			// Check if this position overlaps with existing fish
			overlaps := false
			for j := 0; j < i; j++ {
				dx := fx - fish[j].x
				dy := fy - fish[j].y
				distance := math.Sqrt(dx*dx + dy*dy)
				if distance < FishSize*1.2 { // Require some spacing between fish
					overlaps = true
					break
				}
			}
			
			if !overlaps {
				placed = true
			}
		}
		
		// Store relative offset from leader's center
		fish[i] = &Fish{
			x:       fx,
			y:       fy,
			offsetX: fx - PlayerX,
			offsetY: fy - centerY,
		}
	}
	
	g := &Game{
		// Center the player vertically on the left side
		playerY: centerY,
		obstacles:  make([]*Obstacle, 0),
		fish:       fish,
		score:      0,
		gameOver:   false,
		spawnTimer: 0,
	}
	return g
}

// --- Ebitengine Interface Implementations ---

func (g *Game) Update() error {
	if g.gameOver {
		// Reset game on pressing space
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			*g = *NewGame()
		}
		return nil
	}

	// 1. Handle Player Input
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.playerY -= PlayerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.playerY += PlayerSpeed
	}

	// Clamp playerY within the screen bounds
	if g.playerY < 0 {
		g.playerY = 0
	}
	if g.playerY > ScreenHeight-PlayerSize {
		g.playerY = float64(ScreenHeight - PlayerSize)
	}

	// 2. Update Fish Positions (following behavior with delay)
	for _, fish := range g.fish {
		// Calculate target position: leader position + fish's relative offset
		targetX := PlayerX + fish.offsetX
		targetY := g.playerY + fish.offsetY
		
		// Move fish towards target position smoothly (with delay)
		dx := targetX - fish.x
		dy := targetY - fish.y
		
		// Calculate distance to target
		distance := math.Sqrt(dx*dx + dy*dy)
		
		// Move towards target at follow speed
		if distance > FishFollowSpeed {
			// Normalize direction and apply speed
			fish.x += (dx / distance) * FishFollowSpeed
			fish.y += (dy / distance) * FishFollowSpeed
		} else {
			// Close enough, snap to target
			fish.x = targetX
			fish.y = targetY
		}
		
		// Clamp fish within screen bounds
		if fish.y < 0 {
			fish.y = 0
		}
		if fish.y > ScreenHeight-FishSize {
			fish.y = float64(ScreenHeight - FishSize)
		}
		if fish.x < 0 {
			fish.x = 0
		}
		if fish.x > ScreenWidth-FishSize {
			fish.x = float64(ScreenWidth - FishSize)
		}
	}

	// 3. Update Score
	g.score++

	// 4. Move and Cleanup Obstacles
	newObstacles := make([]*Obstacle, 0)
	for _, obs := range g.obstacles {
		obs.x -= ScrollSpeed // Scroll left
		if obs.x > -obs.width {
			newObstacles = append(newObstacles, obs)
		}
	}
	g.obstacles = newObstacles

	// 5. Collision Detection for Leader
	playerRect := collisionRect{
		x: PlayerX,
		y: g.playerY,
		w: PlayerSize,
		h: PlayerSize,
	}

	for _, obs := range g.obstacles {
		obsRect := collisionRect{
			x: obs.x,
			y: obs.y,
			w: obs.width,
			h: obs.height,
		}
		if checkAABBCollision(playerRect, obsRect) {
			g.gameOver = true
			break
		}
	}

	// 6. Collision Detection for all Fish
	if !g.gameOver {
		for _, fish := range g.fish {
			fishRect := collisionRect{
				x: fish.x,
				y: fish.y,
				w: FishSize,
				h: FishSize,
			}
			
			for _, obs := range g.obstacles {
				obsRect := collisionRect{
					x: obs.x,
					y: obs.y,
					w: obs.width,
					h: obs.height,
				}
				if checkAABBCollision(fishRect, obsRect) {
					g.gameOver = true
					break
				}
			}
			if g.gameOver {
				break
			}
		}
	}

	// 7. Spawn New Obstacles
	g.spawnTimer++
	// Spawn a new set of obstacles every 150 frames (approx. 2.5 seconds)
	if g.spawnTimer >= 150 {
		g.spawnTimer = 0
		g.spawnObstaclePair()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the background
	screen.Fill(color.RGBA{135, 206, 250, 255}) // Sky Blue (Water/Air)

	// Draw Obstacles (Trees/Rocks)
	obsColor := color.RGBA{150, 75, 0, 255} // Brown
	for _, obs := range g.obstacles {
		ebitenutil.DrawRect(screen, obs.x, obs.y, obs.width, obs.height, obsColor)
	}

	// Draw Player (The Leader)
	playerColor := color.RGBA{0, 100, 255, 255} // Deep Blue
	ebitenutil.DrawRect(screen, PlayerX, g.playerY, PlayerSize, PlayerSize, playerColor)
	
	// Draw all following fish
	fishColor := color.RGBA{0, 150, 255, 255} // Lighter Blue for followers
	for _, fish := range g.fish {
		ebitenutil.DrawRect(screen, fish.x, fish.y, FishSize, FishSize, fishColor)
	}

	// Draw Score
	scoreText := fmt.Sprintf("Time Survived: %.2f seconds", float64(g.score)/60.0)
	ebitenutil.DebugPrintAt(screen, scoreText, 10, 10)

	// Draw Game Over Message
	if g.gameOver {
		ebitenutil.DebugPrint(screen, "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n    GAME OVER!\n    Survival Time: "+fmt.Sprintf("%.2f s", float64(g.score)/60.0)+"\n    Press SPACE to restart.")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

// --- Game Logic Helpers ---

// spawnObstaclePair creates an upper and lower obstacle with a gap between them.
func (g *Game) spawnObstaclePair() {
	// Determine the gap size
	gapSize := ObstacleMinGap + rand.Float64()*(ObstacleMaxGap-ObstacleMinGap)

	// Determine the y-position of the gap (center)
	gapCenter := gapSize/2 + rand.Float64()*(ScreenHeight-gapSize)

	// Define the obstacle width
	obsWidth := float64(50)

	// 1. Create the Top Obstacle
	topHeight := gapCenter - gapSize/2
	if topHeight > 0 {
		topObs := &Obstacle{
			x:      ScreenWidth,
			y:      0,
			width:  obsWidth,
			height: topHeight,
		}
		g.obstacles = append(g.obstacles, topObs)
	}

	// 2. Create the Bottom Obstacle
	bottomY := gapCenter + gapSize/2
	bottomHeight := float64(ScreenHeight) - bottomY
	if bottomHeight > 0 {
		bottomObs := &Obstacle{
			x:      ScreenWidth,
			y:      bottomY,
			width:  obsWidth,
			height: bottomHeight,
		}
		g.obstacles = append(g.obstacles, bottomObs)
	}
}

// --- Collision Logic ---

// A simple structure to represent a bounding box for collision checking
type collisionRect struct {
	x, y, w, h float64
}

// checkAABBCollision performs Axis-Aligned Bounding Box collision detection.
func checkAABBCollision(r1, r2 collisionRect) bool {
	return r1.x < r2.x+r2.w &&
		r1.x+r1.w > r2.x &&
		r1.y < r2.y+r2.h &&
		r1.y+r1.h > r2.y
}

// --- Main Function ---

func main() {
	game := NewGame()

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("The Migratory Path (Wildlife Game)")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
