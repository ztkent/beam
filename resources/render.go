package resources

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
)

func (rm *ResourceManager) RenderTexture(texture *beam.AnimatedTexture, pos rl.Rectangle, tileSize int) {
	if texture == nil {
		return
	}

	if !texture.IsAnimated {
		for _, frame := range texture.Frames {
			origin := rl.Vector2{
				X: float32(tileSize) / 2,
				Y: float32(tileSize) / 2,
			}
			info, err := rm.GetTexture("default", frame.Name)
			if err != nil {
				fmt.Println("Error getting texture:", err)
				return
			}
			destRect := rl.Rectangle{
				X:      pos.X + pos.Width/2 + float32(frame.OffsetX*float64(tileSize)),
				Y:      pos.Y + pos.Height/2 + float32(frame.OffsetY*float64(tileSize)),
				Width:  pos.Width * float32(frame.ScaleX),
				Height: pos.Height * float32(frame.ScaleY),
			}
			if frame.Tint == (rl.Color{}) {
				frame.Tint = rl.White
			}

			rl.DrawTexturePro(
				info.Texture,
				info.Region,
				destRect,
				origin,
				float32(frame.Rotation),
				frame.Tint,
			)
		}
	} else {
		// Render complex textures
		frame := texture.GetCurrentFrame(rl.GetTime())
		origin := rl.Vector2{
			X: float32(tileSize) / 2,
			Y: float32(tileSize) / 2,
		}
		info, err := rm.GetTexture("default", frame.Name)
		if err != nil {
			fmt.Println("Error getting texture:", err)
			return
		}
		destRect := rl.Rectangle{
			X:      pos.X + pos.Width/2 + float32(frame.OffsetX*float64(tileSize)),
			Y:      pos.Y + pos.Height/2 + float32(frame.OffsetY*float64(tileSize)),
			Width:  pos.Width * float32(frame.ScaleX),
			Height: pos.Height * float32(frame.ScaleY),
		}
		if frame.Tint == (rl.Color{}) {
			frame.Tint = rl.White
		}
		rl.DrawTexturePro(
			info.Texture,
			info.Region,
			destRect,
			origin,
			float32(frame.Rotation),
			frame.Tint,
		)
	}
}

func (rm *ResourceManager) RenderNPC(npc *beam.NPC, pos rl.Rectangle, tileSize int) {
	if npc.Data.Dead {
		// Calculate alpha based on dying frames (fade out over 32 frames)
		totalDyingFrames := 32
		alpha := uint8(255 * (1.0 - (float32(npc.Data.DyingFrames) / float32(totalDyingFrames))))
		fadeColor := rl.NewColor(255, 255, 255, alpha)
		rm.RenderTexture(&beam.AnimatedTexture{
			Frames: []beam.Texture{
				{
					Name:     npc.GetCurrentTexture().GetCurrentFrame(rl.GetTime()).Name,
					Rotation: 0,
					ScaleX:   1,
					ScaleY:   1,
					OffsetX:  0,
					OffsetY:  0,
					Tint:     fadeColor,
				},
			},
			IsAnimated: false,
		}, pos, tileSize)

	} else if npc.Data.TookDamageThisFrame {
		// Calculate damage flash alpha
		const totalDamageFrames = 32.0
		const peakAlpha = 0.8
		progress := float32(npc.Data.DamageFrames) / totalDamageFrames

		// Start at peak and fade out using cosine for smooth transition
		alpha := peakAlpha * float32(math.Cos(float64(progress)*math.Pi/2))
		damageColor := rl.NewColor(255, 0, 0, uint8(255*alpha)) // Bright red with fade

		// Render both the enemy and the damage overlay
		rm.RenderTexture(npc.Data.Texture.Up, pos, tileSize)
		rm.RenderTexture(&beam.AnimatedTexture{
			Frames: []beam.Texture{
				{
					Name:     npc.GetCurrentTexture().GetCurrentFrame(rl.GetTime()).Name,
					Rotation: 0,
					ScaleX:   1,
					ScaleY:   1,
					OffsetX:  0,
					OffsetY:  0,
					Tint:     damageColor,
				},
			},
			Layer:      npc.GetCurrentTexture().Layer,
			IsAnimated: false,
		}, pos, tileSize)

		if npc.Data.DamageFrames >= int(totalDamageFrames) {
			npc.Data.DamageFrames = 0
			npc.Data.TookDamageThisFrame = false
		}
	} else {
		rm.RenderTexture(beam.NewSimpleTileTexture(npc.GetCurrentTexture().GetCurrentFrame(rl.GetTime()).Name), pos, tileSize)
	}

	// Only show health bar for 5 seconds after health changes
	if !npc.Data.Dead {
		currentTime := float32(rl.GetTime())
		if npc.Data.LastHealthChange != 0 && currentTime-npc.Data.LastHealthChange < 5.0 {
			// Draw health bar
			barWidth := float32(tileSize)
			barHeight := float32(4)
			healthPercent := float32(npc.Data.Health) / float32(npc.Data.MaxHealth)

			// Background (gray)
			rl.DrawRectangle(
				int32(pos.X),
				int32(pos.Y-barHeight-2),
				int32(barWidth),
				int32(barHeight),
				rl.Gray,
			)

			// Health bar (green to red based on health percentage)
			barColor := rl.ColorFromHSV(120.0*healthPercent, 1.0, 1.0)
			rl.DrawRectangle(
				int32(pos.X),
				int32(pos.Y-barHeight-2),
				int32(barWidth*healthPercent),
				int32(barHeight),
				barColor,
			)
		}
	}
}
