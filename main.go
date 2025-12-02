package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Main Function ---

func main() {
	game := NewGame()

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("The Migratory Path (Wildlife Game)")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
