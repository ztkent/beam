package beam

import (
	"math"
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	beam_math "github.com/ztkent/beam/math"
)

type NPCTexture struct {
	Up    *AnimatedTexture
	Down  *AnimatedTexture
	Left  *AnimatedTexture
	Right *AnimatedTexture
}

type NPC struct {
	Pos            Position
	LastMoveTime   float32
	LastActionTime time.Time
	Data           NPCData
}

type NPCData struct {
	Name     string
	Texture  *NPCTexture
	SpawnPos Position

	Health           int
	MaxHealth        int
	LastHealthChange float32
	LastAttackTime   time.Time
	Attack           int
	BaseAttack       int
	Defense          int
	BaseDefense      int
	AttackSpeed      float64
	BaseAttackSpeed  float64
	AttackRange      float64
	BaseAttackRange  float64
	MoveSpeed        float64
	Direction        Direction

	Attackable          bool
	Hostile             bool
	AggroRange          int
	AttackState         int
	AttackStateTime     float32
	TookDamageThisFrame bool
	DamageFrames        int
	DyingFrames         int
	Dead                bool
}

func NewSimpleNPCTexture(name string) *NPCTexture {
	return &NPCTexture{
		Up: &AnimatedTexture{
			Frames: []Texture{{Name: name, Rotation: 0, ScaleX: 1, ScaleY: 1, OffsetX: 0, OffsetY: 0, Tint: rl.White}},
		},
		Down: &AnimatedTexture{
			Frames: []Texture{{Name: name, Rotation: 0, ScaleX: 1, ScaleY: 1, OffsetX: 0, OffsetY: 0, Tint: rl.White}},
		},
		Left: &AnimatedTexture{
			Frames: []Texture{{Name: name, Rotation: 0, ScaleX: 1, ScaleY: 1, OffsetX: 0, OffsetY: 0, Tint: rl.White}},
		},
		Right: &AnimatedTexture{
			Frames: []Texture{{Name: name, Rotation: 0, ScaleX: 1, ScaleY: 1, OffsetX: 0, OffsetY: 0, Tint: rl.White}},
		},
	}
}

// Move the NPC towards the player if within aggro range
// or move randomly if not. The NPC will also check for obstacles.
func (npc *NPC) Update(playerPos Position, tiles [][]Tile) (died bool) {
	if npc.Data.Dead {
		totalDyingFrames := 32
		npc.Data.DyingFrames++
		if npc.Data.DyingFrames >= totalDyingFrames {
			return true
		}
	} else if npc.Data.TookDamageThisFrame {
		totalDamageFrames := 32
		npc.Data.DamageFrames++
		if npc.Data.DamageFrames == 1 {
			npc.knockback(tiles, 1)
		}
		if npc.Data.DamageFrames >= int(totalDamageFrames) {
			npc.Data.DamageFrames = 0
			npc.Data.TookDamageThisFrame = false
		}
	}

	currentTime := float32(rl.GetTime())
	if npc.LastMoveTime < currentTime && (currentTime-npc.LastMoveTime < 1.0-(float32(npc.Data.MoveSpeed-1)*0.1)) {
		return
	}

	// Calculate distance to player
	dist := beam_math.ManhattanDistance(npc.Pos.X, npc.Pos.Y, playerPos.X, playerPos.Y)
	var dx, dy int

	if dist == 0 {
		directions := Positions{
			{X: 0, Y: -1}, // North
			{X: 1, Y: 0},  // East
			{X: 0, Y: 1},  // South
			{X: -1, Y: 0}, // West
		}
		for _, dir := range directions {
			newX := npc.Pos.X + dir.X
			newY := npc.Pos.Y + dir.Y
			if newX > 0 && newX < len(tiles[0])-1 &&
				newY > 0 && newY < len(tiles)-1 &&
				tiles[newY][newX].Type == FloorTile {
				dx, dy = dir.X, dir.Y
				break
			}
		}
	} else if dist <= npc.Data.AggroRange && npc.Data.Hostile {
		isDiagonal := npc.Pos.X != playerPos.X && npc.Pos.Y != playerPos.Y
		xDiff := playerPos.X - npc.Pos.X
		yDiff := playerPos.Y - npc.Pos.Y

		if isDiagonal && dist > 1 {
			if math.Abs(float64(xDiff)) >= math.Abs(float64(yDiff)) {
				dx = beam_math.Sign(xDiff)
				dy = 0
				newX := npc.Pos.X + dx
				if newX <= 0 || newX >= len(tiles[0])-1 ||
					tiles[npc.Pos.Y][newX].Type != FloorTile {
					dx = 0
					dy = beam_math.Sign(yDiff)
				}
			} else {
				dy = beam_math.Sign(yDiff)
				dx = 0
				newY := npc.Pos.Y + dy
				if newY <= 0 || newY >= len(tiles)-1 ||
					tiles[newY][npc.Pos.X].Type != FloorTile {
					dy = 0
					dx = beam_math.Sign(xDiff)
				}
			}
		} else if dist > 1 {
			if npc.Pos.X < playerPos.X {
				dx = 1
			} else if npc.Pos.X > playerPos.X {
				dx = -1
			}
			if npc.Pos.Y < playerPos.Y {
				dy = 1
			} else if npc.Pos.Y > playerPos.Y {
				dy = -1
			}
		}

		newDist := beam_math.ManhattanDistance(npc.Pos.X+dx, npc.Pos.Y+dy, playerPos.X, playerPos.Y)
		if newDist < 1 {
			dx, dy = 0, 0
		}
	} else {
		if rand.Float32() < 0.75 {
			directions := Positions{
				{X: 0, Y: -1},
				{X: 1, Y: 0},
				{X: 0, Y: 1},
				{X: -1, Y: 0},
			}
			dir := directions[rand.Intn(len(directions))]
			dx, dy = dir.X, dir.Y

			newDist := beam_math.ManhattanDistance(npc.Pos.X+dx, npc.Pos.Y+dy, playerPos.X, playerPos.Y)
			if newDist < 1 {
				dx, dy = 0, 0
			}
		}
	}

	newX := npc.Pos.X + dx
	newY := npc.Pos.Y + dy
	if newX > 0 && newX < len(tiles[0])-1 &&
		newY > 0 && newY < len(tiles)-1 &&
		tiles[newY][newX].Type == FloorTile {
		npc.Pos.X = newX
		npc.Pos.Y = newY
	}

	if dx > 0 {
		npc.Data.Direction = DirRight
	} else if dx < 0 {
		npc.Data.Direction = DirLeft
	} else if dy > 0 {
		npc.Data.Direction = DirDown
	} else if dy < 0 {
		npc.Data.Direction = DirUp
	}

	npc.LastMoveTime = currentTime
	return false
}

// Attack the player if within attack range and the NPC is hostile.
func (npc *NPC) Attack(playerPos Position) (hit bool) {
	if !npc.Data.Hostile || npc.Data.Dead {
		return false
	}
	dist := beam_math.ManhattanDistance(npc.Pos.X, npc.Pos.Y, playerPos.X, playerPos.Y)
	if dist <= int(math.Round(npc.Data.AttackRange)) {
		if time.Since(npc.LastActionTime) > time.Duration(npc.Data.AttackSpeed*1000)*time.Millisecond {
			npc.LastActionTime = time.Now()
			return true
		}
	}
	return false
}

func (npc *NPC) GetCurrentTexture() *AnimatedTexture {
	switch npc.Data.Direction {
	case DirUp:
		return npc.Data.Texture.Up
	case DirDown:
		return npc.Data.Texture.Down
	case DirLeft:
		return npc.Data.Texture.Left
	case DirRight:
		return npc.Data.Texture.Right
	default:
		return nil
	}
}

// Knockback the NPC in the opposite direction theyre facing
func (npc *NPC) knockback(tiles [][]Tile, dist int) {
	height := len(tiles)
	width := 0
	if height > 0 {
		width = len(tiles[0])
	}

	// Store initial beam.Position
	tempX := npc.Pos.X
	tempY := npc.Pos.Y

	// Calculate target beam.Position based on direction
	switch npc.Data.Direction {
	case DirRight:
		// Check each tile in the knockback path
		for i := 1; i <= dist; i++ {
			nextX := tempX - i
			// Stop if we hit bounds or a wall
			if nextX < 0 || tiles[tempY][nextX].Type != FloorTile {
				break
			}
			tempX = nextX
		}
	case DirLeft:
		for i := 1; i <= dist; i++ {
			nextX := tempX + i
			if nextX >= width || tiles[tempY][nextX].Type != FloorTile {
				break
			}
			tempX = nextX
		}
	case DirUp:
		for i := 1; i <= dist; i++ {
			nextY := tempY + i
			if nextY >= height || tiles[nextY][tempX].Type != FloorTile {
				break
			}
			tempY = nextY
		}
	case DirDown:
		for i := 1; i <= dist; i++ {
			nextY := tempY - i
			if nextY < 0 || tiles[nextY][tempX].Type != FloorTile {
				break
			}
			tempY = nextY
		}
	}

	// Update enemy beam.Position to final valid location
	tiles[npc.Pos.Y][npc.Pos.X].Type = FloorTile
	npc.Pos.X = tempX
	npc.Pos.Y = tempY
}
