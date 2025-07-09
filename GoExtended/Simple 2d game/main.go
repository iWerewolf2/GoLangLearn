package main

import (
	"image/color" // Import for colors
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector" // Import for drawing shapes
)

// Direction represents the direction of the snake.
type Direction int

const (
	DirNone Direction = iota // No direction (initial state)
	DirUp
	DirDown
	DirLeft
	DirRight
)

// Game implements the ebiten.Game interface
type Game struct {
	snakeX       int       // Snake's head X position
	snakeY       int       // Snake's head Y position
	direction    Direction // Current direction of the snake
	screenWidth  int
	screenHeight int
}

// NewGame creates and initializes a new Game instance.
func NewGame() *Game {
	return &Game{
		snakeX:       10,       // Starting position
		snakeY:       10,       // Starting position
		direction:    DirRight, // Start moving right
		screenWidth:  320,
		screenHeight: 240,
	}
}

// Update updates the game logic.
func (g *Game) Update() error {
	// Handle input for changing direction
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		if g.direction != DirDown { // Prevent immediately reversing
			g.direction = DirUp
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		if g.direction != DirUp { // Prevent immediately reversing
			g.direction = DirDown
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		if g.direction != DirRight { // Prevent immediately reversing
			g.direction = DirLeft
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		if g.direction != DirLeft { // Prevent immediately reversing
			g.direction = DirRight
		}
	}

	// For now, let's just update the snake's position based on direction.
	// In a real game, this would be tied to a game tick rate.
	switch g.direction {
	case DirUp:
		g.snakeY--
	case DirDown:
		g.snakeY++
	case DirLeft:
		g.snakeX--
	case DirRight:
		g.snakeX++
	}

	// Basic boundary check (optional, but good for testing)
	if g.snakeX < 0 {
		g.snakeX = 0
	} else if g.snakeX >= g.screenWidth {
		g.snakeX = g.screenWidth - 1
	}
	if g.snakeY < 0 {
		g.snakeY = 0
	} else if g.snakeY >= g.screenHeight {
		g.snakeY = g.screenHeight - 1
	}

	return nil
}

// Draw draws the game world.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen (optional, but good practice)
	screen.Fill(color.Black) // Fill background with black

	// Draw the snake (for now, just a square representing the head)
	// We'll use vector.DrawRect for simplicity.
	// Arguments: image, x, y, width, height, color
	vector.DrawRect(screen, float32(g.snakeX), float32(g.snakeY), 10, 10, color.RGBA{0x00, 0xFF, 0x00, 0xFF}, false) // Green square
}

// Layout defines the size of the game window.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.screenWidth, g.screenHeight // Return desired window size
}

func main() {
	ebiten.SetWindowTitle("Snake Game")
	ebiten.SetWindowSize(640, 480) // Increase visible window size

	// Create an instance of our game
	game := NewGame() // Use the NewGame constructor

	// Run the game loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
