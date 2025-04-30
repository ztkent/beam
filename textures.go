package beam

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Texture struct {
	Name     string
	Rotation float64
	Scale    float64
	OffsetX  float64
	OffsetY  float64
	Tint     rl.Color
}

type AnimatedTexture struct {
	Frames []Texture

	IsAnimated    bool
	AnimationTime float64
	CurrentFrame  int
	lastFrameTime float64
}

func (t *AnimatedTexture) GetCurrentFrame(currentTime float64) Texture {
	if len(t.Frames) == 0 {
		return Texture{Scale: 1.0, Tint: rl.White}
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
