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

type Positions []Position
type Position struct {
	X, Y int
}

type TileType int

func (tile TileType) ToInt() int {
	return int(tile)
}

const (
	WallTile TileType = iota
	FloorTile
	EnemyTile
	DungeonEntrance
	DungeonExit
	WeaponTile
	ChestTile
)

type Map struct {
	Width, Height    int
	Tiles            [][]TileType
	Textures         [][][]string
	TextureRotations [][][]float64
	Start            Position
	Exit             Position
	Respawn          Position
	DungeonEntry     Positions
}
