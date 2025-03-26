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
	WaterTile
	WeaponTile
	ChestTile
	FenceTile
	CampfireTile
	BoatTile
	DockTile
	TreeTile
	TreeBaseTile
)

type Map struct {
	Width, Height  int
	Tiles          [][]TileType
	WallTextures   [][]string
	FloorTextures  [][]string
	FloorRotations [][]float64
	Start          Position
	Exit           Position
	Respawn        Position
	DungeonEntry   Position
}
