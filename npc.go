package beam

import "time"

type NPC struct {
	Pos            Position
	LastMoveTime   float32
	LastActionTime time.Time
	MoveDelay      float32
	Data           NPCData
}

type NPCData struct {
	Name string

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
