package main

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// --- Sprite Creation Functions ---

// Helper function to set a pixel in an image
func setPixel(img *image.RGBA, x, y int, c color.Color) {
	img.Set(x, y, c)
}

// loadImageFromFile loads an image from a file path
func loadImageFromFile(path string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// createFishSprite loads the fish sprite from a PNG file
func createFishSprite() *ebiten.Image {
	img, err := loadImageFromFile("assets/fish.png")
	if err != nil {
		panic("Failed to load fish.png: " + err.Error())
	}
	return img
}

// createKelpSprite creates a pixel art kelp sprite (vertical strip)
func createKelpSprite() *ebiten.Image {
	// Create a 8x32 pixel art kelp strip (will be tiled/scaled)
	width := 8
	height := 32
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Define colors
	transparent := color.RGBA{0, 0, 0, 0}
	kelpDark := color.RGBA{0, 80, 0, 255}      // Dark green
	kelpMedium := color.RGBA{0, 120, 0, 255}   // Medium green
	kelpLight := color.RGBA{0, 160, 0, 255}    // Light green
	kelpAccent := color.RGBA{20, 100, 20, 255} // Accent green
	
	// Create a wavy kelp pattern (vertical)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a wavy pattern
			waveOffset := int(math.Sin(float64(y)*0.3) * 1.5)
			centerX := width/2 + waveOffset
			
			if x == centerX || x == centerX-1 || x == centerX+1 {
				// Main stem - darker in center
				if x == centerX {
					setPixel(img, x, y, kelpDark)
				} else {
					setPixel(img, x, y, kelpMedium)
				}
			} else if x == centerX-2 || x == centerX+2 {
				// Outer edge - lighter
				setPixel(img, x, y, kelpLight)
			} else {
				setPixel(img, x, y, transparent)
			}
			
			// Add some texture variation
			if (x+y)%3 == 0 && (x == centerX-1 || x == centerX || x == centerX+1) {
				setPixel(img, x, y, kelpAccent)
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// drawFish draws a fish sprite at the given position
func (g *Game) drawFish(screen *ebiten.Image, x, y, size float64, isLeader bool) {
	op := &ebiten.DrawImageOptions{}
	
	// Get the sprite dimensions for proper scaling
	spriteW, _ := g.fishSprite.Size()
	// Scale the sprite to the desired size (assuming square sprites)
	scale := size / float64(spriteW)
	op.GeoM.Scale(scale, scale)
	
	// Position
	op.GeoM.Translate(x, y)
	
	// Use different tint for leader vs followers
	if isLeader {
		// Leader is brighter - scale RGB values
		op.ColorM.Scale(1.2, 1.2, 1.2, 1.0)
	} else {
		// Followers are slightly lighter blue
		op.ColorM.Scale(0.9, 1.0, 1.1, 1.0)
	}
	
	screen.DrawImage(g.fishSprite, op)
}

// drawKelp draws kelp by tiling the kelp sprite vertically with wave animation
func (g *Game) drawKelp(screen *ebiten.Image, x, y, width, height float64) {
	kelpTileHeight := 32.0 // Height of one kelp tile
	tiles := int(height / kelpTileHeight) + 1
	
	for i := 0; i < tiles; i++ {
		tileY := y + float64(i)*kelpTileHeight
		tileHeight := kelpTileHeight
		if tileY+tileHeight > y+height {
			tileHeight = (y + height) - tileY
		}
		
		// Calculate wave offset based on time and position
		// Different kelp plants wave at different speeds based on their x position
		timeOffset := float64(g.gameTime) * 0.05 // Animation speed
		positionOffset := x * 0.01 // Offset based on x position for variety
		yOffset := tileY * 0.015 // More wave at the top
		
		// Create a smooth wave motion using sine
		waveAmplitude := 3.0 + (tileY-y)/height*8.0 // Stronger wave at top of kelp
		waveX := math.Sin(timeOffset+positionOffset+yOffset) * waveAmplitude
		
		op := &ebiten.DrawImageOptions{}
		// Scale to match width and tile height
		scaleX := width / 8.0
		scaleY := tileHeight / kelpTileHeight
		op.GeoM.Scale(scaleX, scaleY)
		
		// Apply wave offset and translate to position
		op.GeoM.Translate(x+waveX, tileY)
		
		screen.DrawImage(g.kelpSprite, op)
	}
}

