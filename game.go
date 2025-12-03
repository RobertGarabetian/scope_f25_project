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
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/bitmapfont/v4"
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
		baseOffsetX := fx - PlayerX
		baseOffsetY := fy - centerY
		fish[i] = &Fish{
			x:             fx,
			y:             fy,
			offsetX:       baseOffsetX,
			offsetY:       baseOffsetY,
			targetOffsetX: baseOffsetX,
			targetOffsetY: baseOffsetY,
			wanderTimer:   rand.Intn(FishWanderIntervalMax), // Random start time
			wanderInterval: FishWanderIntervalMin + rand.Intn(FishWanderIntervalMax-FishWanderIntervalMin+1), // Random interval for this fish
		}
	}
	
	// Initialize background fish - swimming across the screen at various depths
	backgroundFish := make([]*BackgroundFish, NumBackgroundFish)
	for i := 0; i < NumBackgroundFish; i++ {
		// Random starting position
		startX := rand.Float64() * ScreenWidth
		startY := rand.Float64() * ScreenHeight
		
		// Random direction (1 for right, -1 for left)
		direction := 1
		if rand.Float64() < 0.5 {
			direction = -1
		}
		
		// Random depth (0.3 to 0.7, lower = further back, more faded)
		depth := 0.3 + rand.Float64()*0.4
		
		// Random speed (slower for background fish)
		speed := 0.5 + rand.Float64()*1.0
		
		// Size based on depth (further = smaller)
		size := 30.0 + depth*30.0 // 30-48 pixels
		
		backgroundFish[i] = &BackgroundFish{
			x:         startX,
			y:         startY,
			speed:     speed,
			direction: direction,
			size:      size,
			depth:     depth,
		}
	}
	
	// Initialize bubbles - floating upward at various speeds
	bubbles := make([]*Bubble, NumBubbles)
	for i := 0; i < NumBubbles; i++ {
		// Random starting position
		startX := rand.Float64() * ScreenWidth
		startY := rand.Float64() * ScreenHeight
		
		// Random rising speed (slower bubbles)
		speed := 0.5 + rand.Float64()*1.5 // 0.5 to 2.0 pixels per frame
		
		// Random size (smaller bubbles)
		size := 3.0 + rand.Float64()*8.0 // 3-11 pixels
		
		// Random wobble speed for horizontal movement
		wobbleSpeed := 0.02 + rand.Float64()*0.03 // 0.02 to 0.05
		
		bubbles[i] = &Bubble{
			x:           startX,
			y:           startY,
			speed:       speed,
			size:        size,
			wobble:      rand.Float64() * 2 * math.Pi, // Random starting phase
			wobbleSpeed: wobbleSpeed,
		}
	}
	
	g := &Game{
		// Center the player vertically on the left side
		playerY: centerY,
		obstacles:  make([]*Obstacle, 0),
		coins:      make([]*Coin, 0),
		fish:       fish,
		backgroundFish: backgroundFish,
		bubbles:    bubbles,
		score:      0,
		coinsCollected: 0,
		gameOver:   false,
		gameStarted: false, // Start with difficulty selection
		difficulty: DifficultyNone,
		spawnTimer: 0,
		restartInput: "",
		gameTime:   0,
		speedMultiplier: 1.0,
		fishSprite: createFishSprite(),
		kelpSprite: createKelpSprite(),
	}
	return g
}

// --- Ebitengine Interface Implementations ---

func (g *Game) Update() error {
	// Handle difficulty selection before game starts
	if !g.gameStarted {
		// Check for difficulty selection keys
		if inpututil.IsKeyJustPressed(ebiten.Key1) || inpututil.IsKeyJustPressed(ebiten.KeyE) {
			g.difficulty = DifficultyEasy
			g.gameStarted = true
		} else if inpututil.IsKeyJustPressed(ebiten.Key2) || inpututil.IsKeyJustPressed(ebiten.KeyM) {
			g.difficulty = DifficultyMedium
			g.gameStarted = true
		} else if inpututil.IsKeyJustPressed(ebiten.Key3) || inpututil.IsKeyJustPressed(ebiten.KeyH) {
			g.difficulty = DifficultyHard
			g.gameStarted = true
		}
		return nil
	}
	
	if g.gameOver {
		// Handle text input for restart code "anay"
		// Check for Enter key to submit
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if g.restartInput == "anay" {
				*g = *NewGame()
			} else {
				// Wrong code, clear input
				g.restartInput = ""
			}
			return nil
		}
		
		// Handle backspace
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			if len(g.restartInput) > 0 {
				g.restartInput = g.restartInput[:len(g.restartInput)-1]
			}
			return nil
		}
		
		// Capture typed characters (letters only, lowercase)
		keys := inpututil.AppendJustPressedKeys(nil)
		for _, key := range keys {
			// Convert key to character (a-z only)
			if key >= ebiten.KeyA && key <= ebiten.KeyZ {
				char := string(rune('a' + (key - ebiten.KeyA)))
				g.restartInput += char
				// Limit input length to prevent overflow
				if len(g.restartInput) > 10 {
					g.restartInput = g.restartInput[:10]
				}
			}
		}
		
		return nil
	}

	// 0. Update game time and speed multiplier
	g.gameTime++
	
	// Calculate speed multiplier based on difficulty
	var accelerationRate float64
	switch g.difficulty {
	case DifficultyEasy:
		accelerationRate = 8000.0 // Slow acceleration (increased from 12000)
	case DifficultyMedium:
		accelerationRate = 4000.0 // Medium acceleration (increased from 5000)
	case DifficultyHard:
		accelerationRate = 2000.0 // Fast acceleration (increased from 2500)
	default:
		accelerationRate = 8000.0 // Default to easy
	}
	
	g.speedMultiplier = 1.0 + float64(g.gameTime)/accelerationRate
	// Cap the maximum speed multiplier at 5.0 (5x original speed, increased from 3.0)
	if g.speedMultiplier > 5.0 {
		g.speedMultiplier = 5.0
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

	// 2. Update Fish Positions (following behavior with random wandering)
	for _, fish := range g.fish {
		// Update wander timer and pick new random target when timer expires
		fish.wanderTimer++
		if fish.wanderTimer >= fish.wanderInterval {
			fish.wanderTimer = 0
			
			// Pick a new random wander interval for next time (adds variety to movement)
			fish.wanderInterval = FishWanderIntervalMin + rand.Intn(FishWanderIntervalMax-FishWanderIntervalMin+1)
			
			// Pick a new random target offset within the wander radius
			// Use random angle and distance from base offset
			angle := rand.Float64() * 2 * math.Pi
			radius := FishWanderRadius * math.Sqrt(rand.Float64())
			
			fish.targetOffsetX = fish.offsetX + radius*math.Cos(angle)
			fish.targetOffsetY = fish.offsetY + radius*math.Sin(angle)
		}
		
		// Calculate target position: leader position + fish's target offset
		targetX := PlayerX + fish.targetOffsetX
		targetY := g.playerY + fish.targetOffsetY
		
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

	// 2.5. Update Background Fish Positions
	for _, bgFish := range g.backgroundFish {
		// Move fish in their direction
		bgFish.x += bgFish.speed * float64(bgFish.direction)
		
		// Wrap around when fish goes off screen
		if bgFish.direction > 0 && bgFish.x > ScreenWidth+bgFish.size {
			// Moving right, wrap to left
			bgFish.x = -bgFish.size
			bgFish.y = rand.Float64() * ScreenHeight
		} else if bgFish.direction < 0 && bgFish.x < -bgFish.size {
			// Moving left, wrap to right
			bgFish.x = ScreenWidth + bgFish.size
			bgFish.y = rand.Float64() * ScreenHeight
		}
	}

	// 2.6. Update Bubble Positions
	for _, bubble := range g.bubbles {
		// Move bubble upward
		bubble.y -= bubble.speed
		
		// Apply wobble (horizontal sway)
		bubble.wobble += bubble.wobbleSpeed
		wobbleOffset := math.Sin(bubble.wobble) * 10.0 // Sway 10 pixels left/right
		bubble.x += wobbleOffset * 0.05 // Apply small amount each frame
		
		// Wrap around when bubble goes off top of screen
		if bubble.y < -bubble.size {
			// Reset to bottom with new random x position
			bubble.y = ScreenHeight + bubble.size
			bubble.x = rand.Float64() * ScreenWidth
			bubble.wobble = rand.Float64() * 2 * math.Pi
		}
		
		// Keep bubble within horizontal bounds (with some slack for wobble)
		if bubble.x < -20 {
			bubble.x = ScreenWidth + 20
		} else if bubble.x > ScreenWidth+20 {
			bubble.x = -20
		}
	}

	// 3. Move and Cleanup Obstacles, Update Score
	currentScrollSpeed := ScrollSpeed * g.speedMultiplier
	newObstacles := make([]*Obstacle, 0)
	for _, obs := range g.obstacles {
		obs.x -= currentScrollSpeed // Scroll left with speed multiplier
		
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
		coin.x -= currentScrollSpeed // Scroll left with speed multiplier
		
		// Remove coins that are off-screen or collected
		if coin.x > -coin.size && !coin.collected {
			newCoins = append(newCoins, coin)
		}
	}
	g.coins = newCoins

	// 5. Collision Detection for Leader with Obstacles
	// Use circle-based collision for fish (more accurate than rectangle)
	playerCircle := circleCollision{
		x:      PlayerX + PlayerSize/2,
		y:      g.playerY + PlayerSize/2,
		radius: PlayerSize * 0.35, // Use 40% of size as radius for tighter fit
	}

	for _, obs := range g.obstacles {
		obsRect := collisionRect{
			x: obs.x,
			y: obs.y,
			w: obs.width,
			h: obs.height,
		}
		if checkCircleRectCollision(playerCircle, obsRect) {
			g.gameOver = true
			break
		}
	}

	// 6. Coin Collection Detection for Leader
	if !g.gameOver {
		for _, coin := range g.coins {
			if !coin.collected {
				coinCircle := circleCollision{
					x:      coin.x + coin.size/2,
					y:      coin.y + coin.size/2,
					radius: coin.size * 0.4,
				}
				if checkCircleCollision(playerCircle, coinCircle) {
					coin.collected = true
					g.coinsCollected++
				}
			}
		}
	}

	// 7. Collision Detection for all Fish with Obstacles
	if !g.gameOver {
		for _, fish := range g.fish {
			fishCircle := circleCollision{
				x:      fish.x + FishSize/2,
				y:      fish.y + FishSize/2,
				radius: FishSize * 0.4, // Use 40% of size as radius for tighter fit
			}
			
			for _, obs := range g.obstacles {
				obsRect := collisionRect{
					x: obs.x,
					y: obs.y,
					w: obs.width,
					h: obs.height,
				}
				if checkCircleRectCollision(fishCircle, obsRect) {
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
			fishCircle := circleCollision{
				x:      fish.x + FishSize/2,
				y:      fish.y + FishSize/2,
				radius: FishSize * 0.4,
			}
			
			for _, coin := range g.coins {
				if !coin.collected {
					coinCircle := circleCollision{
						x:      coin.x + coin.size/2,
						y:      coin.y + coin.size/2,
						radius: coin.size * 0.4,
					}
					if checkCircleCollision(fishCircle, coinCircle) {
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

	// Draw Background Fish (drawn first so they appear behind everything)
	for _, bgFish := range g.backgroundFish {
		g.drawBackgroundFish(screen, bgFish)
	}

	// Draw Bubbles (in the background layer)
	for _, bubble := range g.bubbles {
		g.drawBubble(screen, bubble)
	}

	// If game hasn't started, show difficulty selection menu
	if !g.gameStarted {
		g.drawDifficultyMenu(screen)
		return
	}

	// Draw Obstacles (Kelp)
	for _, obs := range g.obstacles {
		g.drawKelp(screen, obs.x, obs.y, obs.width, obs.height)
	}

	// Draw Coins
	for _, coin := range g.coins {
		if !coin.collected {
			g.drawCoin(screen, coin)
		}
	}

	// Draw Player (The Leader)
	g.drawFish(screen, PlayerX, g.playerY, PlayerSize, true)
	
	// Draw all following fish
	for _, fish := range g.fish {
		g.drawFish(screen, fish.x, fish.y, FishSize, false)
	}

	// Draw Score, Coin Count, and Speed (larger text)
	scoreText := fmt.Sprintf("Score: %d", g.score)
	coinText := fmt.Sprintf("Coins: %d", g.coinsCollected)
	speedText := fmt.Sprintf("Speed: %.2fx", g.speedMultiplier)
	
	// Use bitmap font with scaling for larger, more readable text
	statsFace := text.NewGoXFace(bitmapfont.Face)
	statsColor := color.White
	
	// Draw Score
	scoreOpts := &text.DrawOptions{}
	scoreOpts.GeoM.Scale(2.0, 2.0) // 2x scale
	scoreOpts.GeoM.Translate(10, 10)
	scoreOpts.ColorScale.ScaleWithColor(statsColor)
	text.Draw(screen, scoreText, statsFace, scoreOpts)
	
	// Draw Coins
	coinOpts := &text.DrawOptions{}
	coinOpts.GeoM.Scale(2.0, 2.0)
	coinOpts.GeoM.Translate(10, 35)
	coinOpts.ColorScale.ScaleWithColor(statsColor)
	text.Draw(screen, coinText, statsFace, coinOpts)
	
	// Draw Speed
	speedOpts := &text.DrawOptions{}
	speedOpts.GeoM.Scale(2.0, 2.0)
	speedOpts.GeoM.Translate(10, 60)
	speedOpts.ColorScale.ScaleWithColor(statsColor)
	text.Draw(screen, speedText, statsFace, speedOpts)

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
		
		// Draw game over text and stats with larger font
		gameOverText := "GAME OVER"
		scoreText := fmt.Sprintf("Final Score: %d", g.score)
		coinsText := fmt.Sprintf("Coins Collected: %d", g.coinsCollected)
		restartText := "Type 'anay' and press ENTER to restart"
		
		// Calculate text positions (centered in panel)
		textStartX := float64(panelX + 50)
		textStartY := float64(panelY + 60)
		lineSpacing := 60.0
		
		// Use bitmap font with scaling for larger text
		titleFace := text.NewGoXFace(bitmapfont.Face)
		titleOpts := &text.DrawOptions{}
		titleOpts.GeoM.Scale(3.0, 3.0) // Scale up 3x for title
		titleOpts.GeoM.Translate(textStartX, textStartY)
		
		statsFace := text.NewGoXFace(bitmapfont.Face)
		statsOpts := &text.DrawOptions{}
		statsOpts.GeoM.Scale(2.0, 2.0) // Scale up 2x for stats
		statsOpts.GeoM.Translate(textStartX, textStartY+lineSpacing)
		
		restartFace := text.NewGoXFace(bitmapfont.Face)
		restartOpts := &text.DrawOptions{}
		restartOpts.GeoM.Scale(1.5, 1.5) // Scale up 1.5x for restart text
		restartOpts.GeoM.Translate(textStartX, textStartY+lineSpacing*4)
		
		textColor := color.White
		
		// Draw title
		titleOpts.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, gameOverText, titleFace, titleOpts)
		
		// Draw stats
		statsOpts.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, scoreText, statsFace, statsOpts)
		
		statsOpts2 := &text.DrawOptions{}
		statsOpts2.GeoM.Scale(2.0, 2.0)
		statsOpts2.GeoM.Translate(textStartX, textStartY+lineSpacing*2)
		statsOpts2.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, coinsText, statsFace, statsOpts2)
		
		// Draw restart instruction
		restartOpts.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, restartText, restartFace, restartOpts)
		
		// Draw input field
		inputText := "Input: " + g.restartInput + "_"
		inputOpts := &text.DrawOptions{}
		inputOpts.GeoM.Scale(1.5, 1.5)
		inputOpts.GeoM.Translate(textStartX, textStartY+lineSpacing*5)
		inputOpts.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, inputText, restartFace, inputOpts)
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
	obsWidth := float64(80)

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

// drawDifficultyMenu draws the difficulty selection screen
func (g *Game) drawDifficultyMenu(screen *ebiten.Image) {
	// Draw semi-transparent overlay
	overlayColor := color.RGBA{0, 0, 0, 180}
	ebitenutil.DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, overlayColor)
	
	// Draw menu panel
	panelWidth := 600.0
	panelHeight := 400.0
	panelX := (ScreenWidth - panelWidth) / 2
	panelY := (ScreenHeight - panelHeight) / 2
	panelColor := color.RGBA{40, 40, 40, 255}
	ebitenutil.DrawRect(screen, panelX, panelY, panelWidth, panelHeight, panelColor)
	
	// Draw panel border
	borderColor := color.RGBA{255, 255, 255, 255}
	borderWidth := 3.0
	ebitenutil.DrawRect(screen, panelX, panelY, panelWidth, borderWidth, borderColor)
	ebitenutil.DrawRect(screen, panelX, panelY+panelHeight-borderWidth, panelWidth, borderWidth, borderColor)
	ebitenutil.DrawRect(screen, panelX, panelY, borderWidth, panelHeight, borderColor)
	ebitenutil.DrawRect(screen, panelX+panelWidth-borderWidth, panelY, borderWidth, panelHeight, borderColor)
	
	// Create text face
	textFace := text.NewGoXFace(bitmapfont.Face)
	textColor := color.White
	
	// Draw title
	titleText := "SELECT DIFFICULTY"
	titleOpts := &text.DrawOptions{}
	titleOpts.GeoM.Scale(3.0, 3.0)
	titleOpts.GeoM.Translate(panelX+100, panelY+50)
	titleOpts.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, titleText, textFace, titleOpts)
	
	// Draw difficulty options
	easyText := "1 or E - EASY"
	easySubText := "Slow acceleration (8000 frames)"
	mediumText := "2 or M - MEDIUM"
	mediumSubText := "Medium acceleration (4000 frames)"
	hardText := "3 or H - HARD"
	hardSubText := "Fast acceleration (2000 frames)"
	
	yOffset := panelY + 140
	lineSpacing := 60.0
	
	// Easy option
	easyOpts := &text.DrawOptions{}
	easyOpts.GeoM.Scale(2.5, 2.5)
	easyOpts.GeoM.Translate(panelX+80, yOffset)
	easyOpts.ColorScale.ScaleWithColor(color.RGBA{100, 255, 100, 255}) // Green
	text.Draw(screen, easyText, textFace, easyOpts)
	
	easySubOpts := &text.DrawOptions{}
	easySubOpts.GeoM.Scale(1.5, 1.5)
	easySubOpts.GeoM.Translate(panelX+120, yOffset+30)
	easySubOpts.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255}) // Gray
	text.Draw(screen, easySubText, textFace, easySubOpts)
	
	// Medium option
	mediumOpts := &text.DrawOptions{}
	mediumOpts.GeoM.Scale(2.5, 2.5)
	mediumOpts.GeoM.Translate(panelX+80, yOffset+lineSpacing*1.5)
	mediumOpts.ColorScale.ScaleWithColor(color.RGBA{255, 255, 100, 255}) // Yellow
	text.Draw(screen, mediumText, textFace, mediumOpts)
	
	mediumSubOpts := &text.DrawOptions{}
	mediumSubOpts.GeoM.Scale(1.5, 1.5)
	mediumSubOpts.GeoM.Translate(panelX+120, yOffset+lineSpacing*1.5+30)
	mediumSubOpts.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
	text.Draw(screen, mediumSubText, textFace, mediumSubOpts)
	
	// Hard option
	hardOpts := &text.DrawOptions{}
	hardOpts.GeoM.Scale(2.5, 2.5)
	hardOpts.GeoM.Translate(panelX+80, yOffset+lineSpacing*3)
	hardOpts.ColorScale.ScaleWithColor(color.RGBA{255, 100, 100, 255}) // Red
	text.Draw(screen, hardText, textFace, hardOpts)
	
	hardSubOpts := &text.DrawOptions{}
	hardSubOpts.GeoM.Scale(1.5, 1.5)
	hardSubOpts.GeoM.Translate(panelX+120, yOffset+lineSpacing*3+30)
	hardSubOpts.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
	text.Draw(screen, hardSubText, textFace, hardSubOpts)
}

