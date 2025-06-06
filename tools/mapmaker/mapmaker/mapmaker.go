package mapmaker

import (
	"fmt"
	"strconv"
	"strings"

	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
	"github.com/ztkent/beam/controls"
	"github.com/ztkent/beam/resources"
	"github.com/ztkent/beam/tools/spritesheet-viewer/viewer"
)

type MapMaker struct {
	window             *Window
	resources          *resources.ResourceManager
	cm                 *controls.ControlsManager
	uiState            *UIState
	tileGrid           *TileGrid
	currentFile        string
	showResourceViewer bool
	showTileInfo       bool
	showRecentTextures bool
	clipboard          [][]beam.Tile
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
	showGridlines   bool
	// Active toast notification
	toast *Toast

	// Resource Viewer
	resourceViewerScroll   int
	resourceViewerOpenTime float64

	// Tile Info Popup
	tileInfoPos     []beam.Position
	tileInfoPopupX  int32
	tileInfoPopupY  int32
	isDraggingPopup bool
	tileInfoScrollY int32

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

	// Select Tool Swap
	hasSwappedSelect bool

	// Grid Width/Height Controls
	gridWidth  int
	gridHeight int

	// Tile Editor Popup
	textureEditor          *TextureEditorState
	activeInput            string
	showAdvancedEditor     bool
	advancedEditorOpenTime float64

	// NPC Editor State
	npcEditor      *NPCEditorState
	activeNPCInput string
	showNPCList    bool

	// Item Editor State
	itemEditor      *ItemEditorState
	activeItemInput string
	showItemList    bool
}

type TileGrid struct {
	offset               beam.Position    // The offset of the grid in the window
	hasSelection         bool             // If the user has any selected tiles
	selectedTiles        beam.Positions   // These are the tiles that are selected by the user
	missingResourceTiles MissingResources // This is every tile that has a texture, that is missing in the resource manager

	// The section of the grid that is currently visible
	viewportOffset beam.Position // Tracks how many tiles to offset the view
	viewportWidth  int           // Width of visible viewport in tiles
	viewportHeight int           // Height of visible viewport in tiles

	// This is the actual map we will use in game with beam.
	beam.Map
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
	gridSizeX   int32
	gridSizeY   int32
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
	m.uiState.uiTextures["selectall"] = rl.LoadTexture("../assets/grab.png")
	m.uiState.uiTextures["layerwall"] = rl.LoadTexture("../assets/wall.png")
	m.uiState.uiTextures["layerground"] = rl.LoadTexture("../assets/soil.png")
	m.uiState.uiTextures["layers"] = m.uiState.uiTextures["layerground"]
	m.uiState.uiTextures["location"] = rl.LoadTexture("../assets/location.png")
	m.uiState.uiTextures["gridlines"] = rl.LoadTexture("../assets/gridlines.png")
	m.uiState.uiTextures["npc"] = rl.LoadTexture("../assets/npc.png")
	m.uiState.uiTextures["items"] = rl.LoadTexture("../assets/sword.png")

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

		// Clipboard copy
		if rl.IsKeyPressed(rl.KeyC) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyLeftSuper)) {
			if !m.tileGrid.hasSelection || len(m.tileGrid.selectedTiles) == 0 {
				m.showToast("No tiles to copy!", ToastError)
				continue
			}

			// Find bounds of selection
			minX, minY := m.tileGrid.Width, m.tileGrid.Height
			maxX, maxY := 0, 0
			for _, pos := range m.tileGrid.selectedTiles {
				if pos.X < minX {
					minX = pos.X
				}
				if pos.Y < minY {
					minY = pos.Y
				}
				if pos.X > maxX {
					maxX = pos.X
				}
				if pos.Y > maxY {
					maxY = pos.Y
				}
			}

			// Create clipboard array of correct size
			width := maxX - minX + 1
			height := maxY - minY + 1
			m.clipboard = make([][]beam.Tile, height)
			for i := range m.clipboard {
				m.clipboard[i] = make([]beam.Tile, width)
			}

			// Copy selected tiles to clipboard
			for _, pos := range m.tileGrid.selectedTiles {
				relX := pos.X - minX
				relY := pos.Y - minY
				m.clipboard[relY][relX] = m.tileGrid.Tiles[pos.Y][pos.X]
			}

			m.showToast("Tiles copied!", ToastSuccess)
		}

		// Clipboard paste
		if rl.IsKeyPressed(rl.KeyV) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyLeftSuper)) {
			// Verify we have something to paste and somewhere to paste it
			if len(m.clipboard) == 0 || !m.tileGrid.hasSelection {
				m.showToast("Nothing to paste!", ToastError)
				continue
			}

			// Get the target position (first selected tile)
			targetPos := m.tileGrid.selectedTiles[0]

			// Calculate paste bounds
			pasteHeight := len(m.clipboard)
			pasteWidth := len(m.clipboard[0])

			// Iterate through clipboard and paste where possible
			for clipY := 0; clipY < pasteHeight; clipY++ {
				for clipX := 0; clipX < pasteWidth; clipX++ {
					// Calculate target grid position
					gridX := targetPos.X + clipX
					gridY := targetPos.Y + clipY

					// Skip if outside grid bounds
					if gridX >= m.tileGrid.Width || gridY >= m.tileGrid.Height {
						continue
					}

					// Skip if clipboard tile is empty
					if len(m.clipboard[clipY][clipX].Textures) == 0 {
						continue
					}

					// Copy the tile data
					m.tileGrid.Tiles[gridY][gridX] = m.clipboard[clipY][clipX]
					// Update the position to match the new location
					m.tileGrid.Tiles[gridY][gridX].Pos = beam.Position{X: gridX, Y: gridY}
				}
			}

			m.showToast("Tiles pasted!", ToastSuccess)
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

func (m *MapMaker) isUIBlocked() bool {
	return m.showResourceViewer || (m.uiState.textureEditor != nil && m.uiState.textureEditor.visible) || m.uiState.showAdvancedEditor
}

func (m *MapMaker) update() {
	tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, closeMapBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn, gridlinesBtn, npcBtn, itemsBtn := m.getUIButtons()

	// Only handle UI interactions if no modal is blocking
	if !m.isUIBlocked() {
		m.handleResizeGrid(tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn)
		m.handleSaveLoadClose(saveBtn, loadBtn, closeMapBtn)
		m.handleResourceViewer(viewResourcesBtn, loadResourceBtn)
		m.handleMapTools(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn, gridlinesBtn, npcBtn, itemsBtn)

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
			// Check if click is within grid bounds and below menu bar
			if gridX >= 0 && gridX < m.tileGrid.Width &&
				gridY >= 0 && gridY < m.tileGrid.Height &&
				mousePos.Y > float32(m.uiState.menuBarHeight) {
				if m.uiState.selectedTool == "paintbucket" || m.uiState.selectedTool == "selectall" {
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
				(m.uiState.selectedTool == "location" && (m.uiState.locationMode == 1 || m.uiState.locationMode == 3)) {
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

			// Check if we're clicking within the tile info popup (Tile info doesn't block)
			if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && m.showTileInfo {
				dialogWidth := 350
				closeBtn := rl.Rectangle{
					X:      float32(m.uiState.tileInfoPopupX + int32(dialogWidth) - 30),
					Y:      float32(m.uiState.tileInfoPopupY + 5),
					Width:  25,
					Height: 25,
				}
				if rl.CheckCollisionPointRec(mousePos, closeBtn) {
					m.showTileInfo = false
				}
			}

			// Handle right-click actions on selected tiles
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
								beam.NewSimpleTileTexture(m.uiState.activeTexture.Name),
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
							lastTexture := tile.Textures[len(tile.Textures)-1]
							if lastTexture.IsAnimated && len(lastTexture.Frames) > 0 {
								lastTexture.Frames = lastTexture.Frames[:len(lastTexture.Frames)-1]
								if len(lastTexture.Frames) == 0 {
									tile.Textures = tile.Textures[:len(tile.Textures)-1]
								}
							} else {
								tile.Textures = tile.Textures[:len(tile.Textures)-1]
							}
						}
					}
				case "select":
					if !m.showTileInfo {
						// Only show if not already open
						pos := m.tileGrid.selectedTiles[0]
						mousePos := rl.GetMousePosition()
						m.uiState.tileInfoPopupX = int32(mousePos.X)
						m.uiState.tileInfoPopupY = int32(mousePos.Y)
						m.showTileInfo = true
						m.uiState.tileInfoPos = []beam.Position{pos}
					}
				case "selectall":
					if !m.showTileInfo {
						pos := m.tileGrid.selectedTiles
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
						tileType := beam.FloorTile
						if m.uiState.hasSwappedLayers {
							tileType = beam.WallTile
						}
						m.tileGrid.Tiles[selectedY][selectedX].Type = tileType
					}
					break
				case "location":
					// Reset the list if were about to add new positions
					if m.uiState.locationMode == 1 {
						m.tileGrid.DungeonEntry = beam.Positions{}
					} else if m.uiState.locationMode == 3 {
						m.tileGrid.Exit = beam.Positions{}
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
							m.tileGrid.Exit = append(m.tileGrid.Exit, tile)
						}
					}
					break
				case "npc":
					// Initialize NPC editor
					if m.uiState.npcEditor == nil || !m.uiState.npcEditor.visible {
						selectedTile := m.tileGrid.selectedTiles[0] // Use first selected tile
						m.uiState.npcEditor = &NPCEditorState{
							visible:  true,
							spawnPos: selectedTile,
							name:     "New NPC",
							textures: &beam.NPCTexture{
								Up: &beam.AnimatedTexture{
									Frames:     make([]beam.Texture, 0),
									IsAnimated: true,
								},
								Down: &beam.AnimatedTexture{
									Frames:     make([]beam.Texture, 0),
									IsAnimated: true,
								},
								Left: &beam.AnimatedTexture{
									Frames:     make([]beam.Texture, 0),
									IsAnimated: true,
								},
								Right: &beam.AnimatedTexture{
									Frames:     make([]beam.Texture, 0),
									IsAnimated: true,
								},
							},
							health:                 "100",
							attack:                 "10",
							defense:                "5",
							attackSpeed:            "1.0",
							attackRange:            "1.0",
							moveSpeed:              "3.0",
							aggroRange:             "5",
							isHostile:              true,
							editingDirection:       beam.DirDown,
							frameCountStr:          "1",
							animationTimeStr:       "0.5",
							selectedFrames:         make([]string, 0),
							advSelectingFrameIndex: -1,
							selectedFrameIndex:     -1,
						}
					}
					break
				case "items":
					// Initialize item editor
					if m.uiState.itemEditor == nil || !m.uiState.itemEditor.visible {
						selectedTile := m.tileGrid.selectedTiles[0] // Use first selected tile
						m.uiState.itemEditor = &ItemEditorState{
							visible:  true,
							spawnPos: selectedTile,
							name:     "New Item",
							id:       "new_item",
							texture: &beam.AnimatedTexture{
								Frames:     make([]beam.Texture, 0),
								IsAnimated: true,
							},
							maxStack:               "1",
							quantity:               "1",
							attack:                 "0",
							defense:                "0",
							attackSpeed:            "0",
							attackRange:            "0",
							levelReq:               "1",
							itemType:               beam.ItemTypeMisc,
							blocking:               true,
							equippable:             false,
							consumable:             false,
							stackable:              false,
							selectedFrameIndex:     -1,
							frameCountStr:          "1",
							animationTimeStr:       "0.5",
							selectedFrames:         make([]string, 0),
							advSelectingFrameIndex: -1,
						}
					}
					break
				}
			}
		}
	}

	// Handle NPC updates
	for _, npc := range m.tileGrid.NPCs {
		npc.Update(beam.Position{
			X: -1,
			Y: -1,
		}, &m.tileGrid.Map, m.cm)
	}
}

// handleMapTools handles the selecting and swapping of tools
func (m *MapMaker) handleMapTools(paintbrushBtn IconButton, paintbucketBtn IconButton, eraseBtn IconButton, selectBtn IconButton, layersBtn IconButton, locationBtn IconButton, gridlinesBtn IconButton, npcBtn IconButton, itemsBtn IconButton) {
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
			if m.uiState.hasSwappedEraser {
				name = "pencileraser"
			}
			m.uiState.selectedTool = name
			m.showToast(name+" tool selected", ToastInfo)
		}
	}
	if m.isIconButtonClicked(selectBtn) {
		if m.uiState.selectedTool == "select" || m.uiState.selectedTool == "selectall" {
			m.uiState.selectedTool = ""
		} else {
			name := "select"
			if m.uiState.hasSwappedSelect {
				name = "selectall"
			}
			m.uiState.selectedTool = name
			m.tileGrid.hasSelection = false
			m.tileGrid.selectedTiles = beam.Positions{}
			m.showToast(name+" tool selected", ToastInfo)
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
	if m.isIconButtonClicked(gridlinesBtn) {
		m.uiState.showGridlines = !m.uiState.showGridlines
		m.showToast("Gridlines tool selected", ToastInfo)
	}
	if m.isIconButtonClicked(npcBtn) {
		if m.uiState.selectedTool == "npc" {
			m.uiState.selectedTool = ""
		} else {
			m.uiState.selectedTool = "npc"
			m.showToast("NPC Editor tool selected", ToastInfo)
		}
	}
	if m.isIconButtonClicked(itemsBtn) {
		if m.uiState.selectedTool == "items" {
			m.uiState.selectedTool = ""
		} else {
			m.uiState.selectedTool = "items"
			m.showToast("Items Editor tool selected", ToastInfo)
		}
	}

	// Handle tool swaps
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		if m.uiState.rightClickStartTime == 0 {
			m.uiState.rightClickStartTime = rl.GetTime()
		} else if rl.GetTime()-m.uiState.rightClickStartTime > 0.5 {
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

			// Handle select swap
			if m.uiState.selectedTool == "select" || m.uiState.selectedTool == "selectall" {
				m.uiState.uiTextures["select"], m.uiState.uiTextures["selectall"] =
					m.uiState.uiTextures["selectall"], m.uiState.uiTextures["select"]
				if m.uiState.selectedTool == "select" {
					m.uiState.selectedTool = "selectall"
				} else {
					m.uiState.selectedTool = "select"
				}
				m.uiState.hasSwappedSelect = !m.uiState.hasSwappedSelect
			}

			// Handle location swap
			if m.uiState.selectedTool == "location" {
				m.uiState.locationMode = (m.uiState.locationMode + 1) % 4
				modeNames := []string{"Player Start", "Dungeon Entrance", "Respawn", "Exit"}
				m.showToast(fmt.Sprintf("Location Mode: %s", modeNames[m.uiState.locationMode]), ToastInfo)
			}

			// Handle NPC list view swap
			if m.uiState.selectedTool == "npc" {
				m.uiState.showNPCList = true
				m.uiState.rightClickStartTime = 0
			}

			// Handle Item list view swap
			if m.uiState.selectedTool == "items" {
				m.uiState.showItemList = true
				m.uiState.rightClickStartTime = 0
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
		m.uiState.resourceViewerOpenTime = rl.GetTime()
	}
	if m.isIconButtonClicked(loadResourceBtn) {
		name, filepath, isSheet, sheetMargin, gridSizeX, gridSizeY, err := openLoadResourceDialog()
		if err != "" {
			fmt.Println("Error loading resource:", err)
			m.showToast(err, ToastError)
		} else if err := m.loadResource(name, filepath, isSheet, sheetMargin, gridSizeX, gridSizeY); err != nil {
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
			m.tileGrid.Map.NPCs = beam.NPCs{}
			m.tileGrid.Map.Items = beam.Items{}

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
				Textures: make([]*beam.AnimatedTexture, 0),
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
func (m *MapMaker) loadResource(name string, filepath string, isSheet bool, sheetMargin int32, gridSizeX int32, gridSizeY int32) error {
	newRes := resources.Resource{
		Name:        name,
		Path:        filepath,
		IsSheet:     isSheet,
		SheetMargin: sheetMargin,
		GridSizeX:   gridSizeX,
		GridSizeY:   gridSizeY,
	}

	// Display the spritesheet viewer, have user confirm sheet options
	if isSheet {
		finalGridSizeX, finalGridSizeY, finalSheetMargin, err := viewer.ViewSpritesheet(newRes)
		if err != nil {
			return err
		}
		newRes.GridSizeX = finalGridSizeX
		newRes.GridSizeY = finalGridSizeY
		newRes.SheetMargin = finalSheetMargin
	}

	err := m.resources.AddResource("default", newRes)
	if err != nil {
		return err
	}
	m.ValidateTileGrid()
	return nil
}

func (m *MapMaker) getUIButtons() (tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn Button, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, closeMapBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn, gridlinesBtn, npcBtn, itemsButton IconButton) {
	widthSmallerBtn = m.NewButton(10, 8, 30, 20, "-")
	widthLargerBtn = m.NewButton(85, 8, 30, 20, "+")
	heightSmallerBtn = m.NewButton(10, 33, 30, 20, "-")
	heightLargerBtn = m.NewButton(85, 33, 30, 20, "+")
	tileSmallerBtn = m.NewButton(10, 64, 30, 20, "-")
	tileLargerBtn = m.NewButton(85, 64, 30, 20, "+")

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

	eraseText := "Eraser"
	if m.uiState.hasSwappedEraser {
		eraseText = "Pencil Eraser"
	}
	eraseBtn = m.NewIconButton(
		270,
		15,
		40,
		30,
		m.uiState.uiTextures["eraser"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["eraser"].Width), Height: float32(m.uiState.uiTextures["eraser"].Height)},
		eraseText,
	)

	selectText := "Select"
	if m.uiState.hasSwappedSelect {
		selectText = "Select All"
	}
	selectBtn = m.NewIconButton(
		320,
		15,
		40,
		30,
		m.uiState.uiTextures["select"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["select"].Width), Height: float32(m.uiState.uiTextures["select"].Height)},
		selectText,
	)

	layersText := "Ground"
	if m.uiState.hasSwappedLayers {
		layersText = "Wall"
	}
	layersBtn = m.NewIconButton(
		370,
		15,
		40,
		30,
		m.uiState.uiTextures["layers"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["layers"].Width), Height: float32(m.uiState.uiTextures["layers"].Height)},
		layersText,
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

	gridlinesBtn = m.NewIconButton(
		470,
		15,
		40,
		30,
		m.uiState.uiTextures["gridlines"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["gridlines"].Width), Height: float32(m.uiState.uiTextures["gridlines"].Height)},
		"Gridlines",
	)

	npcBtn = m.NewIconButton(
		520,
		15,
		40,
		30,
		m.uiState.uiTextures["npc"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["npc"].Width), Height: float32(m.uiState.uiTextures["npc"].Height)},
		"NPC Editor",
	)

	itemsButton = m.NewIconButton(
		570,
		15,
		40,
		30,
		m.uiState.uiTextures["items"],
		rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["items"].Width), Height: float32(m.uiState.uiTextures["items"].Height)},
		"Item Editor",
	)

	return
}

// handleTextureSelect handles the selection of a texture from the resource viewer
func (m *MapMaker) handleTextureSelect(texInfo *resources.TextureInfo) {
	// Check if selection is for the advanced texture editor frame
	if m.uiState.textureEditor != nil && m.uiState.textureEditor.advSelectingFrameIndex != -1 {
		editor := m.uiState.textureEditor
		frameIndex := editor.advSelectingFrameIndex
		if frameIndex >= 0 && frameIndex < len(editor.advSelectedFrames) {
			editor.advSelectedFrames[frameIndex] = texInfo.Name
		}
		editor.advSelectingFrameIndex = -1 // Reset selection index
		m.showResourceViewer = false       // Close viewer after selection
		return
	}

	// Check if selection is for NPC editor frame
	if m.uiState.npcEditor != nil && m.uiState.npcEditor.visible {
		editor := m.uiState.npcEditor

		frameCount, _ := strconv.Atoi(editor.frameCountStr)
		if frameCount > 0 && editor.advSelectingFrameIndex >= 0 {
			selectedFrame := editor.advSelectingFrameIndex
			// Get the current direction's texture
			var currentTex *beam.AnimatedTexture
			switch editor.editingDirection {
			case beam.DirUp:
				currentTex = editor.textures.Up
			case beam.DirDown:
				currentTex = editor.textures.Down
			case beam.DirLeft:
				currentTex = editor.textures.Left
			case beam.DirRight:
				currentTex = editor.textures.Right
			}
			animationTime, _ := strconv.ParseFloat(editor.animationTimeStr, 64)
			currentTex.AnimationTime = animationTime

			// Update just the selected frame
			if selectedFrame < len(editor.selectedFrames) {
				editor.selectedFrames[selectedFrame] = texInfo.Name

				// Ensure frames array is large enough
				if len(currentTex.Frames) < frameCount {
					newFrames := make([]beam.Texture, frameCount)
					copy(newFrames, currentTex.Frames)
					currentTex.Frames = newFrames
				}

				// Update the specific frame
				currentTex.Frames[selectedFrame] = beam.Texture{
					Name:     texInfo.Name,
					Rotation: 0,
					ScaleX:   1.0,
					ScaleY:   1.0,
					OffsetX:  0,
					OffsetY:  0,
					MirrorX:  false,
					MirrorY:  false,
					Tint:     rl.White,
				}
			}
		}
		editor.advSelectingFrameIndex = -1 // Reset selection index
		m.showResourceViewer = false
		return
	}

	// Add inside handleTextureSelect, after the NPC editor section:
	// Check if selection is for Item editor frame
	if m.uiState.itemEditor != nil && m.uiState.itemEditor.visible {
		editor := m.uiState.itemEditor

		frameCount, _ := strconv.Atoi(editor.frameCountStr)
		if frameCount > 0 && editor.advSelectingFrameIndex >= 0 {
			selectedFrame := editor.advSelectingFrameIndex

			// Initialize texture if needed
			if editor.texture == nil {
				editor.texture = &beam.AnimatedTexture{
					Frames:     make([]beam.Texture, 0),
					IsAnimated: true,
				}
			}

			animationTime, _ := strconv.ParseFloat(editor.animationTimeStr, 64)
			editor.texture.AnimationTime = animationTime

			// Update selected frame
			if selectedFrame < len(editor.selectedFrames) {
				editor.selectedFrames[selectedFrame] = texInfo.Name

				// Ensure frames array is large enough
				if len(editor.texture.Frames) < frameCount {
					newFrames := make([]beam.Texture, frameCount)
					copy(newFrames, editor.texture.Frames)
					editor.texture.Frames = newFrames
				}

				// Update the specific frame
				editor.texture.Frames[selectedFrame] = beam.Texture{
					Name:     texInfo.Name,
					Rotation: 0,
					ScaleX:   1.0,
					ScaleY:   1.0,
					OffsetX:  0,
					OffsetY:  0,
					MirrorX:  false,
					MirrorY:  false,
					Tint:     rl.White,
				}
			}
		}
		editor.advSelectingFrameIndex = -1 // Reset selection index
		m.showResourceViewer = false
		return
	}

	// It was a selection from the main UI
	// Add to recent textures if not already present
	m.uiState.activeTexture = texInfo
	for i, name := range m.uiState.recentTextures {
		if name == texInfo.Name {
			// Move to front
			m.uiState.recentTextures = append(m.uiState.recentTextures[:i], m.uiState.recentTextures[i+1:]...)
			m.uiState.recentTextures = append([]string{texInfo.Name}, m.uiState.recentTextures...)
			return
		}
	}
	// Add to front
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
