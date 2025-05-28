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
	Items         Items
	Start         Position
	Exit          Positions
	Respawn       Position
	DungeonEntry  Positions
}

type Positions []Position
type Position struct {
	X, Y int
}

func (p Positions) PositionExists(pos Position) bool {
	for _, p := range p {
		if p.X == pos.X && p.Y == pos.Y {
			return true
		}
	}
	return false
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
