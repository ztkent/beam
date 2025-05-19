package beam

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam/chat"
	beam_math "github.com/ztkent/beam/math"
)

/*
The NPC system supports:
  - Wandering NPCs with customizable behavior
  - Combat-enabled NPCs with stats and attack patterns
  - Interactive NPCs with dialog system
  - Customizable textures for different directions
  - Collision detection and pathfinding
  - Health and damage system
  - Aggro ranges and wander limits

Example usage:
    npc := &NPC{
        Pos: Position{X: 5, Y: 5},
        Data: NPCData{
            Name: "Guard",
            Texture: NewSimpleNPCTexture("guard"),
            Health: 100,
            MaxHealth: 100,
            Attack: 10,
            Defense: 5,
            Hostile: true,
            AggroRange: 5,
            WanderRange: 3,
            Impassable: true,
        },
    }
*/

type NPCTexture struct {
	Idle  *AnimatedTexture
	Up    *AnimatedTexture
	Down  *AnimatedTexture
	Left  *AnimatedTexture
	Right *AnimatedTexture
}

type NPCs []*NPC

func (npcs NPCs) IsBlocked(x, y int) bool {
	for _, npc := range npcs {
		if !npc.Data.Dead &&
			npc.Data.Impassable &&
			npc.Pos.X == x && npc.Pos.Y == y {
			return true
		}
	}
	return false
}

func (npcs NPCs) LivingNPCs() NPCs {
	targets := make(NPCs, 0)
	for _, e := range npcs {
		if !e.Data.Dead {
			targets = append(targets, e)
		}
	}
	return targets
}

func (npcs NPCs) IsInteracting() bool {
	for _, e := range npcs {
		if e.Data.IsInteracting {
			return true
		}
	}
	return false
}

func (npcs NPCs) InteractableNearby(playerPos Position) NPCs {
	targets := make(NPCs, 0)
	for _, e := range npcs {
		if e.Data.Interactable && !e.Data.Dead && !e.Data.IsInteracting {
			dist := beam_math.ManhattanDistance(e.Pos.X, e.Pos.Y, playerPos.X, playerPos.Y)
			if dist <= 1 {
				targets = append(targets, e)
			}
		}
	}
	return targets
}

type NPC struct {
	Pos         Position
	Data        NPCData
	CurrentChat *chat.Chat
}

type NPCData struct {
	Name     string
	Texture  *NPCTexture
	SpawnPos Position

	LastMoveTime     float32
	LastHealthChange float32
	LastAttackTime   float32

	Health          int
	MaxHealth       int
	Attack          int
	BaseAttack      int
	Defense         int
	BaseDefense     int
	AttackSpeed     float64
	BaseAttackSpeed float64
	AttackRange     float64
	BaseAttackRange float64
	MoveSpeed       float64
	Direction       Direction

	Attackable          bool
	Impassable          bool
	Hostile             bool
	WanderRange         int
	AggroRange          int
	AttackState         int
	AttackStateTime     float32
	TookDamageThisFrame bool
	DamageFrames        int
	DyingFrames         int
	Dead                bool

	Interactable  bool
	IsInteracting bool
	Experience    int
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

// Run the NPC update loop.
func (npc *NPC) Update(playerPos Position, currMap *Map) (died bool) {
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
			npc.knockback(playerPos, currMap.Tiles, 1)
		}
		if npc.Data.DamageFrames >= int(totalDamageFrames) {
			npc.Data.DamageFrames = 0
			npc.Data.TookDamageThisFrame = false
		}
	}

	if npc.Data.IsInteracting {
		if npc.CurrentChat.State == chat.DialogFinished || npc.CurrentChat.State == chat.DialogHidden {
			npc.Data.IsInteracting = false
		}
		npc.CurrentChat.Update()
		npc.CurrentChat.Draw()
		return false
	}

	npc.Wander(playerPos, currMap)
	return false
}

// Interact with a player
func (npc *NPC) Interact(playerPos Position, currChat *chat.Chat) {
	if !npc.Data.Interactable {
		return
	}

	dist := beam_math.ManhattanDistance(npc.Pos.X, npc.Pos.Y, playerPos.X, playerPos.Y)
	if dist <= 1 {
		npc.Data.IsInteracting = true
		if currChat == nil {
			npc.CurrentChat = chat.NewChat()
		} else {
			npc.CurrentChat = currChat
		}
	} else {
		npc.Data.IsInteracting = false
		return
	}

	// Turn and face the player
	if playerPos.X > npc.Pos.X {
		npc.Data.Direction = DirRight
	}
	if playerPos.X < npc.Pos.X {
		npc.Data.Direction = DirLeft
	}
	if playerPos.Y > npc.Pos.Y {
		npc.Data.Direction = DirDown
	}
	if playerPos.Y < npc.Pos.Y {
		npc.Data.Direction = DirUp
	}

	// Start the chat
	npc.CurrentChat.Show()
	return
}

// A simple wandering algo that moves the NPC towards the player if within aggro range.
// If not, it will wander randomly. The NPC will also check for obstacles.
// The NPC will try to stay within its wander range, if possible.
func (npc *NPC) Wander(playerPos Position, currMap *Map) {
	currentTime := float32(rl.GetTime())
	if npc.Data.MoveSpeed <= 0 || ((currentTime - npc.Data.LastMoveTime) < 1.0/float32(npc.Data.MoveSpeed)) {
		return
	}

	// Calculate distance to player
	distToPlayer := beam_math.ManhattanDistance(npc.Pos.X, npc.Pos.Y, playerPos.X, playerPos.Y)
	distToSpawn := beam_math.ManhattanDistance(npc.Pos.X, npc.Pos.Y, npc.Data.SpawnPos.X, npc.Data.SpawnPos.Y)
	var dx, dy int

	if distToPlayer == 0 {
		directions := Positions{
			{X: 0, Y: -1}, // North
			{X: 1, Y: 0},  // East
			{X: 0, Y: 1},  // South
			{X: -1, Y: 0}, // West
		}
		for _, dir := range directions {
			newX := npc.Pos.X + dir.X
			newY := npc.Pos.Y + dir.Y
			if newX > 0 && newX < len(currMap.Tiles[0])-1 &&
				newY > 0 && newY < len(currMap.Tiles)-1 &&
				currMap.Tiles[newY][newX].Type == FloorTile {
				dx, dy = dir.X, dir.Y
				break
			}
		}
	} else if distToPlayer <= npc.Data.AggroRange && npc.Data.Hostile {
		isDiagonal := npc.Pos.X != playerPos.X && npc.Pos.Y != playerPos.Y
		xDiff := playerPos.X - npc.Pos.X
		yDiff := playerPos.Y - npc.Pos.Y

		if isDiagonal && distToPlayer > 1 {
			if math.Abs(float64(xDiff)) >= math.Abs(float64(yDiff)) {
				dx = beam_math.Sign(xDiff)
				dy = 0
				newX := npc.Pos.X + dx
				if newX <= 0 || newX >= len(currMap.Tiles[0])-1 ||
					currMap.Tiles[npc.Pos.Y][newX].Type != FloorTile {
					dx = 0
					dy = beam_math.Sign(yDiff)
				}
			} else {
				dy = beam_math.Sign(yDiff)
				dx = 0
				newY := npc.Pos.Y + dy
				if newY <= 0 || newY >= len(currMap.Tiles)-1 ||
					currMap.Tiles[newY][npc.Pos.X].Type != FloorTile {
					dy = 0
					dx = beam_math.Sign(xDiff)
				}
			}
		} else if distToPlayer > 1 {
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
			// If we're beyond wander range, try to move back toward spawn point
			if npc.Data.WanderRange > 0 && distToSpawn >= npc.Data.WanderRange {
				xDiff := npc.Data.SpawnPos.X - npc.Pos.X
				yDiff := npc.Data.SpawnPos.Y - npc.Pos.Y

				if math.Abs(float64(xDiff)) >= math.Abs(float64(yDiff)) {
					dx = beam_math.Sign(xDiff)
					dy = 0
				} else {
					dx = 0
					dy = beam_math.Sign(yDiff)
				}
			} else {
				directions := Positions{
					{X: 0, Y: -1},
					{X: 1, Y: 0},
					{X: 0, Y: 1},
					{X: -1, Y: 0},
				}
				dir := directions[rand.Intn(len(directions))]
				dx, dy = dir.X, dir.Y

				// Check if new position would exceed wander range
				if npc.Data.WanderRange > 0 {
					newDistToSpawn := beam_math.ManhattanDistance(npc.Pos.X+dx, npc.Pos.Y+dy, npc.Data.SpawnPos.X, npc.Data.SpawnPos.Y)
					if newDistToSpawn > npc.Data.WanderRange {
						dx, dy = 0, 0
					}
				}
			}

			newDist := beam_math.ManhattanDistance(npc.Pos.X+dx, npc.Pos.Y+dy, playerPos.X, playerPos.Y)
			if newDist < 1 {
				dx, dy = 0, 0
			}
		}
	}

	newX := npc.Pos.X + dx
	newY := npc.Pos.Y + dy
	if newX > 0 && newX < len(currMap.Tiles[0])-1 &&
		newY > 0 && newY < len(currMap.Tiles)-1 &&
		currMap.Tiles[newY][newX].Type != WallTile &&
		currMap.Tiles[newY][newX].Type != ChestTile &&
		!currMap.NPCs.IsBlocked(newX, newY) &&
		!currMap.Items.IsBlocked(newX, newY) {
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
	npc.Data.LastMoveTime = currentTime
}

// Attack the player if within attack range and the NPC is hostile.
func (npc *NPC) Attack(playerPos Position) (hit bool) {
	if !npc.Data.Hostile || npc.Data.Dead {
		return false
	}
	dist := beam_math.ManhattanDistance(npc.Pos.X, npc.Pos.Y, playerPos.X, playerPos.Y)
	if dist <= int(math.Round(npc.Data.AttackRange)) {
		// Face the player before attacking
		if playerPos.X > npc.Pos.X {
			npc.Data.Direction = DirRight
		} else if playerPos.X < npc.Pos.X {
			npc.Data.Direction = DirLeft
		} else if playerPos.Y > npc.Pos.Y {
			npc.Data.Direction = DirDown
		} else if playerPos.Y < npc.Pos.Y {
			npc.Data.Direction = DirUp
		}

		currentTime := float32(rl.GetTime())
		if npc.Data.AttackSpeed > 0 && (currentTime-npc.Data.LastAttackTime) > (1.0/float32(npc.Data.AttackSpeed)) {
			npc.Data.LastAttackTime = currentTime
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
func (npc *NPC) knockback(playerPos Position, tiles [][]Tile, dist int) {
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

	// Face the player after knockback
	if playerPos.X > npc.Pos.X {
		npc.Data.Direction = DirRight
	} else if playerPos.X < npc.Pos.X {
		npc.Data.Direction = DirLeft
	} else if playerPos.Y > npc.Pos.Y {
		npc.Data.Direction = DirDown
	} else if playerPos.Y < npc.Pos.Y {
		npc.Data.Direction = DirUp
	}
}
