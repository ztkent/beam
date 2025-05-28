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
            Texture:	{
				Up: &beam.AnimatedTexture{
					Frames: []beam.Texture{
						{Name: "orc_chief_attacks_5_0", ScaleX: 1, ScaleY: 1, OffsetX: 0, OffsetY: 0},
						...
					},
					IsAnimated:    true,
					AnimationTime: 0.2,
					Layer:         0,
				},
				...
			},
            IdleTexture: NewSimpleNPCTexture("guard"),
            AttackTexture: NewSimpleNPCTexture("guard"),
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
	Up    *AnimatedTexture
	Down  *AnimatedTexture
	Left  *AnimatedTexture
	Right *AnimatedTexture
}

// AttackState represents the different stages of an NPC's attack.
type AttackState int

const (
	// AttackIdle means the NPC is not currently attacking.
	AttackIdle AttackState = iota
	// AttackStart is the wind-up phase of an attack.
	AttackStart
	// AttackMid is the active part of the attack.
	AttackMid
	// AttackEnd is the recovery phase after an attack.
	AttackEnd
)

const (
	// AttackAnimProportionOfCooldown defines how much of the attack cooldown is used for the animation.
	AttackAnimProportionOfCooldown = 0.8 // 80%
	AttackStartProportion          = 0.25
	AttackMidProportion            = 0.35
	AttackEndProportion            = 0.40
	MinAttackPhaseDuration         = 0.05 // seconds
)

type NPCSize int

func (s NPCSize) GetDimensions() (width, height int) {
	size := int(s)
	if size == 0 {
		size = 1
	}
	return size, size
}

const (
	NPCSize1x1 NPCSize = iota + 1 // 1x1 (default)
	NPCSize2x2                    // 2x2
	NPCSize3x3                    // 3x3
	NPCSize4x4                    // 4x4
)

type NPCs []*NPC

func (npcs NPCs) IsBlocked(x, y int) bool {
	for _, npc := range npcs {
		if !npc.Data.Dead && npc.Data.Impassable {
			if npc.occupiesTile(x, y) {
				return true
			}
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
			dist := e.distanceToNPC(playerPos.X, playerPos.Y)
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
	Name string

	Texture       *NPCTexture
	IdleTexture   *NPCTexture
	AttackTexture *NPCTexture

	SpawnPos Position
	Size     NPCSize

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

	Direction Direction
	IsIdle    bool

	Attackable          bool
	Impassable          bool
	Hostile             bool
	WanderRange         int
	AggroRange          int
	AttackState         AttackState
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

// GetCurrentTexture returns the appropriate AnimatedTexture for the NPC
// based on its direction, idle, and attacking state.
func (npc *NPC) GetCurrentTexture() *AnimatedTexture {
	var base, idle, attack *AnimatedTexture
	switch npc.Data.Direction {
	case DirUp:
		base = npc.Data.Texture.Up
		if npc.Data.IdleTexture != nil {
			idle = npc.Data.IdleTexture.Up
		}
		if npc.Data.AttackTexture != nil {
			attack = npc.Data.AttackTexture.Up
		}
	case DirDown:
		base = npc.Data.Texture.Down
		if npc.Data.IdleTexture != nil {
			idle = npc.Data.IdleTexture.Down
		}
		if npc.Data.AttackTexture != nil {
			attack = npc.Data.AttackTexture.Down
		}
	case DirLeft:
		base = npc.Data.Texture.Left
		if npc.Data.IdleTexture != nil {
			idle = npc.Data.IdleTexture.Left
		}
		if npc.Data.AttackTexture != nil {
			attack = npc.Data.AttackTexture.Left
		}
	case DirRight:
		base = npc.Data.Texture.Right
		if npc.Data.IdleTexture != nil {
			idle = npc.Data.IdleTexture.Right
		}
		if npc.Data.AttackTexture != nil {
			attack = npc.Data.AttackTexture.Right
		}
	default:
		return nil
	}

	// Priority: Attack > Idle > Base
	currentTime := float32(rl.GetTime())

	// Don't swap to idle immediately after finishing a long attack
	isAttacking := (npc.Data.AttackState != AttackIdle) || (currentTime-npc.Data.LastAttackTime < 2.0)
	if isAttacking && attack != nil {
		return attack
	}
	if npc.Data.IsIdle && idle != nil {
		return idle
	}
	return base
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

	npc.updateAttackState()
	if npc.Data.AttackState == AttackIdle {
		npc.Wander(playerPos, currMap)
	}
	return false
}

func (npc *NPC) updateAttackState() {
	if npc.Data.AttackState != AttackIdle {
		npc.Data.AttackStateTime += rl.GetFrameTime()

		var currentPhaseExpectedDuration float32
		calculateAttackPhaseDuration := func(attackSpeed float64, phaseProportion float32) float32 {
			if attackSpeed <= 0 {
				return MinAttackPhaseDuration * 3
			}
			totalCooldown := float32(1.0 / attackSpeed)
			totalAnimationTime := totalCooldown * AttackAnimProportionOfCooldown
			calculatedDuration := totalAnimationTime * phaseProportion
			return float32(math.Max(float64(MinAttackPhaseDuration), float64(calculatedDuration)))
		}

		switch npc.Data.AttackState {
		case AttackStart:
			currentPhaseExpectedDuration = calculateAttackPhaseDuration(npc.Data.AttackSpeed, AttackStartProportion)
			if npc.Data.AttackStateTime >= currentPhaseExpectedDuration {
				npc.Data.AttackState = AttackMid
				npc.Data.AttackStateTime = 0
			}
		case AttackMid:
			currentPhaseExpectedDuration = calculateAttackPhaseDuration(npc.Data.AttackSpeed, AttackMidProportion)
			if npc.Data.AttackStateTime >= currentPhaseExpectedDuration {
				npc.Data.AttackState = AttackEnd
				npc.Data.AttackStateTime = 0
			}
		case AttackEnd:
			currentPhaseExpectedDuration = calculateAttackPhaseDuration(npc.Data.AttackSpeed, AttackEndProportion)
			if npc.Data.AttackStateTime >= currentPhaseExpectedDuration {
				npc.Data.AttackState = AttackIdle
				npc.Data.AttackStateTime = 0
			}
		}
	}
}

// Interact with a player
func (npc *NPC) Interact(playerPos Position, currChat *chat.Chat) {
	if !npc.Data.Interactable {
		return
	}

	dist := npc.distanceToNPC(playerPos.X, playerPos.Y)
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
	startPos := Position{X: npc.Pos.X, Y: npc.Pos.Y}
	distToPlayer := npc.distanceToNPC(playerPos.X, playerPos.Y)
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
			if npc.canMoveTo(newX, newY, currMap) {
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
				if !npc.canMoveTo(npc.Pos.X+dx, npc.Pos.Y, currMap) {
					dx = 0
					dy = beam_math.Sign(yDiff)
				}
			} else {
				dy = beam_math.Sign(yDiff)
				dx = 0
				if !npc.canMoveTo(npc.Pos.X, npc.Pos.Y+dy, currMap) {
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

		newDist := npc.distanceToNPC(playerPos.X-dx, playerPos.Y-dy)
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

			newDist := npc.distanceToNPC(playerPos.X-dx, playerPos.Y-dy)
			if newDist < 1 {
				dx, dy = 0, 0
			}
		}
	}

	newX := npc.Pos.X + dx
	newY := npc.Pos.Y + dy
	if npc.canMoveTo(newX, newY, currMap) {
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

	idleThreshold := float32(3.0)
	if npc.Pos.X != startPos.X || npc.Pos.Y != startPos.Y {
		npc.Data.LastMoveTime = currentTime
		npc.Data.IsIdle = false
	} else if currentTime-npc.Data.LastMoveTime > idleThreshold {
		npc.Data.IsIdle = true
	}
}

// Attack the player if within attack range and the NPC is hostile.
func (npc *NPC) Attack(playerPos Position) (hit bool) {
	if !npc.Data.Hostile || npc.Data.Dead || npc.Data.AttackState != AttackIdle {
		return false
	}

	dist := npc.distanceToNPC(playerPos.X, playerPos.Y)
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
		attackCooldown := float32(0.0)
		if npc.Data.AttackSpeed > 0 {
			attackCooldown = 1.0 / float32(npc.Data.AttackSpeed)
		} else {
			attackCooldown = 60
		}

		if (currentTime - npc.Data.LastAttackTime) >= attackCooldown {
			npc.Data.LastAttackTime = currentTime
			npc.Data.LastMoveTime = currentTime
			npc.Data.AttackState = AttackStart
			npc.Data.AttackStateTime = 0
			npc.Data.IsIdle = false
			return true
		}
	}
	return false
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

func (npc *NPC) occupiesTile(x, y int) bool {
	width, height := npc.Data.Size.GetDimensions()

	// Calculate the NPC's bounding box (top-left to bottom-right)
	left := npc.Pos.X
	right := npc.Pos.X + width - 1
	top := npc.Pos.Y
	bottom := npc.Pos.Y + height - 1

	return x >= left && x <= right && y >= top && y <= bottom
}

func (npc *NPC) distanceToNPC(x, y int) int {
	width, height := npc.Data.Size.GetDimensions()

	// Calculate the NPC's bounding box
	left := npc.Pos.X
	right := npc.Pos.X + width - 1
	top := npc.Pos.Y
	bottom := npc.Pos.Y + height - 1

	// Find closest point on the NPC to the given position
	closestX := x
	closestY := y

	if x < left {
		closestX = left
	} else if x > right {
		closestX = right
	}

	if y < top {
		closestY = top
	} else if y > bottom {
		closestY = bottom
	}

	return beam_math.ManhattanDistance(x, y, closestX, closestY)
}

// canMoveTo checks if the NPC can move to the given position
func (npc *NPC) canMoveTo(newX, newY int, currMap *Map) bool {
	width, height := npc.Data.Size.GetDimensions()

	// Check all tiles the NPC would occupy at the new position
	for dx := 0; dx < width; dx++ {
		for dy := 0; dy < height; dy++ {
			checkX := newX + dx
			checkY := newY + dy

			// Check bounds
			if checkX <= 0 || checkX >= len(currMap.Tiles[0])-1 ||
				checkY <= 0 || checkY >= len(currMap.Tiles)-1 {
				return false
			}

			// Check tile type
			if currMap.Tiles[checkY][checkX].Type == WallTile ||
				currMap.Tiles[checkY][checkX].Type == ChestTile {
				return false
			}

			// Check for other NPCs (excluding self)
			for _, otherNPC := range currMap.NPCs {
				if otherNPC != npc && !otherNPC.Data.Dead &&
					otherNPC.Data.Impassable && otherNPC.occupiesTile(checkX, checkY) {
					return false
				}
			}

			// Check for items
			if currMap.Items.IsBlocked(checkX, checkY) {
				return false
			}
		}
	}
	return true
}
