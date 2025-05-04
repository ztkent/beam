package beam

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Texture struct {
	Name     string
	Rotation float64
	ScaleX   float64
	ScaleY   float64
	OffsetX  float64
	OffsetY  float64
	Tint     rl.Color
}

// Layers for rendering -
// Used to determine the order in which textures are rendered.
// Tiles on the same layer are render top down, left to right.
type Layer int

const (
	BaseLayer Layer = iota
	BackgroundLayer
	ForegroundLayer
)

func (l Layer) String() string {
	switch l {
	case BaseLayer:
		return "Base Layer"
	case BackgroundLayer:
		return "Background Layer"
	case ForegroundLayer:
		return "Foreground Layer"
	default:
		return "Unknown Layer"
	}
}

func OrderedLayers() []Layer {
	return []Layer{
		BackgroundLayer,
		BaseLayer,
		ForegroundLayer,
	}
}

type AnimatedTexture struct {
	Frames []Texture

	IsAnimated    bool
	AnimationTime float64
	CurrentFrame  int
	Layer         Layer

	lastFrameTime float64
}

func (t *AnimatedTexture) GetCurrentFrame(currentTime float64) Texture {
	if len(t.Frames) == 0 {
		return Texture{ScaleX: 1.0, ScaleY: 1.0, Tint: rl.White}
	}
	if t.IsAnimated {
		if len(t.Frames) > 1 {
			if currentTime-t.lastFrameTime >= t.AnimationTime {
				t.CurrentFrame = (t.CurrentFrame + 1) % len(t.Frames)
				t.lastFrameTime = currentTime
			}
			if t.CurrentFrame >= len(t.Frames) {
				t.CurrentFrame = 0
			}
			return t.Frames[t.CurrentFrame]
		} else {
			return t.Frames[0]
		}
	}
	return t.Frames[0]
}
