package beam

// ItemType represents the category of an item
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
	Stackable   bool
	MaxStack    int
	Position    Position // Where the item is in the world
	Properties  map[string]interface{}
}

// Items is a collection of items with helper methods
type Items []Item

// Equippable is an interface for items that can be equipped
type Equippable interface {
	GetStats() map[string]int
	GetSlot() string
	GetRequirements() map[string]int
}

// ConsumableEffect represents what happens when an item is used
type ConsumableEffect func(target interface{}) error

// Consumable is an interface for items that can be used/consumed
type Consumable interface {
	Use(target interface{}) error
	GetEffect() ConsumableEffect
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
		if items[i].Position == pos {
			return &items[i]
		}
	}
	return nil
}

// Helper methods for individual items
func (i *Item) SetProperty(key string, value interface{}) {
	if i.Properties == nil {
		i.Properties = make(map[string]interface{})
	}
	i.Properties[key] = value
}

func (i *Item) GetProperty(key string) interface{} {
	if i.Properties == nil {
		return nil
	}
	return i.Properties[key]
}
