package save

import (
	"encoding/json"
	"os"
)

// Saveable is an interface that any game state must implement to be saved
type Saveable interface {
	// ToSaveData converts the game state to a serializable format
	ToSaveData() interface{}
	// FromSaveData restores the game state from saved data
	FromSaveData(data interface{}) error
}

// Manager handles saving and loading game state
type Manager struct {
	filename string
}

// NewManager creates a new save manager with the specified save file
func NewManager(filename string) *Manager {
	return &Manager{
		filename: filename,
	}
}

// SaveExists checks if a save file exists
func (m *Manager) SaveExists() bool {
	_, err := os.Stat(m.filename)
	return !os.IsNotExist(err)
}

// Save writes the game state to disk
func (m *Manager) Save(state Saveable) error {
	file, err := os.Create(m.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(state.ToSaveData())
}

// Load reads the game state from disk
func (m *Manager) Load(state Saveable) error {
	file, err := os.Open(m.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var data interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	return state.FromSaveData(data)
}

// DeleteSave removes the save file
func (m *Manager) DeleteSave() error {
	if m.SaveExists() {
		return os.Remove(m.filename)
	}
	return nil
}

func (m *Manager) LoadRaw() (interface{}, error) {
	file, err := os.Open(m.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
