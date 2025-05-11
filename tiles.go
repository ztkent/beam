package beam

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type TileType int

func (tile TileType) ToInt() int {
	return int(tile)
}

const (
	WallTile TileType = iota
	FloorTile
	ChestTile
)

type Tile struct {
	Type     TileType
	Pos      Position
	Textures []*AnimatedTexture
}

func NewSimpleTileTexture(name string) *AnimatedTexture {
	return &AnimatedTexture{
		Frames: []Texture{
			{
				Name:     name,
				Rotation: 0,
				ScaleX:   1,
				ScaleY:   1,
				OffsetX:  0,
				OffsetY:  0,
				Tint:     rl.White,
			},
		},
		IsAnimated: false,
	}
}
