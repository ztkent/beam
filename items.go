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
	ID           string
	Name         string
	Description  string
	Type         ItemType
	Stackable    bool
	MaxStack     int
	Position     Position // Where the item is in the world
	Properties   map[string]interface{}
	Stats        map[string]int   // Equipment stats
	Slot         string           // Equipment slot
	Requirements map[string]int   // Equipment requirements
	Effect       ConsumableEffect // Consumable effect
}

// Items is a collection of items with helper methods
type Items []Item

func NewItem(id string, name string, itemType ItemType) *Item {
	return &Item{
		ID:           id,
		Name:         name,
		Type:         itemType,
		MaxStack:     1,
		Properties:   make(map[string]interface{}),
		Stats:        make(map[string]int),
		Requirements: make(map[string]int),
	}
}

// Configures the item as equipment
func (i *Item) AsEquipment(slot string, stats map[string]int) *Item {
	i.Type = ItemTypeEquipment
	i.Slot = slot
	i.Stats = stats
	i.Stackable = false
	i.MaxStack = 1
	return i
}

// AsConsumable configures the item as consumable
func (i *Item) AsConsumable(effect ConsumableEffect, stackable bool) *Item {
	i.Type = ItemTypeConsumable
	i.Effect = effect
	i.Stackable = stackable
	if stackable {
		i.MaxStack = 99
	}
	return i
}

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

func (i *Item) WithDescription(desc string) *Item {
	i.Description = desc
	return i
}

func (i *Item) WithProperty(key string, value interface{}) *Item {
	i.SetProperty(key, value)
	return i
}

func (i *Item) WithRequirements(reqs map[string]int) *Item {
	i.Requirements = reqs
	return i
}
