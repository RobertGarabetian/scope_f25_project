package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

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
		coins:      make([]*Coin, 0),
		fish:       fish,
		score:      0,
		coinsCollected: 0,
		gameOver:   false,
		spawnTimer: 0,
		fishSprite: createFishSprite(),
		kelpSprite: createKelpSprite(),
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

	// 3. Move and Cleanup Obstacles, Update Score
	newObstacles := make([]*Obstacle, 0)
	for _, obs := range g.obstacles {
		obs.x -= ScrollSpeed // Scroll left
		
		// Check if obstacle has been passed (player has passed it)
		if !obs.passed && obs.x+obs.width < PlayerX {
			obs.passed = true
			g.score++ // Increment score when obstacle is passed
		}
		
		if obs.x > -obs.width {
			newObstacles = append(newObstacles, obs)
		}
	}
	g.obstacles = newObstacles

	// 4. Move and Cleanup Coins
	newCoins := make([]*Coin, 0)
	for _, coin := range g.coins {
		coin.x -= ScrollSpeed // Scroll left
		
		// Remove coins that are off-screen or collected
		if coin.x > -coin.size && !coin.collected {
			newCoins = append(newCoins, coin)
		}
	}
	g.coins = newCoins

	// 5. Collision Detection for Leader with Obstacles
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

	// 6. Coin Collection Detection for Leader
	if !g.gameOver {
		for _, coin := range g.coins {
			if !coin.collected {
				coinRect := collisionRect{
					x: coin.x,
					y: coin.y,
					w: coin.size,
					h: coin.size,
				}
				if checkAABBCollision(playerRect, coinRect) {
					coin.collected = true
					g.coinsCollected++
				}
			}
		}
	}

	// 7. Collision Detection for all Fish with Obstacles
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

	// 8. Coin Collection Detection for Fish
	if !g.gameOver {
		for _, fish := range g.fish {
			fishRect := collisionRect{
				x: fish.x,
				y: fish.y,
				w: FishSize,
				h: FishSize,
			}
			
			for _, coin := range g.coins {
				if !coin.collected {
					coinRect := collisionRect{
						x: coin.x,
						y: coin.y,
						w: coin.size,
						h: coin.size,
					}
					if checkAABBCollision(fishRect, coinRect) {
						coin.collected = true
						g.coinsCollected++
					}
				}
			}
		}
	}

	// 9. Spawn New Obstacles
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

	// Draw Obstacles (Kelp)
	for _, obs := range g.obstacles {
		g.drawKelp(screen, obs.x, obs.y, obs.width, obs.height)
	}

	// Draw Coins
	coinColor := color.RGBA{255, 215, 0, 255} // Gold
	for _, coin := range g.coins {
		if !coin.collected {
			ebitenutil.DrawRect(screen, coin.x, coin.y, coin.size, coin.size, coinColor)
		}
	}

	// Draw Player (The Leader)
	g.drawFish(screen, PlayerX, g.playerY, PlayerSize, true)
	
	// Draw all following fish
	for _, fish := range g.fish {
		g.drawFish(screen, fish.x, fish.y, FishSize, false)
	}

	// Draw Score and Coin Count
	scoreText := fmt.Sprintf("Score: %d", g.score)
	coinText := fmt.Sprintf("Coins: %d", g.coinsCollected)
	ebitenutil.DebugPrintAt(screen, scoreText, 10, 10)
	ebitenutil.DebugPrintAt(screen, coinText, 10, 30)

	// Draw Game Over Screen
	if g.gameOver {
		// Draw semi-transparent overlay
		overlayColor := color.RGBA{0, 0, 0, 180} // Black with transparency
		ebitenutil.DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, overlayColor)
		
		// Draw game over panel
		panelWidth := 500.0
		panelHeight := 300.0
		panelX := (ScreenWidth - panelWidth) / 2
		panelY := (ScreenHeight - panelHeight) / 2
		panelColor := color.RGBA{40, 40, 40, 255} // Dark gray
		ebitenutil.DrawRect(screen, panelX, panelY, panelWidth, panelHeight, panelColor)
		
		// Draw panel border
		borderColor := color.RGBA{255, 255, 255, 255} // White
		borderWidth := 3.0
		// Top border
		ebitenutil.DrawRect(screen, panelX, panelY, panelWidth, borderWidth, borderColor)
		// Bottom border
		ebitenutil.DrawRect(screen, panelX, panelY+panelHeight-borderWidth, panelWidth, borderWidth, borderColor)
		// Left border
		ebitenutil.DrawRect(screen, panelX, panelY, borderWidth, panelHeight, borderColor)
		// Right border
		ebitenutil.DrawRect(screen, panelX+panelWidth-borderWidth, panelY, borderWidth, panelHeight, borderColor)
		
		// Draw game over text and stats
		gameOverText := "GAME OVER"
		scoreText := fmt.Sprintf("Final Score: %d", g.score)
		coinsText := fmt.Sprintf("Coins Collected: %d", g.coinsCollected)
		restartText := "Press SPACE to Restart"
		
		// Calculate text positions (centered in panel)
		textStartX := int(panelX + 50)
		textStartY := int(panelY + 60)
		lineSpacing := 40
		
		ebitenutil.DebugPrintAt(screen, gameOverText, textStartX, textStartY)
		ebitenutil.DebugPrintAt(screen, scoreText, textStartX, textStartY+lineSpacing)
		ebitenutil.DebugPrintAt(screen, coinsText, textStartX, textStartY+lineSpacing*2)
		ebitenutil.DebugPrintAt(screen, restartText, textStartX, textStartY+lineSpacing*4)
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
			passed: false,
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
			passed: false,
		}
		g.obstacles = append(g.obstacles, bottomObs)
	}

	// 3. Spawn coins in the gap between obstacles
	coinSize := float64(16)
	gapTop := gapCenter - gapSize/2
	gapBottom := gapCenter + gapSize/2
	
	// Spawn 2-3 coins randomly in the gap
	numCoins := 2 + rand.Intn(2) // 2 or 3 coins
	for i := 0; i < numCoins; i++ {
		// Random y position within the gap, with some padding
		coinY := gapTop + 20 + rand.Float64()*(gapBottom-gapTop-40)
		coin := &Coin{
			x:        ScreenWidth + obsWidth + 20 + float64(i*40), // Space coins horizontally
			y:        coinY,
			size:     coinSize,
			collected: false,
		}
		g.coins = append(g.coins, coin)
	}
}

