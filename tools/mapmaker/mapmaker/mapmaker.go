package mapmaker

import (
	"fmt"
	"strings"

	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
	"github.com/ztkent/beam/resources"
	"github.com/ztkent/beam/tools/spritesheet-viewer/viewer"
)

type MapMaker struct {
	window             *Window
	resources          *resources.ResourceManager
	uiState            *UIState
	tileGrid           *TileGrid
	currentFile        string
	showResourceViewer bool
	showTileInfo       bool
	showRecentTextures bool
}

type Window struct {
	width  int32
	height int32
	title  string
}

type UIState struct {
	tileSize        int
	menuBarHeight   int
	statusBarHeight int
	uiTextures      map[string]rl.Texture2D
	activeTexture   *resources.TextureInfo
	selectedTool    string
	// Active toast notification
	toast *Toast

	// Resource Viewer
	resourceViewerScroll int

	// Tile Info Popup
	tileInfoPos     beam.Position
	tileInfoPopupX  int32
	tileInfoPopupY  int32
	isDraggingPopup bool

	// Recent Textures
	recentTextures []string

	// Resource Manage Mode
	resourceManageMode bool

	// Track long right click for tool swap
	rightClickStartTime float64

	// Eraser Tool Swap
	hasSwappedEraser bool

	// Layers Tool Swap
	hasSwappedLayers bool

	// Location Tool Mode
	locationMode int

	// Grid Width/Height Controls
	gridWidth  int
	gridHeight int
}

type TileGrid struct {
	offset               beam.Position // The offset of the grid in the window
	hasSelection         bool
	selectedTiles        beam.Positions   // These are the tiles that are selected by the user
	missingResourceTiles MissingResources // This is every tile that has a texture, that is missing in the resource manager

	// This is the actual map we will use in game with beam.
	beam.Map

	viewportOffset beam.Position // Tracks how many tiles to offset the view
	viewportWidth  int           // Width of visible viewport in tiles
	viewportHeight int           // Height of visible viewport in tiles
}

type MissingResources []MissingResource
type MissingResource struct {
	tile        beam.Position
	textureName string
}

func (p MissingResources) Contains(pos beam.Position, texture string) bool {
	for _, item := range p {
		if item.tile == pos && item.textureName == texture {
			return true
		}
	}
	return false
}

const (
	// Gutter sizes for the window, since we define the grid size directly
	WidthGutter       = 150
	HeightGutter      = 150
	DefaultTileSize   = 20
	DefaultGridWidth  = 64
	DefaultGridHeight = 40
	MaxDisplayWidth   = 64
	MaxDisplayHeight  = 40
)

type ResourceDialog struct {
	name        string
	path        string
	isSheet     bool
	sheetMargin int32
	gridSize    int32
	visible     bool
}

func NewMapMaker(width, height int32) *MapMaker {
	mm := &MapMaker{
		window: &Window{
			width:  1280 + WidthGutter,
			height: 800 + HeightGutter,
			title:  "2D Map Editor",
		},
		uiState: &UIState{
			tileSize:   DefaultTileSize,   // Default size
			gridWidth:  DefaultGridWidth,  // Default size
			gridHeight: DefaultGridHeight, // Default size

			menuBarHeight:   60,
			statusBarHeight: 25,
			uiTextures:      make(map[string]rl.Texture2D),
			activeTexture:   nil,
			selectedTool:    "",
			toast:           nil,
			recentTextures:  make([]string, 0),

			resourceManageMode: false,
			hasSwappedEraser:   false,
			hasSwappedLayers:   false,
			locationMode:       0,
		},
		tileGrid: &TileGrid{
			offset:         beam.Position{X: 0, Y: 0},
			selectedTiles:  beam.Positions{{X: -1, Y: -1}},
			hasSelection:   false,
			viewportOffset: beam.Position{X: 0, Y: 0},
			viewportWidth:  MaxDisplayWidth,
			viewportHeight: MaxDisplayHeight,
		},
		currentFile: "",
	}
	mm.updateGridSize()
	return mm
}

func (m *MapMaker) Init() {
	rl.InitWindow(m.window.width, m.window.height, m.window.title)
	rl.SetTargetFPS(60)

	// Load UI textures
	m.uiState.uiTextures["add"] = rl.LoadTexture("../assets/add.png")
	m.uiState.uiTextures["view"] = rl.LoadTexture("../assets/view.png")
	m.uiState.uiTextures["save"] = rl.LoadTexture("../assets/save.png")
	m.uiState.uiTextures["load"] = rl.LoadTexture("../assets/load.png")
	m.uiState.uiTextures["close"] = rl.LoadTexture("../assets/reset.png")

	//  Control Textures
	m.uiState.uiTextures["paintbrush"] = rl.LoadTexture("../assets/paintbrush.png")
	m.uiState.uiTextures["paintbucket"] = rl.LoadTexture("../assets/paintbucket.png")
	m.uiState.uiTextures["eraser"] = rl.LoadTexture("../assets/eraser.png")
	m.uiState.uiTextures["pencileraser"] = rl.LoadTexture("../assets/pencileraser.png")
	m.uiState.uiTextures["select"] = rl.LoadTexture("../assets/select.png")
	m.uiState.uiTextures["layerwall"] = rl.LoadTexture("../assets/wall.png")
	m.uiState.uiTextures["layerground"] = rl.LoadTexture("../assets/soil.png")
	m.uiState.uiTextures["layers"] = m.uiState.uiTextures["layerground"]
	m.uiState.uiTextures["location"] = rl.LoadTexture("../assets/location.png")

	// Add directional arrows for viewport
	m.uiState.uiTextures["up"] = rl.LoadTexture("../assets/up.png")
	m.uiState.uiTextures["down"] = rl.LoadTexture("../assets/down.png")
	m.uiState.uiTextures["left"] = rl.LoadTexture("../assets/left.png")
	m.uiState.uiTextures["right"] = rl.LoadTexture("../assets/right.png")

	m.resources = resources.NewResourceManager()
	m.initTileGrid()
}

func (m *MapMaker) Run() {
	for {
		// Handle Exit/Escape behavior
		if rl.WindowShouldClose() {
			if rl.IsKeyPressed(rl.KeyEscape) {
				if m.tileGrid.hasSelection {
					m.tileGrid.hasSelection = false
					m.tileGrid.selectedTiles = beam.Positions{}
					continue
				}
			} else {
				break
			}
		}

		// Capture cmd/ctrl+s for save
		if rl.IsKeyPressed(rl.KeyS) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyLeftSuper)) {
			if m.currentFile != "" {
				if err := m.SaveMap(m.currentFile); err != nil {
					m.showToast("Error saving map: "+err.Error(), ToastError)
				} else {
					m.showToast("Map saved successfully!", ToastSuccess)
				}
			}
		}

		m.update() // Update settings, configs, and UI state.
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		m.renderGrid()  // Render the current map
		m.renderUI()    // Render the UI
		m.renderToast() // Render any active toasts
		rl.EndDrawing()
	}
}

func (m *MapMaker) update() {
	tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, closeMapBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn := m.getUIButtons()

	m.handleResizeGrid(tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn)
	m.handleSaveLoadClose(saveBtn, loadBtn, closeMapBtn)
	m.handleResourceViewer(viewResourcesBtn, loadResourceBtn)

	// Tool button handlers
	m.handleMapTools(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn)

	// Center the grid in the window
	maxVisibleWidth := MaxDisplayWidth * DefaultTileSize / m.uiState.tileSize
	maxVisibleHeight := MaxDisplayHeight * DefaultTileSize / m.uiState.tileSize
	displayWidth := min(m.tileGrid.Width, maxVisibleWidth)
	displayHeight := min(m.tileGrid.Height, maxVisibleHeight)
	totalGridWidth := displayWidth * m.uiState.tileSize
	totalGridHeight := displayHeight * m.uiState.tileSize

	// Calculate available workspace excluding UI elements
	workspaceWidth := int(m.window.width)
	workspaceHeight := int(m.window.height) - m.uiState.menuBarHeight - m.uiState.statusBarHeight

	// Center the grid in the available workspace
	m.tileGrid.offset = beam.Position{
		X: (workspaceWidth - totalGridWidth) / 2,
		Y: (workspaceHeight-totalGridHeight)/2 + m.uiState.menuBarHeight,
	}

	// Handle tile selection - Handle the viewport offset
	mousePos := rl.GetMousePosition()
	gridX := int((mousePos.X-float32(m.tileGrid.offset.X))/float32(m.uiState.tileSize)) + m.tileGrid.viewportOffset.X
	gridY := int((mousePos.Y-float32(m.tileGrid.offset.Y))/float32(m.uiState.tileSize)) + m.tileGrid.viewportOffset.Y

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Check if any dialogs are open
		// Check if click is within grid bounds and below menu bar
		if !m.showResourceViewer &&
			!m.showTileInfo &&
			!m.showRecentTextures &&
			gridX >= 0 && gridX < m.tileGrid.Width &&
			gridY >= 0 && gridY < m.tileGrid.Height &&
			mousePos.Y > float32(m.uiState.menuBarHeight) {
			if m.uiState.selectedTool == "paintbucket" {
				m.tileGrid.selectedTiles = m.floodFillSelection(gridX, gridY)
			} else {
				m.tileGrid.selectedTiles = beam.Positions{{X: gridX, Y: gridY}}
			}
			m.tileGrid.hasSelection = true
		}
	} else if rl.IsMouseButtonDown(rl.MouseLeftButton) && m.tileGrid.hasSelection {

		// Allow drag selection for some tools
		if m.uiState.selectedTool == "paintbrush" ||
			m.uiState.selectedTool == "eraser" ||
			m.uiState.selectedTool == "pencileraser" ||
			m.uiState.selectedTool == "layers" ||
			(m.uiState.selectedTool == "location" && m.uiState.locationMode == 1) {
			if gridX >= 0 && gridX < m.tileGrid.Width &&
				gridY >= 0 && gridY < m.tileGrid.Height &&
				mousePos.Y > float32(m.uiState.menuBarHeight) {
				newPos := beam.Position{X: gridX, Y: gridY}
				alreadySelected := slices.Contains(m.tileGrid.selectedTiles, newPos)
				if !alreadySelected {
					m.tileGrid.selectedTiles = append(m.tileGrid.selectedTiles, newPos)
				}
			}
		}
	}

	if m.tileGrid.hasSelection {
		mousePos := rl.GetMousePosition()

		// Check if we're clicking within the tile info popup
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && m.showTileInfo {
			dialogWidth := 300
			closeBtn := rl.Rectangle{
				X:      float32(m.uiState.tileInfoPopupX + int32(dialogWidth) - 30),
				Y:      float32(m.uiState.tileInfoPopupY + 5),
				Width:  25,
				Height: 25,
			}
			if rl.CheckCollisionPointRec(mousePos, closeBtn) {
				m.showTileInfo = false
				return
			}
			dragArea := rl.Rectangle{
				X:      float32(m.uiState.tileInfoPopupX),
				Y:      float32(m.uiState.tileInfoPopupY),
				Width:  float32(dialogWidth),
				Height: 30,
			}
			if rl.CheckCollisionPointRec(mousePos, dragArea) {
				return
			}
		}

		if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			switch m.uiState.selectedTool {
			case "paintbrush", "paintbucket":
				if m.uiState.activeTexture != nil {
					for _, pos := range m.tileGrid.selectedTiles {
						selectedX := int(pos.X)
						selectedY := int(pos.Y)
						m.tileGrid.Tiles[selectedY][selectedX].Type = beam.FloorTile
						m.tileGrid.Tiles[selectedY][selectedX].Textures = append(
							m.tileGrid.Tiles[selectedY][selectedX].Textures,
							beam.TileTexture{
								Name:     m.uiState.activeTexture.Name,
								Rotation: 0.0,
							},
						)
					}
				}
			case "eraser":
				for _, pos := range m.tileGrid.selectedTiles {
					selectedX := int(pos.X)
					selectedY := int(pos.Y)
					m.tileGrid.Tiles[selectedY][selectedX].Type = beam.FloorTile
					m.tileGrid.Tiles[selectedY][selectedX].Textures = nil
				}
			case "pencileraser":
				for _, pos := range m.tileGrid.selectedTiles {
					selectedX := int(pos.X)
					selectedY := int(pos.Y)
					tile := &m.tileGrid.Tiles[selectedY][selectedX]
					if len(tile.Textures) > 0 {
						tile.Textures = tile.Textures[:len(tile.Textures)-1]
					}
				}
			case "select":
				if !m.showTileInfo {
					pos := m.tileGrid.selectedTiles[0]
					mousePos := rl.GetMousePosition()
					m.uiState.tileInfoPopupX = int32(mousePos.X)
					m.uiState.tileInfoPopupY = int32(mousePos.Y)
					m.showTileInfo = true
					m.uiState.tileInfoPos = pos
				}
			case "layers":
				for _, pos := range m.tileGrid.selectedTiles {
					selectedX := int(pos.X)
					selectedY := int(pos.Y)
					tileType := beam.FloorTile // Ground
					if m.uiState.hasSwappedLayers {
						tileType = beam.WallTile // Wall
					}
					m.tileGrid.Tiles[selectedY][selectedX].Type = tileType
				}
				break
			case "location":
				// Reset the list if were about to add new positions
				if m.uiState.locationMode == 1 {
					m.tileGrid.DungeonEntry = beam.Positions{}
				}

				for _, tile := range m.tileGrid.selectedTiles {
					switch m.uiState.locationMode {
					case 0:
						m.tileGrid.Start = tile
					case 1:
						m.tileGrid.DungeonEntry = append(m.tileGrid.DungeonEntry, tile)
					case 2:
						m.tileGrid.Respawn = tile
					case 3:
						m.tileGrid.Exit = tile
					}
				}
				break
			}
		}
	}
}

// handleMapTools handles the selecting and swapping of tools
func (m *MapMaker) handleMapTools(paintbrushBtn IconButton, paintbucketBtn IconButton, eraseBtn IconButton, selectBtn IconButton, layersBtn IconButton, locationBtn IconButton) {
	if m.isIconButtonClicked(paintbrushBtn) {
		if m.uiState.selectedTool == "paintbrush" {
			m.uiState.selectedTool = ""
		} else {
			m.uiState.selectedTool = "paintbrush"
			m.showToast("Paintbrush tool selected", ToastInfo)
		}
	}
	if m.isIconButtonClicked(paintbucketBtn) {
		if m.uiState.selectedTool == "paintbucket" {
			m.uiState.selectedTool = ""
		} else {
			m.uiState.selectedTool = "paintbucket"
			m.showToast("Paint bucket tool selected", ToastInfo)
		}
	}
	if m.isIconButtonClicked(eraseBtn) {
		if m.uiState.selectedTool == "eraser" || m.uiState.selectedTool == "pencileraser" {
			m.uiState.selectedTool = ""
		} else {
			name := "eraser"
			// Use pencileraser if swapped
			if m.uiState.hasSwappedEraser {
				name = "pencileraser"
			}
			m.uiState.selectedTool = name
			m.showToast(name+" tool selected", ToastInfo)
		}
	}
	if m.isIconButtonClicked(selectBtn) {
		if m.uiState.selectedTool == "select" {
			m.uiState.selectedTool = ""
		} else {
			m.uiState.selectedTool = "select"
			m.tileGrid.hasSelection = false
			m.tileGrid.selectedTiles = beam.Positions{}
			m.showToast("Select tool selected", ToastInfo)
		}
	}
	if m.isIconButtonClicked(layersBtn) {
		if m.uiState.selectedTool == "layers" {
			m.uiState.selectedTool = ""
		} else {
			m.uiState.selectedTool = "layers"
			m.showToast("Layers tool selected", ToastInfo)
		}
	}
	if m.isIconButtonClicked(locationBtn) {
		if m.uiState.selectedTool == "location" {
			m.uiState.selectedTool = ""
		} else {
			m.uiState.selectedTool = "location"
			m.showToast("Location tool selected", ToastInfo)
		}
	}

	// Handle tool swaps
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		if m.uiState.rightClickStartTime == 0 {
			m.uiState.rightClickStartTime = rl.GetTime()
		} else if rl.GetTime()-m.uiState.rightClickStartTime > 0.5 {
			// Handle eraser swap
			if m.uiState.selectedTool == "eraser" || m.uiState.selectedTool == "pencileraser" {
				m.uiState.uiTextures["eraser"], m.uiState.uiTextures["pencileraser"] =
					m.uiState.uiTextures["pencileraser"], m.uiState.uiTextures["eraser"]
				if m.uiState.selectedTool == "eraser" {
					m.uiState.selectedTool = "pencileraser"
				} else {
					m.uiState.selectedTool = "eraser"
				}
				m.uiState.hasSwappedEraser = !m.uiState.hasSwappedEraser
			}

			// Handle layers swap
			if m.uiState.selectedTool == "layers" {
				m.uiState.hasSwappedLayers = !m.uiState.hasSwappedLayers
				if m.uiState.hasSwappedLayers {
					m.uiState.uiTextures["layers"] = m.uiState.uiTextures["layerwall"]
				} else {
					m.uiState.uiTextures["layers"] = m.uiState.uiTextures["layerground"]
				}
			}

			// Handle location swap
			if m.uiState.selectedTool == "location" {
				m.uiState.locationMode = (m.uiState.locationMode + 1) % 4 // Cycle through 4 states
				modeNames := []string{"Player Start", "Dungeon Entrance", "Respawn", "Exit"}
				m.showToast(fmt.Sprintf("Location Mode: %s", modeNames[m.uiState.locationMode]), ToastInfo)
			}

			m.uiState.rightClickStartTime = 0
		}
	} else {
		m.uiState.rightClickStartTime = 0
	}
}

// handleResourceViewer handles the resource viewer tool settings
func (m *MapMaker) handleResourceViewer(viewResourcesBtn IconButton, loadResourceBtn IconButton) {
	if m.isIconButtonClicked(viewResourcesBtn) {
		m.showResourceViewer = true
	}
	if m.isIconButtonClicked(loadResourceBtn) {
		name, filepath, isSheet, sheetMargin, gridSize, err := openLoadResourceDialog()
		if err != "" {
			fmt.Println("Error loading resource:", err)
			m.showToast(err, ToastError)
		} else if err := m.loadResource(name, filepath, isSheet, sheetMargin, gridSize); err != nil {
			fmt.Println("Error loading texture:", err)
			m.showToast(err.Error(), ToastError)
		}
	}
}

// handleSaveLoad handles the save, load, and close tools
func (m *MapMaker) handleSaveLoadClose(saveBtn, loadBtn, closeMapBtn IconButton) {
	if m.isIconButtonClicked(saveBtn) {
		if m.currentFile != "" {
			if err := m.SaveMap(m.currentFile); err != nil {
				m.showToast("Error saving map: "+err.Error(), ToastError)
			} else {
				m.showToast("Map saved successfully!", ToastSuccess)
			}
		} else {
			filename := openSaveDialog()
			if filename != "" {
				if !strings.HasSuffix(filename, ".json") {
					filename += ".json"
				}
				if err := m.SaveMap(filename); err != nil {
					m.showToast("Error saving map: "+err.Error(), ToastError)
				} else {
					m.showToast("Map saved successfully!", ToastSuccess)
				}
			}
		}
	}
	if m.isIconButtonClicked(loadBtn) {
		filename := openLoadDialog()
		if filename != "" {
			if err := m.LoadMap(filename); err != nil {
				m.showToast("Error loading map: "+err.Error(), ToastError)
			} else {
				m.showToast("Map loaded successfully!", ToastSuccess)
			}
		}
	}
	if m.isIconButtonClicked(closeMapBtn) {
		if openCloseConfirmationDialog() {
			// Reset to default state
			m.uiState.tileSize = DefaultTileSize
			m.uiState.gridWidth = DefaultGridWidth
			m.uiState.gridHeight = DefaultGridHeight

			m.showResourceViewer = false
			m.uiState.resourceViewerScroll = 0
			m.currentFile = ""
			rl.SetWindowTitle(m.window.title)

			// Reset grid
			m.updateGridSize()
			m.initTileGrid()
		}
	}
}

// handleResizeGrid handles the resizing of the grid based on specified size
func (m *MapMaker) handleResizeGrid(tileSmallerBtn Button, tileLargerBtn Button, widthSmallerBtn Button, widthLargerBtn Button, heightSmallerBtn Button, heightLargerBtn Button) {
	if m.isButtonClicked(tileSmallerBtn) {
		if m.uiState.tileSize > 8 {
			m.uiState.tileSize--
			m.updateGridSize()
			m.resizeGrid()
		}
	}
	if m.isButtonClicked(tileLargerBtn) {
		if m.uiState.tileSize < 64 {
			m.uiState.tileSize++
			m.updateGridSize()
			m.resizeGrid()
		}
	}

	if m.isButtonClicked(widthSmallerBtn) {
		if m.uiState.gridWidth > 10 {
			m.uiState.gridWidth--
			m.updateGridSize()
			m.resizeGrid()
		}
	}
	if m.isButtonClicked(widthLargerBtn) {
		if m.uiState.gridWidth < 100 {
			m.uiState.gridWidth++
			m.updateGridSize()
			m.resizeGrid()
		}
	}
	if m.isButtonClicked(heightSmallerBtn) {
		if m.uiState.gridHeight > 10 {
			m.uiState.gridHeight--
			m.updateGridSize()
			m.resizeGrid()
		}
	}
	if m.isButtonClicked(heightLargerBtn) {
		if m.uiState.gridHeight < 100 {
			m.uiState.gridHeight++
			m.updateGridSize()
			m.resizeGrid()
		}
	}
}

// resizeGrid resizes the grid its current dimensions
func (m *MapMaker) resizeGrid() {
	newTiles := make([][]beam.Tile, m.tileGrid.Height)
	for i := range newTiles {
		newTiles[i] = make([]beam.Tile, m.tileGrid.Width)
	}

	// Copy existing tiles
	for y := 0; y < min(len(m.tileGrid.Tiles), m.tileGrid.Height); y++ {
		for x := 0; x < min(len(m.tileGrid.Tiles[y]), m.tileGrid.Width); x++ {
			newTiles[y][x] = m.tileGrid.Tiles[y][x]
			newTiles[y][x].Pos = beam.Position{X: x, Y: y}
		}
	}

	m.tileGrid.Tiles = newTiles
}

// initTileGrid initializes the tile grid with default values
func (m *MapMaker) initTileGrid() {
	m.tileGrid.Tiles = make([][]beam.Tile, m.tileGrid.Height)
	for i := range m.tileGrid.Tiles {
		m.tileGrid.Tiles[i] = make([]beam.Tile, m.tileGrid.Width)
		for j := range m.tileGrid.Tiles[i] {
			m.tileGrid.Tiles[i][j] = beam.Tile{
				Type:     beam.FloorTile,
				Pos:      beam.Position{X: j, Y: i},
				Textures: make([]beam.TileTexture, 0),
			}
		}
	}
}

// updateGridSize updates the grid size based on the UI state
func (m *MapMaker) updateGridSize() {
	m.tileGrid.Width = m.uiState.gridWidth
	m.tileGrid.Height = m.uiState.gridHeight
}

// loadResource loads a resource into the resource manager
// Supports loading both spritesheets and individual textures
// For spritesheets, it will display the spritesheet viewer to confirm options
func (m *MapMaker) loadResource(name string, filepath string, isSheet bool, sheetMargin int32, gridSize int32) error {
	newRes := resources.Resource{
		Name:        name,
		Path:        filepath,
		IsSheet:     isSheet,
		SheetMargin: sheetMargin,
		GridSize:    gridSize,
	}

	// Display the spritesheet viewer, have user confirm sheet options
	if isSheet {
		finalGridSize, finalSheetMargin, err := viewer.ViewSpritesheet(newRes)
		if err != nil {
			return err
		}
		newRes.GridSize = finalGridSize
		newRes.SheetMargin = finalSheetMargin
	}

	err := m.resources.AddResource("default", newRes)
	if err != nil {
		return err
	}
	m.ValidateTileGrid()
	return nil
}

func (m *MapMaker) getUIButtons() (tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn Button, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, closeMapBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn IconButton) {
	widthSmallerBtn = m.NewButton(10, 8, 30, 20, "-")
	widthLargerBtn = m.NewButton(85, 8, 30, 20, "+")
	heightSmallerBtn = m.NewButton(10, 33, 30, 20, "-")
	heightLargerBtn = m.NewButton(85, 33, 30, 20, "+")
	tileSmallerBtn = m.NewButton(10, 58, 30, 20, "-")
	tileLargerBtn = m.NewButton(85, 58, 30, 20, "+")

	// Load, save, close buttons
	loadBtn = m.NewIconButton(
		float32(m.window.width-160),
		15,
		40,
		30,
		m.uiState.uiTextures["load"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["load"].Width), Height: float32(m.uiState.uiTextures["load"].Height)},
		"Load Map",
	)
	saveBtn = m.NewIconButton(
		float32(m.window.width-110),
		15,
		40,
		30,
		m.uiState.uiTextures["save"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["save"].Width), Height: float32(m.uiState.uiTextures["save"].Height)},
		"Save Map",
	)
	closeMapBtn = m.NewIconButton(
		float32(m.window.width-60),
		15,
		40,
		30,
		m.uiState.uiTextures["close"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["close"].Width), Height: float32(m.uiState.uiTextures["close"].Height)},
		"Close Map",
	)

	// Resource viewer / loader buttons
	loadResourceBtn = m.NewIconButton(
		float32(m.window.width-245),
		15,
		40,
		30,
		m.uiState.uiTextures["add"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["add"].Width), Height: float32(m.uiState.uiTextures["add"].Height)},
		"Add Textures",
	)
	viewResourcesBtn = m.NewIconButton(
		float32(m.window.width-295),
		15,
		40,
		30,
		m.uiState.uiTextures["view"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["view"].Width), Height: float32(m.uiState.uiTextures["view"].Height)},
		"View Textures",
	)

	// These are tool icons
	paintbrushBtn = m.NewIconButton(
		170,
		15,
		40,
		30,
		m.uiState.uiTextures["paintbrush"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["paintbrush"].Width), Height: float32(m.uiState.uiTextures["paintbrush"].Height)},
		"Paintbrush",
	)
	paintbucketBtn = m.NewIconButton(
		220,
		15,
		40,
		30,
		m.uiState.uiTextures["paintbucket"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["paintbucket"].Width), Height: float32(m.uiState.uiTextures["paintbucket"].Height)},
		"Paintbucket",
	)
	eraseBtn = m.NewIconButton(
		270,
		15,
		40,
		30,
		m.uiState.uiTextures["eraser"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["eraser"].Width), Height: float32(m.uiState.uiTextures["eraser"].Height)},
		"Erase",
	)
	selectBtn = m.NewIconButton(
		320,
		15,
		40,
		30,
		m.uiState.uiTextures["select"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["select"].Width), Height: float32(m.uiState.uiTextures["select"].Height)},
		"Select",
	)

	layersBtn = m.NewIconButton(
		370,
		15,
		40,
		30,
		m.uiState.uiTextures["layers"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["layers"].Width), Height: float32(m.uiState.uiTextures["layers"].Height)},
		"Layers",
	)

	locationTooltip := "Player Start"
	switch m.uiState.locationMode {
	case 1:
		locationTooltip = "Dungeon Entrance"
	case 2:
		locationTooltip = "Respawn"
	case 3:
		locationTooltip = "Exit"
	}
	locationBtn = m.NewIconButton(
		420,
		15,
		40,
		30,
		m.uiState.uiTextures["location"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["location"].Width), Height: float32(m.uiState.uiTextures["location"].Height)},
		locationTooltip,
	)

	return
}

// handleTextureSelect handles the selection of a texture from the resource viewer
func (m *MapMaker) handleTextureSelect(texInfo *resources.TextureInfo) {
	m.uiState.activeTexture = texInfo

	// Add to recent textures if not already present
	for i, name := range m.uiState.recentTextures {
		if name == texInfo.Name {
			m.uiState.recentTextures = append(m.uiState.recentTextures[:i], m.uiState.recentTextures[i+1:]...)
			m.uiState.recentTextures = append([]string{texInfo.Name}, m.uiState.recentTextures...)
			return
		}
	}
	m.uiState.recentTextures = append([]string{texInfo.Name}, m.uiState.recentTextures...)

	// Keep only last 8 textures
	if len(m.uiState.recentTextures) > 8 {
		m.uiState.recentTextures = m.uiState.recentTextures[:8]
	}
}

func (m *MapMaker) Close() {
	// Save the config to reopen the last file
	SaveConfig(m.currentFile)
	for _, tex := range m.uiState.uiTextures {
		rl.UnloadTexture(tex)
	}
	if m.resources != nil {
		m.resources.Close()
	}
	rl.CloseWindow()
}
