package beam

import beam_math "github.com/ztkent/beam/math"

/*
The items system supports:
  - Equipment items with stats and level requirements
  - Consumable items with effects
  - Quest items

Example usage:
    sword := NewItem("iron_sword", "Iron Sword", ItemTypeEquipment).
        AsEquipment().
        WithStats(ItemStats{
            Attack: 5,
            AttackSpeed: 10,
        }).
        WithRequirements(ItemRequirements{
            Level: 5,
        }).
        WithDescription("A basic iron sword")
*/

type ItemType int

const (
	ItemTypeNone ItemType = iota
	ItemTypeEquipment
	ItemTypeConsumable
	ItemTypeQuestItem
	ItemTypeResource
	ItemTypeMisc
)

type EquipmentType int

const (
	EquipmentTypeNone EquipmentType = iota
	EquipmentTypeWeapon
	EquipmentTypeArmor
	EquipmentTypeAccessory
	EquipmentTypeShield
)

func (t ItemType) String() string {
	switch t {
	case ItemTypeEquipment:
		return "Equipment"
	case ItemTypeConsumable:
		return "Consumable"
	case ItemTypeQuestItem:
		return "Quest Item"
	case ItemTypeResource:
		return "Resource"
	case ItemTypeMisc:
		return "Misc"
	default:
		return "None"
	}
}

func (t EquipmentType) String() string {
	switch t {
	case EquipmentTypeWeapon:
		return "Weapon"
	case EquipmentTypeArmor:
		return "Armor"
	case EquipmentTypeAccessory:
		return "Accessory"
	case EquipmentTypeShield:
		return "Shield"
	default:
		return "None"
	}
}

func AllItemTypes() []ItemType {
	return []ItemType{
		ItemTypeNone,
		ItemTypeEquipment,
		ItemTypeConsumable,
		ItemTypeQuestItem,
		ItemTypeResource,
		ItemTypeMisc,
	}
}

func AllEquipmentTypes() []EquipmentType {
	return []EquipmentType{
		EquipmentTypeNone,
		EquipmentTypeWeapon,
		EquipmentTypeArmor,
		EquipmentTypeAccessory,
		EquipmentTypeShield,
	}
}

type ItemStats struct {
	Attack      int
	Defense     int
	AttackSpeed int
	AttackRange int
	Effects     []ItemEffect
}

type EffectType int

const (
	EffectNone EffectType = iota
	EffectHealth
	EffectSpeed
	EffectAttack
)

type ItemEffect struct {
	ID            int64
	Type          EffectType
	Value         float64
	Duration      float64
	TimeRemaining float64
}

type ItemRequirements struct {
	Level int
}

// Item represents any item in the game world
type Item struct {
	ID          string
	Name        string
	Description string
	Pos         Position
	Texture     *AnimatedTexture

	Type          ItemType
	EquipmentType EquipmentType

	Blocking   bool
	Equippable bool
	Consumable bool
	Quantity   int
	Stackable  bool
	MaxStack   int

	Stats        ItemStats
	Requirements ItemRequirements

	Removed bool
}

// Items is a collection of items with helper methods
type Items []*Item

func (items Items) IsBlocked(x, y int) bool {
	for _, item := range items {
		if !item.Removed && item.Blocking && item.Pos.X == x && item.Pos.Y == y {
			return true
		}
	}
	return false
}

func (items Items) EquippableNearby(playerPos Position) Items {
	var equippableItems Items
	for _, item := range items {
		if !item.Removed && item.Equippable {
			dist := beam_math.ManhattanDistance(item.Pos.X, item.Pos.Y, playerPos.X, playerPos.Y)
			if dist <= 1 {
				equippableItems = append(equippableItems, item)
			}
		}
	}
	return equippableItems
}

func (items Items) Reset() {
	for i := range items {
		items[i].Removed = false
	}
}

func NewItem(id string, name string, itemType ItemType) *Item {
	return &Item{
		ID:           id,
		Name:         name,
		Type:         itemType,
		MaxStack:     1,
		Stats:        ItemStats{},
		Requirements: ItemRequirements{},
	}
}

// WithStats sets the item's stats
func (i *Item) WithStats(stats ItemStats) *Item {
	i.Stats = stats
	return i
}

// WithRequirements sets the item's requirements
func (i *Item) WithRequirements(reqs ItemRequirements) *Item {
	i.Requirements = reqs
	return i
}

// WithTexture sets the item's texture
func (i *Item) WithTexture(texture *AnimatedTexture) *Item {
	i.Texture = texture
	return i
}

// WithDescription sets the item's description
func (i *Item) WithDescription(desc string) *Item {
	i.Description = desc
	return i
}

// Configures the item as equipment
func (i *Item) AsEquipment(slot string, stats ItemStats) *Item {
	i.Type = ItemTypeEquipment
	i.Stats = stats
	i.Stackable = false
	i.MaxStack = 1
	return i
}

// AsConsumable configures the item as consumable
func (i *Item) AsConsumable(stackable bool) *Item {
	i.Type = ItemTypeConsumable
	i.Stackable = stackable
	if stackable {
		i.MaxStack = 99
	}
	i.Consumable = true
	return i
}

// Consumable is an interface for items that can be used/consumed
type Consumable interface {
	Use(target interface{}) error
}

// Helper methods for Items collection
func (items Items) FindByID(id string) *Item {
	for i := range items {
		if items[i].ID == id {
			return items[i]
		}
	}
	return nil
}

func (items Items) FindByPosition(pos Position) *Item {
	for i := range items {
		if items[i].Pos == pos {
			return items[i]
		}
	}
	return nil
}
