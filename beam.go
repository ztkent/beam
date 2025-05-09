package beam

type GameState int

const (
	StateMainMenu GameState = iota
	StateGame
	StateReset
	StateContinueGame
	StateGameOver
	StatePaused
	StateQuit
	StateSettings
	StateHighScores
)

type Map struct {
	Width, Height int
	Tiles         [][]Tile
	NPCs          NPCs
	Start         Position
	Exit          Position
	Respawn       Position
	DungeonEntry  Positions
}

type Positions []Position
type Position struct {
	X, Y int
}
type Direction int

const (
	DirRight Direction = iota
	DirLeft
	DirUp
	DirDown
)

type Viewport struct {
	X, Y                    int
	WidthTiles, HeightTiles int
}
