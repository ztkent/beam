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
	ItemTypeEquipment ItemType = iota
	ItemTypeConsumable
	ItemTypeQuestItem
	ItemTypeResource
	ItemTypeMisc
)

// Item represents any item in the game world
type Item struct {
	ID          string
	Name        string
	Description string
	Type        ItemType
	Pos         Position
	Texture     *AnimatedTexture

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

type ItemStats struct {
	Attack      int
	Defense     int
	AttackSpeed int
	AttackRange int
}

type ItemRequirements struct {
	Level int
}

// Items is a collection of items with helper methods
type Items []Item

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
			return &items[i]
		}
	}
	return nil
}

func (items Items) FindByPosition(pos Position) *Item {
	for i := range items {
		if items[i].Pos == pos {
			return &items[i]
		}
	}
	return nil
}
