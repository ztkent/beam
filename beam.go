package beam

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

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

type Viewport struct {
	X, Y                    int
	WidthTiles, HeightTiles int
}

type Map struct {
	Width, Height int
	Tiles         [][]Tile
	Start         Position
	Exit          Position
	Respawn       Position
	DungeonEntry  Positions
}

type Tile struct {
	Type     TileType
	Pos      Position
	Textures []TileTexture
}

type TileTexture struct {
	Frames []TileTextureFrame

	// Complex textures can have multiple frames, with a transition
	IsComplex     bool
	AnimationTime float64
	CurrentFrame  int
}

type TileTextureFrame struct {
	Name     string
	Rotation float64
	Scale    float64
	OffsetX  float64
	OffsetY  float64
	Tint     rl.Color
}

func NewSimpleTileTexture(name string) TileTexture {
	return TileTexture{
		Frames: []TileTextureFrame{
			{
				Name:     name,
				Rotation: 0,
				Scale:    1,
				OffsetX:  0,
				OffsetY:  0,
				Tint:     rl.White,
			},
		},
		IsComplex: false,
	}
}
