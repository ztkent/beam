package mapmaker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
	"github.com/ztkent/beam/resources"
)

// We can use this to export the map data to be loaded by our game.
func (t *TileGrid) SaveMapToFile(filename string) error {
	jsonData, err := json.MarshalIndent(t.Map, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal map data: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write map file: %w", err)
	}
	return nil
}

// SaveData represents the structure of our mapmaker save files
type SaveData struct {
	TileGrid        *TileGrid               `json:"tileGrid"`
	TileSize        int                     `json:"tileSize"`
	CurrentResIndex int                     `json:"currentResIndex"`
	ResourceState   resources.ResourceState `json:"resourceState"`
	RecentTextures  []string                `json:"recentTextures"`
}

type ConfigData struct {
	LastOpenedFile string `json:"lastOpenedFile"`
}

func SaveConfig(filename string) error {
	config := ConfigData{
		LastOpenedFile: filename,
	}
	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(".mapmaker-config", jsonData, 0644)
}

func LoadConfig() (string, error) {
	data, err := os.ReadFile(".mapmaker-config")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var config ConfigData
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.LastOpenedFile, nil
}

func (m *MapMaker) SaveMap(filename string) error {
	saveData := SaveData{
		TileSize:       m.uiState.tileSize,
		ResourceState:  m.resources.SaveState(),
		TileGrid:       m.tileGrid,
		RecentTextures: m.uiState.recentTextures,
	}

	jsonData, err := json.MarshalIndent(saveData, "", "    ")
	if err != nil {
		return err
	}

	if err := SaveConfig(m.currentFile); err != nil {
		return err
	}
	m.currentFile = filename
	if m.currentFile != "" {
		rl.SetWindowTitle(fmt.Sprintf("%s - (%s)", m.window.title, m.currentFile))
	} else {
		rl.SetWindowTitle(m.window.title)
	}
	return os.WriteFile(filename, jsonData, 0644)
}

// OldMap represents the previous map format for backward compatibility
type OldMap struct {
	Width, Height    int
	Tiles            [][]beam.TileType
	Textures         [][][]string
	TextureRotations [][][]float64
	Start            beam.Position
	Exit             beam.Position
	Respawn          beam.Position
	DungeonEntry     beam.Positions
}

func (m *MapMaker) LoadMap(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var saveData SaveData

	// Try to unmarshal with new format first
	err = json.Unmarshal(data, &saveData)
	if err != nil {
		// If it fails, try loading as old format
		var oldSaveData struct {
			TileGrid struct {
				OldMap
				offset               beam.Position
				hasSelection         bool
				selectedTiles        beam.Positions
				missingResourceTiles MissingResources
				viewportOffset       beam.Position
				viewportWidth        int
				viewportHeight       int
			} `json:"tileGrid"`
			TileSize       int                     `json:"tileSize"`
			ResourceState  resources.ResourceState `json:"resourceState"`
			RecentTextures []string                `json:"recentTextures"`
		}

		if err := json.Unmarshal(data, &oldSaveData); err != nil {
			return err
		}

		// Convert old format to new format
		saveData.TileSize = oldSaveData.TileSize
		saveData.ResourceState = oldSaveData.ResourceState
		saveData.RecentTextures = oldSaveData.RecentTextures

		// Convert tile grid
		saveData.TileGrid = &TileGrid{
			offset:               oldSaveData.TileGrid.offset,
			hasSelection:         oldSaveData.TileGrid.hasSelection,
			selectedTiles:        oldSaveData.TileGrid.selectedTiles,
			missingResourceTiles: oldSaveData.TileGrid.missingResourceTiles,
			viewportOffset:       oldSaveData.TileGrid.viewportOffset,
			viewportWidth:        oldSaveData.TileGrid.viewportWidth,
			viewportHeight:       oldSaveData.TileGrid.viewportHeight,
			Map: beam.Map{
				Width:        oldSaveData.TileGrid.Width,
				Height:       oldSaveData.TileGrid.Height,
				Start:        oldSaveData.TileGrid.Start,
				Exit:         oldSaveData.TileGrid.Exit,
				Respawn:      oldSaveData.TileGrid.Respawn,
				DungeonEntry: oldSaveData.TileGrid.DungeonEntry,
			},
		}

		// Convert tiles
		saveData.TileGrid.Tiles = make([][]beam.Tile, oldSaveData.TileGrid.Height)
		for y := range saveData.TileGrid.Tiles {
			saveData.TileGrid.Tiles[y] = make([]beam.Tile, oldSaveData.TileGrid.Width)
			for x := range saveData.TileGrid.Tiles[y] {
				var textures []beam.TileTexture
				for i, texName := range oldSaveData.TileGrid.Textures[y][x] {
					textures = append(textures, beam.TileTexture{
						Name:     texName,
						Rotation: oldSaveData.TileGrid.TextureRotations[y][x][i],
					})
				}
				saveData.TileGrid.Tiles[y][x] = beam.Tile{
					Type:     oldSaveData.TileGrid.Tiles[y][x],
					Pos:      beam.Position{X: x, Y: y},
					Textures: textures,
				}
			}
		}
	}

	// Close existing resources before loading new state
	if m.resources != nil {
		m.resources.Close()
	}
	m.resources = resources.InitFromState(saveData.ResourceState)

	// Update UI state with loaded map dimensions
	m.uiState.tileSize = saveData.TileSize
	m.uiState.recentTextures = saveData.RecentTextures
	m.uiState.gridWidth = saveData.TileGrid.Width
	m.uiState.gridHeight = saveData.TileGrid.Height

	// Set most recent texture as active
	if len(m.uiState.recentTextures) > 0 {
		if tex, err := m.resources.GetTexture("default", m.uiState.recentTextures[0]); err == nil {
			m.uiState.activeTexture = &tex
		}
	}

	m.updateGridSize()
	m.currentFile = filename

	// Update grid data directly
	m.tileGrid.Width = saveData.TileGrid.Width
	m.tileGrid.Height = saveData.TileGrid.Height
	m.initTileGrid()
	m.tileGrid = saveData.TileGrid

	if m.currentFile != "" {
		rl.SetWindowTitle(fmt.Sprintf("%s - (%s)", m.window.title, m.currentFile))
	} else {
		rl.SetWindowTitle(m.window.title)
	}

	// Validate the tile grid to ensure all textures are loaded, if not add to missing list
	m.ValidateTileGrid()
	return nil
}

func (m *MapMaker) ValidateTileGrid() error {
	// Clear the missing resource tiles list
	newGrid := make(MissingResources, 0)
	// Scan the grid, and make sure that any referenced textures are loaded.
	// If we cant find them, add them to the missing textures list.
	for y, textureY := range m.tileGrid.Tiles {
		for x, tile := range textureY {
			for _, texture := range tile.Textures {
				if _, err := m.resources.GetTexture("default", texture.Name); err != nil {
					newGrid = append(newGrid, MissingResource{tile: beam.Position{X: x, Y: y}, textureName: texture.Name})
				}
			}
		}
	}
	m.tileGrid.missingResourceTiles = newGrid
	return nil
}

func openLoadDialog() string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e", `POSIX path of (choose file with prompt "Choose a map file")`)
	case "linux":
		cmd = exec.Command("zenity", "--file-selection", "--file-filter=JSON (*.json)")
	default:
		return ""
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func openSaveDialog() string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e", `POSIX path of (choose file name with prompt "Save map as:" default name "untitled.json")`)
	case "linux":
		cmd = exec.Command("zenity", "--file-selection", "--save", "--file-filter=JSON (*.json)", "--confirm-overwrite")
	default:
		return ""
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
