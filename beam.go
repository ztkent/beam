package beam

import "image/color"

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

type TileTexture struct {
	Name     string
	Rotation float64
	Scale    float64
	OffsetX  float64
	OffsetY  float64
	Tint     color.RGBA
}

type Tile struct {
	Type     TileType
	Pos      Position
	Textures []TileTexture
}

type Map struct {
	Width, Height int
	Tiles         [][]Tile
	Start         Position
	Exit          Position
	Respawn       Position
	DungeonEntry  Positions
}

type Viewport struct {
	X, Y                    int
	WidthTiles, HeightTiles int
}
