package mapmaker

import (
	"fmt"
	"log" // Use log package for better logging
	"math"
	"os"
	"path/filepath"
	"strings"
	"time" // For debouncing/timing

	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
	"github.com/ztkent/beam/resources"
	"github.com/ztkent/beam/tools/spritesheet-viewer/viewer" // Assuming this dependency exists and works
)

// --- Undo/Redo Action Interface ---
type UndoableAction interface {
	Do(m *MapMaker)   // Apply the action
	Undo(m *MapMaker) // Revert the action
	Description() string // For potential history view
}

// --- Concrete Action Implementations ---

// TileChangeAction handles changes to textures and type for multiple tiles
type TileChangeAction struct {
	Positions []beam.Position
	OldTiles  []beam.Tile // Store complete old tile state
	NewTiles  []beam.Tile // Store complete new tile state (or relevant changes)
	desc      string
}

func NewTileChangeAction(positions []beam.Position, oldTiles []beam.Tile, newTiles []beam.Tile, description string) *TileChangeAction {
	// Ensure slices are copied, not referenced directly if they might change
	oldCopy := make([]beam.Tile, len(oldTiles))
	copy(oldCopy, oldTiles)
	newCopy := make([]beam.Tile, len(newTiles))
	copy(newCopy, newTiles)

	posCopy := make([]beam.Position, len(positions))
	copy(posCopy, positions)

	return &TileChangeAction{
		Positions: posCopy,
		OldTiles:  oldCopy,
		NewTiles:  newCopy,
		desc:      description,
	}
}

func (a *TileChangeAction) Do(m *MapMaker) {
	for i, pos := range a.Positions {
		if pos.Y >= 0 && pos.Y < len(m.tileGrid.Tiles) && pos.X >= 0 && pos.X < len(m.tileGrid.Tiles[pos.Y]) {
			// Ensure texture slice capacity if needed when applying NewTiles state directly
			if len(m.tileGrid.Tiles[pos.Y][pos.X].Textures) < len(a.NewTiles[i].Textures) {
				m.tileGrid.Tiles[pos.Y][pos.X].Textures = make([]beam.TileTexture, len(a.NewTiles[i].Textures))
			}
			m.tileGrid.Tiles[pos.Y][pos.X] = a.NewTiles[i] // Overwrite with the new state
		}
	}
	m.ValidateTileGrid() // Re-validate after applying changes
}

func (a *TileChangeAction) Undo(m *MapMaker) {
	for i, pos := range a.Positions {
		if pos.Y >= 0 && pos.Y < len(m.tileGrid.Tiles) && pos.X >= 0 && pos.X < len(m.tileGrid.Tiles[pos.Y]) {
			// Ensure texture slice capacity if needed
			if len(m.tileGrid.Tiles[pos.Y][pos.X].Textures) < len(a.OldTiles[i].Textures) {
				m.tileGrid.Tiles[pos.Y][pos.X].Textures = make([]beam.TileTexture, len(a.OldTiles[i].Textures))
			}
			m.tileGrid.Tiles[pos.Y][pos.X] = a.OldTiles[i] // Restore the old state
		}
	}
	m.ValidateTileGrid() // Re-validate after reverting changes
}

func (a *TileChangeAction) Description() string {
	return a.desc
}

// --- End Undo/Redo ---

type MapMaker struct {
	window             *Window
	resources          *resources.ResourceManager
	uiState            *UIState
	tileGrid           *TileGrid
	currentFile        string
	showResourceViewer bool
	showTileInfo       bool
	showRecentTextures bool

	// Undo/Redo Stacks
	undoStack []UndoableAction
	redoStack []UndoableAction

	// Panning State
	isPanning       bool
	panStartMouse   rl.Vector2
	panStartViewport beam.Position

	// Performance/Debug
	lastRenderTime float32
	lastUpdateTime float32
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
	rightClickStartTime float64 // Using float64 from GetTime()

	// Eraser Tool Swap
	hasSwappedEraser bool

	// Layers Tool Swap
	hasSwappedLayers bool

	// Location Tool Mode
	locationMode int

	// Grid Width/Height Controls
	gridWidth  int
	gridHeight int

	// Tile Editor Popup
	textureEditor *TextureEditorState // Assuming this exists elsewhere
	activeInput   string              // Assuming this relates to textureEditor

	// --- New UI State Fields ---
	zoomLevel     float32 // Current zoom multiplier
	currentLayer  int     // Active layer index for editing
	maxLayers     int     // Maximum layers allowed
	showGridLines bool    // Toggle for grid lines
}

type TileGrid struct {
	offset               beam.Position    // The visual offset of the grid top-left corner in the window (pixels)
	hasSelection         bool
	selectedTiles        beam.Positions   // These are the tiles that are selected by the user (grid coordinates)
	missingResourceTiles MissingResources // This is every tile that has a texture, that is missing in the resource manager

	// This is the actual map we will use in game with beam.
	beam.Map // Embed the beam.Map struct directly

	// Viewport tracking (in tile coordinates)
	viewportOffset beam.Position // Top-left tile coordinate of the visible viewport
	// Viewport dimensions are now calculated dynamically based on window size and zoom
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
	WidthGutter       = 150 // Added some space for potential sidebars/palettes
	HeightGutter      = 150
	DefaultTileSize   = 20 // Base size at zoomLevel 1.0
	DefaultGridWidth  = 64
	DefaultGridHeight = 40
	// MaxDisplayWidth and Height removed - viewport is dynamic

	// --- New Constants ---
	MinZoom           = 0.2
	MaxZoom           = 5.0
	ZoomIncrement     = 0.1
	PanSpeed          = 500.0 // Pixels per second for keyboard panning
	MaxUndoHistory    = 100   // Limit undo stack size
	DefaultMaxLayers  = 4     // Default number of texture layers
)

// ResourceDialog definition remains unchanged
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
			// Adjust initial size slightly if needed
			width:  1280 + WidthGutter,
			height: 800 + HeightGutter,
			title:  "Beam MapMaker Enhanced", // New title!
		},
		uiState: &UIState{
			tileSize:        DefaultTileSize, // Effective size based on zoom
			gridWidth:       DefaultGridWidth,
			gridHeight:      DefaultGridHeight,
			menuBarHeight:   60,
			statusBarHeight: 25,
			uiTextures:      make(map[string]rl.Texture2D),
			activeTexture:   nil,
			selectedTool:    "", // Start with no tool selected
			toast:           nil,
			recentTextures:  make([]string, 0),

			resourceManageMode: false,
			hasSwappedEraser:   false,
			hasSwappedLayers:   false,
			locationMode:       0,

			// --- New State Initialization ---
			zoomLevel:     1.0,
			currentLayer:  0, // Start editing layer 0
			maxLayers:     DefaultMaxLayers,
			showGridLines: true,
		},
		tileGrid: &TileGrid{
			offset:               beam.Position{X: 0, Y: 0}, // Offset is calculated dynamically
			selectedTiles:        beam.Positions{{X: -1, Y: -1}},
			hasSelection:         false,
			viewportOffset:       beam.Position{X: 0, Y: 0},
			missingResourceTiles: make(MissingResources, 0),
			// beam.Map is initialized within initTileGrid
		},
		currentFile: "",
		undoStack:   make([]UndoableAction, 0, MaxUndoHistory),
		redoStack:   make([]UndoableAction, 0, MaxUndoHistory),
	}
	// mm.updateGridSize() // No longer needed here, grid size is part of beam.Map
	return mm
}

func (m *MapMaker) Init() {
	rl.SetConfigFlags(rl.FlagWindowResizable) // Allow window resizing
	rl.InitWindow(m.window.width, m.window.height, m.window.title)
	rl.SetWindowMinSize(800, 600) // Set a minimum size
	rl.SetTargetFPS(60)
	rl.SetExitKey(0) // Disable Escape key for closing window directly

	// Load UI textures (same as before)
	m.uiState.uiTextures["add"] = rl.LoadTexture("../assets/add.png")
	m.uiState.uiTextures["view"] = rl.LoadTexture("../assets/view.png")
	m.uiState.uiTextures["save"] = rl.LoadTexture("../assets/save.png")
	m.uiState.uiTextures["load"] = rl.LoadTexture("../assets/load.png")
	m.uiState.uiTextures["close"] = rl.LoadTexture("../assets/reset.png")
	m.uiState.uiTextures["paintbrush"] = rl.LoadTexture("../assets/paintbrush.png")
	m.uiState.uiTextures["paintbucket"] = rl.LoadTexture("../assets/paintbucket.png")
	m.uiState.uiTextures["eraser"] = rl.LoadTexture("../assets/eraser.png")
	m.uiState.uiTextures["pencileraser"] = rl.LoadTexture("../assets/pencileraser.png")
	m.uiState.uiTextures["select"] = rl.LoadTexture("../assets/select.png")
	m.uiState.uiTextures["layerwall"] = rl.LoadTexture("../assets/wall.png")
	m.uiState.uiTextures["layerground"] = rl.LoadTexture("../assets/soil.png")
	m.uiState.uiTextures["layers"] = m.uiState.uiTextures["layerground"] // Initial state
	m.uiState.uiTextures["location"] = rl.LoadTexture("../assets/location.png")
	m.uiState.uiTextures["up"] = rl.LoadTexture("../assets/up.png") // Keep viewport arrows for now? Maybe remove later if pan is good.
	m.uiState.uiTextures["down"] = rl.LoadTexture("../assets/down.png")
	m.uiState.uiTextures["left"] = rl.LoadTexture("../assets/left.png")
	m.uiState.uiTextures["right"] = rl.LoadTexture("../assets/right.png")
	// --- New UI Textures ---
	m.uiState.uiTextures["undo"] = rl.LoadTexture("../assets/undo.png") // Assuming you have these icons
	m.uiState.uiTextures["redo"] = rl.LoadTexture("../assets/redo.png")
	m.uiState.uiTextures["grid"] = rl.LoadTexture("../assets/grid.png") // Toggle grid icon

	m.resources = resources.NewResourceManager()
	m.initTileGrid(DefaultGridWidth, DefaultGridHeight) // Initialize with default size
	m.loadConfig()                                      // Load last file path
}

// --- Action Recording Helper ---
func (m *MapMaker) recordAction(action UndoableAction) {
	m.undoStack = append(m.undoStack, action)
	// Limit undo history
	if len(m.undoStack) > MaxUndoHistory {
		m.undoStack = m.undoStack[len(m.undoStack)-MaxUndoHistory:]
	}
	// Clear redo stack whenever a new action is performed
	m.redoStack = make([]UndoableAction, 0, MaxUndoHistory)
	log.Printf("Action Recorded: %s (Undo stack size: %d)", action.Description(), len(m.undoStack))
}

// --- Undo/Redo Functions ---
func (m *MapMaker) undoLastAction() {
	if len(m.undoStack) > 0 {
		action := m.undoStack[len(m.undoStack)-1]
		m.undoStack = m.undoStack[:len(m.undoStack)-1]
		action.Undo(m)
		m.redoStack = append(m.redoStack, action)
		log.Printf("Action Undone: %s (Undo: %d, Redo: %d)", action.Description(), len(m.undoStack), len(m.redoStack))
		m.showToast("Undo: "+action.Description(), ToastInfo)
		m.clearSelection() // Clear selection after undo/redo usually makes sense
	} else {
		m.showToast("Nothing to undo", ToastWarning)
	}
}

func (m *MapMaker) redoLastAction() {
	if len(m.redoStack) > 0 {
		action := m.redoStack[len(m.redoStack)-1]
		m.redoStack = m.redoStack[:len(m.redoStack)-1]
		action.Do(m) // Redo the action
		m.undoStack = append(m.undoStack, action)
		log.Printf("Action Redone: %s (Undo: %d, Redo: %d)", action.Description(), len(m.undoStack), len(m.redoStack))
		m.showToast("Redo: "+action.Description(), ToastInfo)
		m.clearSelection() // Clear selection after undo/redo
	} else {
		m.showToast("Nothing to redo", ToastWarning)
	}
}

func (m *MapMaker) Run() {
	defer m.Close() // Ensure Close is called on exit

	for !rl.WindowShouldClose() { // Use standard exit check
		startTime := time.Now()

		// --- Input Handling ---
		m.handleInput()

		// --- Update Logic ---
		m.update() // Update settings, configs, and UI state.

		updateDuration := time.Since(startTime)
		m.lastUpdateTime = float32(updateDuration.Seconds()) * 1000 // ms

		// --- Drawing ---
		drawStartTime := time.Now()
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		m.renderGrid()  // Render the current map (culled)
		m.renderUI()    // Render the UI
		m.renderToast() // Render any active toasts

		rl.EndDrawing()
		renderDuration := time.Since(drawStartTime)
		m.lastRenderTime = float32(renderDuration.Seconds()) * 1000 // ms
	}
}

// handleInput processes all keyboard and mouse input
func (m *MapMaker) handleInput() {
	mousePos := rl.GetMousePosition()
	mouseWheel := rl.GetMouseWheelMove()

	// --- Global Shortcuts ---
	isCtrlDown := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) || rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper)

	if rl.IsKeyPressed(rl.KeyEscape) {
		if m.showResourceViewer || m.showTileInfo || m.showRecentTextures {
			m.showResourceViewer = false
			m.showTileInfo = false
			m.showRecentTextures = false
			// Add closing logic for other potential popups/dialogs here
		} else if m.tileGrid.hasSelection {
			m.clearSelection()
		} else if m.uiState.selectedTool != "" {
			m.selectTool("") // Deselect tool
		} else {
			// Consider adding a confirmation dialog before closing if changes are unsaved
			// For now, let the main loop handle rl.WindowShouldClose() logic
			// rl.CloseWindow() // Avoid direct close here
		}
	}

	if isCtrlDown && rl.IsKeyPressed(rl.KeyS) { // Save
		m.handleSaveAction()
	}
	if isCtrlDown && rl.IsKeyPressed(rl.KeyO) { // Load
		m.handleLoadAction()
	}
	if isCtrlDown && rl.IsKeyPressed(rl.KeyZ) { // Undo
		m.undoLastAction()
	}
	if isCtrlDown && rl.IsKeyPressed(rl.KeyY) { // Redo
		m.redoLastAction()
	}

	// --- Tool Selection Shortcuts ---
	if !isCtrlDown { // Avoid conflict with save/load etc.
		if rl.IsKeyPressed(rl.KeyP) {
			m.selectTool("paintbrush")
		}
		if rl.IsKeyPressed(rl.KeyB) {
			m.selectTool("paintbucket")
		}
		if rl.IsKeyPressed(rl.KeyE) {
			m.selectTool("eraser") // Selects current eraser type
		}
		if rl.IsKeyPressed(rl.KeyS) {
			m.selectTool("select")
		}
		if rl.IsKeyPressed(rl.KeyL) {
			m.selectTool("layers")
		}
		if rl.IsKeyPressed(rl.KeyO) { // Changed from location to avoid conflict with Load
			m.selectTool("location")
		}
		// Add shortcuts for layer cycling
		if rl.IsKeyPressed(rl.KeyPageUp) {
			m.changeLayer(1)
		}
		if rl.IsKeyPressed(rl.KeyPageDown) {
			m.changeLayer(-1)
		}
	}

	// --- Zoom ---
	if mouseWheel != 0 && !m.isMouseOverUI(mousePos) { // Don't zoom if mouse is over UI elements
		m.handleZoom(mouseWheel, mousePos)
	}

	// --- Panning ---
	m.handlePanning(mousePos)

	// --- Grid Interaction (Left Click Drag/Select) ---
	m.handleGridInteraction(mousePos)

	// --- Right Click Actions (Apply tool / Swap tool) ---
	m.handleRightClickActions(mousePos)
}

func (m *MapMaker) update() {
	// --- Window Resizing ---
	if rl.IsWindowResized() {
		m.window.width = int32(rl.GetScreenWidth())
		m.window.height = int32(rl.GetScreenHeight())
		log.Printf("Window resized to: %d x %d", m.window.width, m.window.height)
		// Recalculate grid offset or other layout elements if needed
	}

	// --- UI Button Logic (Example - Needs integration with actual drawing/checking) ---
	// Note: Button checking should happen *after* input handling potentially changes state
	// This section might be better integrated directly into handleInput or renderUI
	// For simplicity, keeping button *definition* here but click logic in input handlers.
	// tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, closeMapBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn := m.getUIButtons()
	// ... handle button clicks from handleInput or directly in renderUI ...

	// Center the grid based on *visible* dimensions and zoom
	m.updateGridOffset()

	// Update Toast
	if m.uiState.toast != nil {
		m.uiState.toast.duration -= rl.GetFrameTime()
		if m.uiState.toast.duration <= 0 {
			m.uiState.toast = nil
		}
	}

	// Update Tile Info Popup Dragging
	if m.uiState.isDraggingPopup {
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			delta := rl.GetMouseDelta()
			m.uiState.tileInfoPopupX += int32(delta.X)
			m.uiState.tileInfoPopupY += int32(delta.Y)
		} else {
			m.uiState.isDraggingPopup = false
		}
	}

	// Check if Resource Viewer needs closing
	if m.showResourceViewer && !m.uiState.resourceManageMode { // Close if not in manage mode
		// Check if clicked outside viewer bounds (implement isMouseOverResourceViewer)
		// if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && !m.isMouseOverResourceViewer(rl.GetMousePosition()) {
		// 	m.showResourceViewer = false
		// }
	}

}

// updateGridOffset calculates the visual top-left pixel offset for drawing the grid
func (m *MapMaker) updateGridOffset() {
	// Calculate the total pixel dimensions of the *entire* grid at current zoom
	// totalPixelWidth := float32(m.tileGrid.Width) * float32(m.uiState.tileSize) * m.uiState.zoomLevel
	// totalPixelHeight := float32(m.tileGrid.Height) * float32(m.uiState.tileSize) * m.uiState.zoomLevel

	// Calculate the dimensions of the visible drawing area
	workspaceWidth := float32(m.window.width)
	workspaceHeight := float32(m.window.height) - float32(m.uiState.menuBarHeight) - float32(m.uiState.statusBarHeight)

	// Simplified centering (doesn't account for panning fully yet, needs refinement)
	// We need to calculate offset based on viewportOffset (tile coords) and zoom level
	effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel

	// Calculate how many tiles fit on screen
	viewportWidthTiles := int(math.Ceil(float64(workspaceWidth / effectiveTileSize)))
	viewportHeightTiles := int(math.Ceil(float64(workspaceHeight / effectiveTileSize)))

	// Calculate total pixel size of the *visible* part of the grid based on tile viewport
	visibleGridPixelWidth := float32(viewportWidthTiles) * effectiveTileSize
	visibleGridPixelHeight := float32(viewportHeightTiles) * effectiveTileSize

	// Center this potentially oversized visible area within the workspace
	// This isn't quite right for panning - offset should be determined by viewportOffset
	// Let's calculate offset based purely on the top-left viewport tile coordinate
	offsetX := workspaceWidth/2 - (float32(m.tileGrid.viewportOffset.X) * effectiveTileSize) - effectiveTileSize/2
	offsetY := float32(m.uiState.menuBarHeight) + workspaceHeight/2 - (float32(m.tileGrid.viewportOffset.Y) * effectiveTileSize) - effectiveTileSize/2

	// This needs more work - the offset should be relative to the window top-left (0,0)
	// Offset X = Center Workspace - (Center of Viewport in World Pixels)
	// Center of Viewport = (viewportOffset.X + viewportWidthTiles/2) * effectiveTileSize
	// Let's try a simpler approach: Offset is the pixel position of tile (0,0) relative to window top-left

	// Pixel position of viewport top-left tile relative to window top-left
	viewportPixelX := -float32(m.tileGrid.viewportOffset.X) * effectiveTileSize
	viewportPixelY := -float32(m.tileGrid.viewportOffset.Y) * effectiveTileSize

	// We want to center the viewport within the workspace area
	// The center of the workspace is:
	centerX := workspaceWidth / 2
	centerY := float32(m.uiState.menuBarHeight) + workspaceHeight/2

	// The center of the current viewport view (in pixels relative to grid 0,0) is approx:
	// viewCenterX := (float32(viewportWidthTiles) / 2.0) * effectiveTileSize
	// viewCenterY := (float32(viewportHeightTiles) / 2.0) * effectiveTileSize

	// Offset required to place viewCenterX at centerX:
	// This still feels overly complex. Let's stick to the simpler definition for now:
	// Offset = pixel coordinate where tile (0,0) should be drawn.
	// When panning, we adjust viewportOffset (tile coordinates).
	// The rendering loop will then use viewportOffset to draw the correct tiles.
	// The visual *offset* of the grid container itself can be calculated based on where the
	// *currently visible* top-left tile (viewportOffset) should appear.

	// Pixel position of the viewport top-left tile (m.tileGrid.viewportOffset)
	// relative to the window's top-left origin (0,0).
	m.tileGrid.offset.X = int(-float32(m.tileGrid.viewportOffset.X) * effectiveTileSize)
	m.tileGrid.offset.Y = int(float32(m.uiState.menuBarHeight) - float32(m.tileGrid.viewportOffset.Y)*effectiveTileSize)

	// This isn't centering. Recalculate centering approach:
	// The top-left corner of the workspace area:
	workAreaX := int32(0)
	workAreaY := int32(m.uiState.menuBarHeight)
	workAreaWidth := m.window.width
	workAreaHeight := m.window.height - int32(m.uiState.menuBarHeight) - int32(m.uiState.statusBarHeight)

	// Calculate the pixel coordinate of the top-left visible tile (viewportOffset)
	// This calculation determines *where* the rendering loop starts drawing relative
	// to the window coordinates.
	// Let grid drawing start at (workAreaX, workAreaY).
	// The first tile drawn will be at viewportOffset.
	// Its position should be (workAreaX, workAreaY). No, that's not right.

	// Let's simplify: The `offset` will be the top-left pixel coordinate of the
	// visible grid area within the window. Panning adjusts `viewportOffset` (tiles).
	// Rendering calculates which tiles to draw based on `viewportOffset` and window size/zoom.
	// The `renderGrid` function will calculate the screen position for each visible tile.
	// We don't strictly need a single top-level grid `offset` if rendering calculates per tile.

	// Let's keep `offset` as the calculated *visual* top-left for the *entire potential grid* if it were centered.
	// This might only be useful for drawing a background rect maybe.
	totalGridWidth := int(float32(m.tileGrid.Width) * effectiveTileSize)
	totalGridHeight := int(float32(m.tileGrid.Height) * effectiveTileSize)

	m.tileGrid.offset = beam.Position{
		X: int(workAreaX) + (int(workAreaWidth) - totalGridWidth) / 2, // Center the whole grid conceptually
		Y: int(workAreaY) + (int(workAreaHeight) - totalGridHeight) / 2,
	}
	// IMPORTANT: This offset is less useful now. Rendering must calculate screen pos based on viewportOffset.

}

// getVisibleTileBounds calculates the start and end tile indices to render based on viewport and window size
func (m *MapMaker) getVisibleTileBounds() (startX, startY, endX, endY int) {
	effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel
	if effectiveTileSize <= 0 {
		effectiveTileSize = 1 // Avoid division by zero
	}

	// Workspace area where grid is drawn
	workAreaX := float32(0)
	workAreaY := float32(m.uiState.menuBarHeight)
	workAreaWidth := float32(m.window.width)
	workAreaHeight := float32(m.window.height) - workAreaY - float32(m.uiState.statusBarHeight)

	// Calculate tile indices corresponding to the corners of the workspace area
	// Adjusting for viewportOffset which defines the top-left tile *in the world* at the top-left of the view.
	startX = m.tileGrid.viewportOffset.X + int(math.Floor(float64((0 - workAreaX) / effectiveTileSize))) // Relative to viewport start
	startY = m.tileGrid.viewportOffset.Y + int(math.Floor(float64((0 - workAreaY + float32(m.uiState.menuBarHeight)) / effectiveTileSize))) // This calculation seems off.

	// Simpler: Start tile is viewportOffset. End tile is viewportOffset + number of tiles that fit.
	startX = m.tileGrid.viewportOffset.X
	startY = m.tileGrid.viewportOffset.Y

	// Add a buffer of 1 tile around the edges for smoother panning/zooming
	buffer := 1
	endX = startX + int(math.Ceil(float64(workAreaWidth/effectiveTileSize))) + buffer*2
	endY = startY + int(math.Ceil(float64(workAreaHeight/effectiveTileSize))) + buffer*2

	// Clamp to grid boundaries
	startX = max(0, startX-buffer)
	startY = max(0, startY-buffer)
	endX = min(m.tileGrid.Width, endX)
	endY = min(m.tileGrid.Height, endY)

	// Adjust start based on buffer clamping near 0
	// This isn't strictly necessary if clamping happens correctly.

	return startX, startY, endX, endY
}

// getTileScreenPosition calculates the top-left pixel coordinate for a given tile index
func (m *MapMaker) getTileScreenPosition(tileX, tileY int) rl.Vector2 {
	effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel

	// Position relative to the viewport's top-left tile's screen position
	screenX := float32(m.uiState.menuBarHeight) + float32(tileX-m.tileGrid.viewportOffset.X)*effectiveTileSize // Error here, X depends on workAreaX
	screenY := float32(m.uiState.menuBarHeight) + float32(tileY-m.tileGrid.viewportOffset.Y)*effectiveTileSize // Y offset starts below menu bar

	// Let's rethink: Calculate relative to the workspace top-left (0, menuBarHeight)
	screenX = 0 + float32(tileX-m.tileGrid.viewportOffset.X)*effectiveTileSize
	screenY = float32(m.uiState.menuBarHeight) + float32(tileY-m.tileGrid.viewportOffset.Y)*effectiveTileSize

	return rl.NewVector2(screenX, screenY)
}

// getTileCoordsFromMouse converts mouse position to grid coordinates
func (m *MapMaker) getTileCoordsFromMouse(mousePos rl.Vector2) (int, int, bool) {
	effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel
	if effectiveTileSize <= 0 {
		return -1, -1, false // Invalid state
	}

	// Adjust mouse position relative to the workspace origin (0, menuBarHeight)
	relativeMouseX := mousePos.X - 0
	relativeMouseY := mousePos.Y - float32(m.uiState.menuBarHeight)

	// Calculate grid coordinates based on viewport offset
	gridX := m.tileGrid.viewportOffset.X + int(math.Floor(float64(relativeMouseX/effectiveTileSize)))
	gridY := m.tileGrid.viewportOffset.Y + int(math.Floor(float64(relativeMouseY/effectiveTileSize)))

	// Check if the calculated coordinates are within the map bounds
	isValid := gridX >= 0 && gridX < m.tileGrid.Width && gridY >= 0 && gridY < m.tileGrid.Height && mousePos.Y >= float32(m.uiState.menuBarHeight)

	return gridX, gridY, isValid
}

// handleGridInteraction processes left mouse button clicks and drags on the grid
func (m *MapMaker) handleGridInteraction(mousePos rl.Vector2) {
	// Only interact if mouse is not over UI and certain popups are closed
	if m.isMouseOverUI(mousePos) || m.showResourceViewer || m.showTileInfo { // Add other modal checks if needed
		return
	}

	gridX, gridY, isValid := m.getTileCoordsFromMouse(mousePos)

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if isValid {
			m.clearSelection() // Clear previous selection on new click
			newPos := beam.Position{X: gridX, Y: gridY}

			if m.uiState.selectedTool == "paintbucket" {
				// Flood fill selection needs to be done carefully with undo
				// For now, skip undo for flood fill or implement complex flood fill action
				m.tileGrid.selectedTiles = m.floodFillSelection(gridX, gridY)
				m.tileGrid.hasSelection = true
				// Don't record flood fill selection itself, only the application of texture/tool
			} else if m.uiState.selectedTool != "" { // Only select if a tool is active
				m.tileGrid.selectedTiles = beam.Positions{newPos}
				m.tileGrid.hasSelection = true
			}
		} else {
			m.clearSelection() // Clicked outside grid or on UI
		}
	} else if rl.IsMouseButtonDown(rl.MouseLeftButton) && m.tileGrid.hasSelection {
		// Handle dragging for selection expansion (for specific tools)
		if isValid && (m.uiState.selectedTool == "paintbrush" ||
			m.uiState.selectedTool == "eraser" ||
			m.uiState.selectedTool == "pencileraser" ||
			m.uiState.selectedTool == "layers" ||
			(m.uiState.selectedTool == "location" && m.uiState.locationMode == 1)) { // Allow multi-select for dungeon entry

			newPos := beam.Position{X: gridX, Y: gridY}
			alreadySelected := slices.ContainsFunc(m.tileGrid.selectedTiles, func(p beam.Position) bool {
				return p == newPos
			})

			if !alreadySelected {
				m.tileGrid.selectedTiles = append(m.tileGrid.selectedTiles, newPos)
				// Drag selection itself isn't typically undoable, the application of the tool is.
			}
		}
	}
}

// handleRightClickActions processes right mouse button clicks (apply tool, swap tool)
func (m *MapMaker) handleRightClickActions(mousePos rl.Vector2) {
	// --- Tool Application (Right Mouse Press) ---
	if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		m.uiState.rightClickStartTime = rl.GetTime() // Record start time for potential hold

		if m.isMouseOverUI(mousePos) { // Don't apply tool if over UI
			return
		}

		// If a selection exists, apply the current tool to the selected tiles
		if m.tileGrid.hasSelection {
			m.applySelectedTool(m.tileGrid.selectedTiles)
			// Right-click apply usually doesn't clear selection immediately, user might want to apply again
		} else {
			// If no selection, apply tool to the single tile under cursor (if valid)
			gridX, gridY, isValid := m.getTileCoordsFromMouse(mousePos)
			if isValid && m.uiState.selectedTool != "" && m.uiState.selectedTool != "select" {
				singleTilePos := []beam.Position{{X: gridX, Y: gridY}}
				m.applySelectedTool(singleTilePos)
			} else if isValid && m.uiState.selectedTool == "select" {
				// Show tile info popup on right click with select tool
				m.showTileInfoPopup(beam.Position{X: gridX, Y: gridY}, mousePos)
			}
		}
	}

	// --- Tool Swapping (Right Mouse Hold) ---
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		// Check if held long enough, but only trigger swap once per hold
		if m.uiState.rightClickStartTime > 0 && (rl.GetTime()-m.uiState.rightClickStartTime) > 0.5 {
			swapped := false
			// Eraser Swap
			if m.uiState.selectedTool == "eraser" || m.uiState.selectedTool == "pencileraser" {
				m.swapEraserTool()
				swapped = true
			}
			// Layers Swap
			if m.uiState.selectedTool == "layers" {
				m.swapLayersTool()
				swapped = true
			}
			// Location Swap
			if m.uiState.selectedTool == "location" {
				m.swapLocationTool()
				swapped = true
			}

			if swapped {
				m.uiState.rightClickStartTime = 0 // Reset timer after swap to prevent rapid swapping
			}
		}
	} else {
		// Reset start time when button is released
		m.uiState.rightClickStartTime = 0
	}
}

// applySelectedTool applies the currently selected tool's logic to the given positions
func (m *MapMaker) applySelectedTool(positions beam.Positions) {
	if len(positions) == 0 {
		return
	}

	oldTiles := make([]beam.Tile, len(positions))
	newTiles := make([]beam.Tile, len(positions))
	changed := false
	actionDesc := ""

	for i, pos := range positions {
		if !(pos.Y >= 0 && pos.Y < len(m.tileGrid.Tiles) && pos.X >= 0 && pos.X < len(m.tileGrid.Tiles[pos.Y])) {
			continue // Skip invalid positions if any somehow got selected
		}
		originalTile := m.tileGrid.Tiles[pos.Y][pos.X]
		oldTiles[i] = originalTile // Store original state for undo

		currentTile := originalTile // Work on a copy for modification

		switch m.uiState.selectedTool {
		case "paintbrush":
			actionDesc = "Paint Texture"
			if m.uiState.activeTexture != nil {
				newTexture := beam.NewSimpleTileTexture(m.uiState.activeTexture.Name)
				// Ensure layer exists
				for len(currentTile.Textures) <= m.uiState.currentLayer {
					currentTile.Textures = append(currentTile.Textures, beam.TileTexture{Name: "", IsComplex: false}) // Add empty placeholder layers
				}
				// Only change if different
				if m.uiState.currentLayer < len(currentTile.Textures) {
					if currentTile.Textures[m.uiState.currentLayer].Name != newTexture.Name ||
						currentTile.Textures[m.uiState.currentLayer].IsComplex != newTexture.IsComplex {
						currentTile.Textures[m.uiState.currentLayer] = newTexture
						changed = true
					}
				} else { // Should not happen with the ensure loop, but safeguard
					currentTile.Textures = append(currentTile.Textures, newTexture)
					changed = true
				}
			}
		case "paintbucket":
			// Apply to all tiles selected by flood fill (which are in 'positions')
			actionDesc = "Fill Texture"
			if m.uiState.activeTexture != nil {
				newTexture := beam.NewSimpleTileTexture(m.uiState.activeTexture.Name)
				for len(currentTile.Textures) <= m.uiState.currentLayer {
					currentTile.Textures = append(currentTile.Textures, beam.TileTexture{Name: "", IsComplex: false})
				}
				if m.uiState.currentLayer < len(currentTile.Textures) {
					if currentTile.Textures[m.uiState.currentLayer].Name != newTexture.Name ||
						currentTile.Textures[m.uiState.currentLayer].IsComplex != newTexture.IsComplex {
						currentTile.Textures[m.uiState.currentLayer] = newTexture
						changed = true
					}
				} else {
					currentTile.Textures = append(currentTile.Textures, newTexture)
					changed = true
				}
			}
		case "eraser":
			actionDesc = "Erase Layer"
			// Erase only the current layer
			if m.uiState.currentLayer < len(currentTile.Textures) {
				// Replace with an "empty" texture or remove? Let's replace with empty name.
				if currentTile.Textures[m.uiState.currentLayer].Name != "" {
					currentTile.Textures[m.uiState.currentLayer] = beam.TileTexture{Name: "", IsComplex: false}
					changed = true
				}
			}
		case "pencileraser":
			actionDesc = "Erase Top Texture"
			// Erase the topmost texture layer only (regardless of currentLayer setting)
			if len(currentTile.Textures) > 0 {
				// Check if the top texture is already empty
				if currentTile.Textures[len(currentTile.Textures)-1].Name != "" {
					// If it's complex with frames, remove last frame, else remove whole texture entry
					lastTexture := &currentTile.Textures[len(currentTile.Textures)-1] // Get pointer
					if lastTexture.IsComplex && len(lastTexture.Frames) > 0 {
						lastTexture.Frames = lastTexture.Frames[:len(lastTexture.Frames)-1] // Modify pointer content
						if len(lastTexture.Frames) == 0 {
							// If removing last frame makes it empty, remove the texture entry
							currentTile.Textures = currentTile.Textures[:len(currentTile.Textures)-1]
						}
					} else {
						// Simple texture or empty complex one, remove the entry
						currentTile.Textures = currentTile.Textures[:len(currentTile.Textures)-1]
					}
					changed = true
				} else {
					// Top layer is already empty, try removing it if more layers exist underneath
					if len(currentTile.Textures) > 1 {
						currentTile.Textures = currentTile.Textures[:len(currentTile.Textures)-1]
						changed = true
					}
				}
			}
		case "layers":
			newType := beam.FloorTile
			actionDesc = "Set Tile Type: Floor"
			if m.uiState.hasSwappedLayers {
				newType = beam.WallTile
				actionDesc = "Set Tile Type: Wall"
			}
			if currentTile.Type != newType {
				currentTile.Type = newType
				changed = true
			}
		case "location":
			actionDesc = "Set Location Marker"
			// Location setting logic remains the same, just ensure 'changed' is set if applicable
			// We don't store old location state in the tile itself for undo easily.
			// This might need a separate LocationChangeAction if undo is required.
			// For now, location setting is NOT undoable via the standard tile action.
			switch m.uiState.locationMode {
			case 0: // Player Start
				if m.tileGrid.Start != pos { // Only set if different
					m.tileGrid.Start = pos
					// No 'changed' flag here as it affects the map, not the tile data directly for this action type
					m.showToast(fmt.Sprintf("Player Start set to (%d, %d)", pos.X, pos.Y), ToastSuccess)
				}
			case 1: // Dungeon Entrance
				// Ensure not already added
				if !slices.Contains(m.tileGrid.DungeonEntry, pos) {
					// Mode 1 is additive, reset happens on tool right-click swap
					m.tileGrid.DungeonEntry = append(m.tileGrid.DungeonEntry, pos)
					m.showToast(fmt.Sprintf("Dungeon Entry added at (%d, %d)", pos.X, pos.Y), ToastSuccess)
				}
			case 2: // Respawn
				if m.tileGrid.Respawn != pos {
					m.tileGrid.Respawn = pos
					m.showToast(fmt.Sprintf("Respawn set to (%d, %d)", pos.X, pos.Y), ToastSuccess)
				}
			case 3: // Exit
				if m.tileGrid.Exit != pos {
					m.tileGrid.Exit = pos
					m.showToast(fmt.Sprintf("Exit set to (%d, %d)", pos.X, pos.Y), ToastSuccess)
				}
			}
		}
		newTiles[i] = currentTile // Store the potentially modified state
	}

	if changed {
		// Create and record the undoable action
		action := NewTileChangeAction(positions, oldTiles, newTiles, actionDesc)
		m.recordAction(action)

		// Apply the changes (already done implicitly by modifying copies, now make permanent)
		// No, the action's Do() method applies it. We just record here.
		// Let's call Do immediately after recording for non-undo/redo application
		action.Do(m) // Apply the recorded changes to the actual grid

		m.ValidateTileGrid() // Validate after changes are applied
	}
}

// showTileInfoPopup displays the tile info dialog
func (m *MapMaker) showTileInfoPopup(pos beam.Position, mousePos rl.Vector2) {
	if !m.showTileInfo { // Prevent overlapping popups if right-clicked fast
		m.uiState.tileInfoPos = pos
		m.uiState.tileInfoPopupX = int32(mousePos.X)
		m.uiState.tileInfoPopupY = int32(mousePos.Y)
		m.showTileInfo = true
		m.uiState.isDraggingPopup = false // Reset dragging state
	}
}

// swapEraserTool swaps between block and pencil eraser
func (m *MapMaker) swapEraserTool() {
	m.uiState.uiTextures["eraser"], m.uiState.uiTextures["pencileraser"] =
		m.uiState.uiTextures["pencileraser"], m.uiState.uiTextures["eraser"]
	m.uiState.hasSwappedEraser = !m.uiState.hasSwappedEraser
	currentToolName := "Eraser"
	if m.uiState.hasSwappedEraser {
		currentToolName = "Pencil Eraser"
	}
	// Update selected tool if one of the erasers was active
	if m.uiState.selectedTool == "eraser" || m.uiState.selectedTool == "pencileraser" {
		if m.uiState.hasSwappedEraser {
			m.uiState.selectedTool = "pencileraser"
		} else {
			m.uiState.selectedTool = "eraser"
		}
	}
	m.showToast(fmt.Sprintf("Swapped to %s", currentToolName), ToastInfo)
}

// swapLayersTool swaps between Floor and Wall type placement
func (m *MapMaker) swapLayersTool() {
	m.uiState.hasSwappedLayers = !m.uiState.hasSwappedLayers
	layerTypeName := "Floor"
	if m.uiState.hasSwappedLayers {
		m.uiState.uiTextures["layers"] = m.uiState.uiTextures["layerwall"]
		layerTypeName = "Wall"
	} else {
		m.uiState.uiTextures["layers"] = m.uiState.uiTextures["layerground"]
	}
	m.showToast(fmt.Sprintf("Layers tool set to: %s", layerTypeName), ToastInfo)
}

// swapLocationTool cycles through location marker types
func (m *MapMaker) swapLocationTool() {
	m.uiState.locationMode = (m.uiState.locationMode + 1) % 4 // Cycle through 4 states
	modeNames := []string{"Player Start", "Dungeon Entrance", "Respawn", "Exit"}

	// Reset DungeonEntry list when switching *to* mode 1 (Dungeon Entrance)
	if m.uiState.locationMode == 1 {
		m.tileGrid.DungeonEntry = beam.Positions{}
		m.showToast("Location Mode: Dungeon Entrance (Entries reset)", ToastInfo)
	} else {
		m.showToast(fmt.Sprintf("Location Mode: %s", modeNames[m.uiState.locationMode]), ToastInfo)
	}
}

// selectTool changes the active tool and provides feedback
func (m *MapMaker) selectTool(toolName string) {
	if m.uiState.selectedTool == toolName {
		m.uiState.selectedTool = "" // Deselect if clicked again
		m.showToast("Tool deselected", ToastInfo)
		m.clearSelection()
	} else {
		m.uiState.selectedTool = toolName
		feedback := toolName
		// Add specific feedback for swapped states
		if toolName == "eraser" && m.uiState.hasSwappedEraser {
			feedback = "Pencil Eraser"
		}
		if toolName == "layers" {
			if m.uiState.hasSwappedLayers {
				feedback = "Layers (Wall)"
			} else {
				feedback = "Layers (Floor)"
			}
		}
		if toolName == "location" {
			modeNames := []string{"Player Start", "Dungeon Entrance", "Respawn", "Exit"}
			feedback = fmt.Sprintf("Location (%s)", modeNames[m.uiState.locationMode])
		}

		m.showToast(fmt.Sprintf("%s tool selected", strings.Title(feedback)), ToastInfo)
		// Clear selection when changing tools, except maybe for select tool itself?
		if toolName != "select" {
			m.clearSelection()
		} else {
			// Ensure selection is cleared specifically when switching TO select tool
			m.clearSelection()
		}
	}
}

// changeLayer adjusts the currently active editing layer
func (m *MapMaker) changeLayer(delta int) {
	newLayer := m.uiState.currentLayer + delta
	if newLayer >= 0 && newLayer < m.uiState.maxLayers {
		m.uiState.currentLayer = newLayer
		m.showToast(fmt.Sprintf("Selected Layer %d", m.uiState.currentLayer+1), ToastInfo) // Show 1-based index
	} else if newLayer < 0 {
		m.showToast("Already on bottom layer", ToastWarning)
	} else {
		m.showToast(fmt.Sprintf("Maximum layers reached (%d)", m.uiState.maxLayers), ToastWarning)
	}
}

// handleZoom manages zooming the view
func (m *MapMaker) handleZoom(wheelMove float32, mousePos rl.Vector2) {
	// Get tile coordinates under mouse before zoom
	oldGridX, oldGridY, _ := m.getTileCoordsFromMouse(mousePos)

	// Calculate new zoom level
	oldZoom := m.uiState.zoomLevel
	m.uiState.zoomLevel += wheelMove * ZoomIncrement * m.uiState.zoomLevel // Zoom faster at higher levels
	m.uiState.zoomLevel = ClampF32(m.uiState.zoomLevel, MinZoom, MaxZoom)

	// If zoom actually changed
	if m.uiState.zoomLevel != oldZoom {
		// Get tile coordinates under mouse after zoom (if no panning adjustment)
		newGridX, newGridY, _ := m.getTileCoordsFromMouse(mousePos) // Needs recalculating based on new zoom

		// Calculate the shift in tile coordinates caused by zoom
		deltaX := oldGridX - newGridX
		deltaY := oldGridY - newGridY

		// Adjust viewport offset to keep the tile under the mouse stationary
		// This requires careful calculation involving the mouse position relative to the viewport center.
		// Simplified approach: Adjust viewport slightly towards mouse proportional to zoom change.
		// This might not be perfectly centered zoom.

		// More accurate centered zoom:
		// 1. World coords of mouse before zoom: mouseWorldBefore := m.screenToWorld(mousePos)
		// 2. Apply zoom
		// 3. World coords of mouse after zoom (if view didn't move): mouseWorldAfter := m.screenToWorld(mousePos)
		// 4. Required viewport shift: shift := mouseWorldBefore - mouseWorldAfter
		// 5. Apply shift to viewportOffset (in world/tile units): m.tileGrid.viewportOffset += shift

		// Let's try implementing screenToWorld/worldToScreen for accurate zoom/pan
		mouseWorldBefore := m.screenToWorld(mousePos)
		// Zoom applied here
		mouseWorldAfter := m.screenToWorld(mousePos) // Recalculate with new zoom

		deltaWorldX := mouseWorldBefore.X - mouseWorldAfter.X
		deltaWorldY := mouseWorldBefore.Y - mouseWorldAfter.Y

		// Adjust viewport offset (which is in tile coordinates)
		effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel // Use new zoom level
		m.tileGrid.viewportOffset.X += int(math.Round(float64(deltaWorldX / effectiveTileSize)))
		m.tileGrid.viewportOffset.Y += int(math.Round(float64(deltaWorldY / effectiveTileSize)))

		// Clamp viewport offset
		m.clampViewportOffset()

		// log.Printf("Zoom changed: %.2f", m.uiState.zoomLevel)
	}
}

// handlePanning manages dragging the view with middle mouse or keys
func (m *MapMaker) handlePanning(mousePos rl.Vector2) {
	// --- Middle Mouse Panning ---
	if rl.IsMouseButtonPressed(rl.MouseButtonMiddle) {
		if !m.isMouseOverUI(mousePos) { // Don't pan if starting over UI
			m.isPanning = true
			m.panStartMouse = mousePos
			m.panStartViewport = m.tileGrid.viewportOffset
			rl.HideCursor() // Optional: Hide cursor while panning
		}
	}
	if m.isPanning {
		if rl.IsMouseButtonDown(rl.MouseButtonMiddle) {
			deltaPixels := rl.Vector2Subtract(mousePos, m.panStartMouse)
			effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel
			if effectiveTileSize > 0 {
				// Calculate delta in tile coordinates
				deltaTilesX := int(math.Round(float64(-deltaPixels.X / effectiveTileSize))) // Invert delta for panning
				deltaTilesY := int(math.Round(float64(-deltaPixels.Y / effectiveTileSize))) // Invert delta

				m.tileGrid.viewportOffset.X = m.panStartViewport.X + deltaTilesX
				m.tileGrid.viewportOffset.Y = m.panStartViewport.Y + deltaTilesY

				m.clampViewportOffset()
			}
		} else {
			m.isPanning = false // Released button
			rl.ShowCursor()    // Restore cursor
		}
	}

	// --- Keyboard Panning ---
	panDelta := rl.Vector2Zero()
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		panDelta.Y -= 1
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		panDelta.Y += 1
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		panDelta.X -= 1
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		panDelta.X += 1
	}

	if panDelta.X != 0 || panDelta.Y != 0 {
		// Normalize and scale by speed and frame time
		panAmount := float32(PanSpeed * rl.GetFrameTime())
		panVector := rl.Vector2Scale(rl.Vector2Normalize(panDelta), panAmount)

		// Convert pixel pan amount to tile offset change
		effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel
		if effectiveTileSize > 0 {
			deltaTilesX := int(math.Round(float64(panVector.X / effectiveTileSize)))
			deltaTilesY := int(math.Round(float64(panVector.Y / effectiveTileSize)))

			if deltaTilesX != 0 || deltaTilesY != 0 {
				m.tileGrid.viewportOffset.X += deltaTilesX
				m.tileGrid.viewportOffset.Y += deltaTilesY
				m.clampViewportOffset()
			}
		}
	}
}

// clampViewportOffset ensures the viewport doesn't pan beyond reasonable map limits
func (m *MapMaker) clampViewportOffset() {
	// Define max scroll based on grid size (can be adjusted)
	// Allow scrolling slightly past the edge maybe? For now, clamp strictly.
	minX, minY := 0, 0
	maxX := m.tileGrid.Width // Max offset is showing only the last tile column/row?
	maxY := m.tileGrid.Height

	// Need to calculate max based on how many tiles fit on screen too
	effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel
	if effectiveTileSize <= 0 {
		return
	}
	workAreaWidth := float32(m.window.width)
	workAreaHeight := float32(m.window.height) - float32(m.uiState.menuBarHeight) - float32(m.uiState.statusBarHeight)
	tilesWide := int(math.Ceil(float64(workAreaWidth / effectiveTileSize)))
	tilesHigh := int(math.Ceil(float64(workAreaHeight / effectiveTileSize)))

	// Adjust max based on viewport size to prevent scrolling too far past the edge
	maxX = max(0, m.tileGrid.Width-tilesWide)
	maxY = max(0, m.tileGrid.Height-tilesHigh)

	m.tileGrid.viewportOffset.X = ClampInt(m.tileGrid.viewportOffset.X, minX, maxX)
	m.tileGrid.viewportOffset.Y = ClampInt(m.tileGrid.viewportOffset.Y, minY, maxY)
}

// screenToWorld converts screen pixel coordinates to world (grid pixel) coordinates
func (m *MapMaker) screenToWorld(screenPos rl.Vector2) rl.Vector2 {
	effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel
	if effectiveTileSize <= 0 {
		return rl.Vector2Zero()
	}
	// Inverse calculation of getTileScreenPosition basically
	worldX := (screenPos.X - 0) / effectiveTileSize // Relative to workspace top-left X=0
	worldY := (screenPos.Y - float32(m.uiState.menuBarHeight)) / effectiveTileSize

	// Add viewport offset (in tiles) converted back to pixels
	worldX += float32(m.tileGrid.viewportOffset.X) * effectiveTileSize
	worldY += float32(m.tileGrid.viewportOffset.Y) * effectiveTileSize

	// This seems wrong. World coords should be independent of viewport.
	// Let world coord (0,0) be the top-left of tile (0,0).
	// Screen(0, menuBarH) corresponds to world (viewX * effSize, viewY * effSize)

	// Calculate world position relative to the grid origin (0,0) in pixels
	// Start with mouse relative to workspace top-left
	relativeMouseX := screenPos.X - 0
	relativeMouseY := screenPos.Y - float32(m.uiState.menuBarHeight)

	// Add the world pixel position of the viewport's top-left corner
	viewportWorldX := float32(m.tileGrid.viewportOffset.X) * effectiveTileSize
	viewportWorldY := float32(m.tileGrid.viewportOffset.Y) * effectiveTileSize

	worldX = viewportWorldX + relativeMouseX
	worldY = viewportWorldY + relativeMouseY

	return rl.NewVector2(worldX, worldY)
}

// worldToScreen converts world (grid pixel) coordinates to screen coordinates
func (m *MapMaker) worldToScreen(worldPos rl.Vector2) rl.Vector2 {
	effectiveTileSize := float32(m.uiState.tileSize) * m.uiState.zoomLevel
	if effectiveTileSize <= 0 {
		return rl.Vector2Zero()
	}

	// Calculate screen position relative to the viewport's top-left position on screen
	viewportWorldX := float32(m.tileGrid.viewportOffset.X) * effectiveTileSize
	viewportWorldY := float32(m.tileGrid.viewportOffset.Y) * effectiveTileSize

	screenX := (worldPos.X - viewportWorldX) + 0 // Add workspace X origin
	screenY := (worldPos.Y - viewportWorldY) + float32(m.uiState.menuBarHeight) // Add workspace Y origin

	return rl.NewVector2(screenX, screenY)
}

// --- Utility Functions ---

func (m *MapMaker) clearSelection() {
	m.tileGrid.hasSelection = false
	m.tileGrid.selectedTiles = beam.Positions{} // Empty the slice
}

// isMouseOverUI checks if the mouse cursor is currently over any interactive UI element.
// Needs to be implemented by checking against all button bounds, dialog bounds etc.
func (m *MapMaker) isMouseOverUI(mousePos rl.Vector2) bool {
	// Check Menu Bar
	if mousePos.Y <= float32(m.uiState.menuBarHeight) {
		return true
	}
	// Check Status Bar
	if mousePos.Y >= float32(m.window.height-m.uiState.statusBarHeight) {
		return true
	}
	// Check Resource Viewer (if open and modal-like)
	if m.showResourceViewer {
		// Define bounds for resource viewer (needs access to its draw logic/rect)
		// Example: viewerWidth := int32(300)
		// viewerRect := rl.NewRectangle(float32(m.window.width-viewerWidth), float32(m.uiState.menuBarHeight), float32(viewerWidth), float32(m.window.height-m.uiState.menuBarHeight-m.uiState.statusBarHeight))
		// if rl.CheckCollisionPointRec(mousePos, viewerRect) { return true }
		// For now, assume it takes right side
		if mousePos.X > float32(m.window.width-300) { // Crude check
			return true
		}
	}
	// Check Tile Info Popup (if open)
	if m.showTileInfo {
		dialogWidth := 350 // Keep consistent with drawTileInfoPopup
		// Estimate height or get dynamically if possible
		dialogHeight := 200 // Estimate
		popupRect := rl.NewRectangle(float32(m.uiState.tileInfoPopupX), float32(m.uiState.tileInfoPopupY), float32(dialogWidth), float32(dialogHeight))
		if rl.CheckCollisionPointRec(mousePos, popupRect) {
			return true
		}
	}
	// Check Recent Textures Popup
	if m.showRecentTextures {
		// Define bounds for recent textures popup (needs access to its draw logic/rect)
		// Example: popupX := float32(m.window.width - 300 - 10) // Example position
		// popupY := float32(m.uiState.menuBarHeight + 10)
		// popupWidth := float32(200)
		// popupHeight := float32(8 * (m.uiState.tileSize + 5) + 10) // Example height
		// recentPopupRect := rl.NewRectangle(popupX, popupY, popupWidth, popupHeight)
		// if rl.CheckCollisionPointRec(mousePos, recentPopupRect) { return true }
		// Crude check for now
		if mousePos.X > float32(m.window.width-310) && mousePos.Y < 300 {
			return true
		}

	}
	// Add checks for other UI elements like texture editor if it's a popup

	return false
}

// Clamp utility functions
func ClampF32(value, minVal, maxVal float32) float32 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}
func ClampInt(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// --- Modified/Existing Functions (with Undo/Layer integration where applicable) ---

// handleSaveAction wraps save logic
func (m *MapMaker) handleSaveAction() {
	if m.currentFile != "" {
		if err := m.SaveMap(m.currentFile); err != nil {
			m.showToast("Error saving map: "+err.Error(), ToastError)
		} else {
			m.showToast("Map saved: "+filepath.Base(m.currentFile), ToastSuccess)
		}
	} else {
		m.handleSaveAsAction() // Trigger save as if no current file
	}
}

// handleSaveAsAction wraps save as logic
func (m *MapMaker) handleSaveAsAction() {
	filename := openSaveDialog() // Assuming this exists
	if filename != "" {
		if !strings.HasSuffix(filename, ".json") {
			filename += ".json"
		}
		if err := m.SaveMap(filename); err != nil {
			m.showToast("Error saving map: "+err.Error(), ToastError)
		} else {
			m.showToast("Map saved: "+filepath.Base(filename), ToastSuccess)
			m.currentFile = filename // Update current file on successful save as
			rl.SetWindowTitle(m.window.title + " - " + filepath.Base(m.currentFile))
		}
	}
}

// handleLoadAction wraps load logic
func (m *MapMaker) handleLoadAction() {
	// TODO: Add check for unsaved changes before loading
	filename := openLoadDialog() // Assuming this exists
	if filename != "" {
		if err := m.LoadMap(filename); err != nil {
			m.showToast("Error loading map: "+err.Error(), ToastError)
		} else {
			m.showToast("Map loaded: "+filepath.Base(filename), ToastSuccess)
			m.currentFile = filename // Update current file on successful load
			rl.SetWindowTitle(m.window.title + " - " + filepath.Base(m.currentFile))
			m.undoStack = make([]UndoableAction, 0, MaxUndoHistory) // Clear history on load
			m.redoStack = make([]UndoableAction, 0, MaxUndoHistory)
			m.resetViewport()    // Reset view on load
			m.clearSelection()   // Clear selection
			m.ValidateTileGrid() // Validate resources for the new map
		}
	}
}

// handleCloseAction wraps close logic
func (m *MapMaker) handleCloseAction() {
	// TODO: Add check for unsaved changes before closing
	if openCloseConfirmationDialog() { // Assuming this exists
		// Reset state
		m.uiState.tileSize = DefaultTileSize
		m.uiState.gridWidth = DefaultGridWidth   // These might be redundant if initTileGrid handles size
		m.uiState.gridHeight = DefaultGridHeight // These might be redundant if initTileGrid handles size
		m.uiState.zoomLevel = 1.0
		m.uiState.currentLayer = 0

		m.showResourceViewer = false
		m.showTileInfo = false
		m.showRecentTextures = false
		m.uiState.resourceViewerScroll = 0
		m.currentFile = ""
		rl.SetWindowTitle(m.window.title) // Reset window title

		// Reset grid and view
		m.initTileGrid(DefaultGridWidth, DefaultGridHeight)
		m.resetViewport()
		m.clearSelection()
		m.undoStack = make([]UndoableAction, 0, MaxUndoHistory) // Clear history
		m.redoStack = make([]UndoableAction, 0, MaxUndoHistory)
		m.ValidateTileGrid() // Validate empty grid (clears missing resources)

		m.showToast("Map closed", ToastInfo)
	}
}

// resetViewport centers the view on the grid
func (m *MapMaker) resetViewport() {
	m.tileGrid.viewportOffset = beam.Position{
		X: max(0, (m.tileGrid.Width/2)-30), // Try to center roughly
		Y: max(0, (m.tileGrid.Height/2)-20),
	}
	m.uiState.zoomLevel = 1.0
	m.clampViewportOffset() // Ensure it's valid
}

// initTileGrid initializes the tile grid with default values and specific size
func (m *MapMaker) initTileGrid(width, height int) {
	m.tileGrid.Width = width // Set dimensions on the embedded beam.Map
	m.tileGrid.Height = height
	m.tileGrid.Tiles = make([][]beam.Tile, height)
	for i := range m.tileGrid.Tiles {
		m.tileGrid.Tiles[i] = make([]beam.Tile, width)
		for j := range m.tileGrid.Tiles[i] {
			m.tileGrid.Tiles[i][j] = beam.Tile{
				Type:     beam.FloorTile,
				Pos:      beam.Position{X: j, Y: i},
				Textures: make([]beam.TileTexture, 0, m.uiState.maxLayers), // Pre-allocate slightly maybe?
			}
		}
	}
	// Reset other map properties if needed
	m.tileGrid.Start = beam.Position{-1, -1}
	m.tileGrid.Exit = beam.Position{-1, -1}
	m.tileGrid.Respawn = beam.Position{-1, -1}
	m.tileGrid.DungeonEntry = make(beam.Positions, 0)

	// Update UI state to reflect grid size (if UI shows these numbers)
	m.uiState.gridWidth = width
	m.uiState.gridHeight = height

	log.Printf("Initialized new grid: %d x %d", width, height)
}

// resizeGrid handles resizing the actual map data, preserving existing tiles
func (m *MapMaker) resizeGrid(newWidth, newHeight int) {
	if newWidth == m.tileGrid.Width && newHeight == m.tileGrid.Height {
		return // No change
	}
	log.Printf("Resizing grid from %dx%d to %dx%d", m.tileGrid.Width, m.tileGrid.Height, newWidth, newHeight)

	oldTiles := m.tileGrid.Tiles
	oldWidth := m.tileGrid.Width
	oldHeight := m.tileGrid.Height

	// Create new grid using init helper
	m.initTileGrid(newWidth, newHeight)

	// Copy existing tiles within bounds
	copyHeight := min(oldHeight, newHeight)
	copyWidth := min(oldWidth, newWidth)

	for y := 0; y < copyHeight; y++ {
		for x := 0; x < copyWidth; x++ {
			m.tileGrid.Tiles[y][x] = oldTiles[y][x]
			// Ensure position is updated (though initTileGrid should do this)
			m.tileGrid.Tiles[y][x].Pos = beam.Position{X: x, Y: y}
		}
	}

	// Update UI state controls if they drive the resize
	m.uiState.gridWidth = newWidth
	m.uiState.gridHeight = newHeight

	m.ValidateTileGrid() // Re-validate after resize
	m.clampViewportOffset() // Ensure viewport is still valid

	m.showToast(fmt.Sprintf("Grid resized to %d x %d", newWidth, newHeight), ToastInfo)
	// TODO: Implement resize as an UndoableAction if desired (complex)
}

// handleResizeGridUI handles button clicks for resizing (calls resizeGrid)
func (m *MapMaker) handleResizeGridUI(widthDelta, heightDelta int) {
	newWidth := m.uiState.gridWidth + widthDelta
	newHeight := m.uiState.gridHeight + heightDelta

	// Add min/max constraints
	newWidth = ClampInt(newWidth, 10, 500)   // Example limits
	newHeight = ClampInt(newHeight, 10, 500) // Example limits

	if newWidth != m.uiState.gridWidth || newHeight != m.uiState.gridHeight {
		m.resizeGrid(newWidth, newHeight)
	}
}

// handleTextureSelect handles the selection of a texture from the resource viewer
func (m *MapMaker) handleTextureSelect(texInfo *resources.TextureInfo) {
	if texInfo == nil {
		m.uiState.activeTexture = nil
		return
	}
	m.uiState.activeTexture = texInfo

	// Add to recent textures if not already present (and move to front)
	foundIndex := -1
	for i, name := range m.uiState.recentTextures {
		if name == texInfo.Name {
			foundIndex = i
			break
		}
	}
	if foundIndex != -1 {
		// Remove existing entry
		m.uiState.recentTextures = append(m.uiState.recentTextures[:foundIndex], m.uiState.recentTextures[foundIndex+1:]...)
	}
	// Add to front
	m.uiState.recentTextures = append([]string{texInfo.Name}, m.uiState.recentTextures...)

	// Keep only last 8 textures
	if len(m.uiState.recentTextures) > 8 {
		m.uiState.recentTextures = m.uiState.recentTextures[:8]
	}
}

// loadResource remains largely the same, but calls Validate
func (m *MapMaker) loadResource(name string, filepath string, isSheet bool, sheetMargin int32, gridSize int32) error {
	newRes := resources.Resource{
		Name:        name,
		Path:        filepath,
		IsSheet:     isSheet,
		SheetMargin: sheetMargin,
		GridSize:    gridSize,
	}

	if isSheet {
		// Assuming viewer.ViewSpritesheet remains synchronous for now
		finalGridSize, finalSheetMargin, err := viewer.ViewSpritesheet(newRes)
		if err != nil {
			log.Printf("Spritesheet viewer cancelled or failed: %v", err)
			return fmt.Errorf("spritesheet processing cancelled: %w", err)
		}
		newRes.GridSize = finalGridSize
		newRes.SheetMargin = finalSheetMargin
	}

	err := m.resources.AddResource("default", newRes) // Assuming "default" category
	if err != nil {
		log.Printf("Error adding resource '%s': %v", name, err)
		return err
	}
	m.ValidateTileGrid() // Re-check for missing resources now that new one is added
	m.showToast("Resource loaded: "+name, ToastSuccess)
	return nil
}

// ValidateTileGrid checks all tiles for missing texture resources
func (m *MapMaker) ValidateTileGrid() {
	m.tileGrid.missingResourceTiles = make(MissingResources, 0)
	if m.resources == nil {
		log.Println("Resource manager not initialized, cannot validate grid.")
		return
	}

	textureMap := m.resources.GetTextureMap("default") // Assuming "default" category

	for y := 0; y < m.tileGrid.Height; y++ {
		for x := 0; x < m.tileGrid.Width; x++ {
			tile := m.tileGrid.Tiles[y][x]
			for _, tex := range tile.Textures {
				if tex.Name != "" { // Only check non-empty texture names
					if _, exists := textureMap[tex.Name]; !exists {
						// Check if already added to prevent duplicates from multiple layers on same tile
						if !m.tileGrid.missingResourceTiles.Contains(tile.Pos, tex.Name) {
							m.tileGrid.missingResourceTiles = append(m.tileGrid.missingResourceTiles, MissingResource{
								tile:        tile.Pos,
								textureName: tex.Name,
							})
						}
					}
				}
			}
		}
	}
	if len(m.tileGrid.missingResourceTiles) > 0 {
		log.Printf("Validation complete: Found %d missing resource references.", len(m.tileGrid.missingResourceTiles))
		// Optionally show a persistent warning or count in the UI
	} else {
		log.Println("Validation complete: No missing resources found.")
	}
}

// Close cleans up resources
func (m *MapMaker) Close() {
	log.Println("Closing MapMaker...")
	m.saveConfig() // Save last file path etc.
	for _, tex := range m.uiState.uiTextures {
		rl.UnloadTexture(tex)
	}
	if m.resources != nil {
		m.resources.Close() // Close resource manager (unloads textures)
	}
	rl.CloseWindow() // Close raylib window
	log.Println("MapMaker Closed.")
}

// --- Config Loading/Saving ---
// (Keep simple file path saving as before, or expand to save UI state)
const configFile = ".mapmaker_config"

func (m *MapMaker) saveConfig() {
	// Save only the current file path for simplicity
	err := os.WriteFile(configFile, []byte(m.currentFile), 0644)
	if err != nil {
		log.Printf("Warning: Failed to save config: %v", err)
	}
}

func (m *MapMaker) loadConfig() {
	data, err := os.ReadFile(configFile)
	if err == nil {
		filepath := strings.TrimSpace(string(data))
		if filepath != "" {
			// Optionally try to load the map immediately
			if err := m.LoadMap(filepath); err == nil {
				m.currentFile = filepath
				rl.SetWindowTitle(m.window.title + " - " + filepath) // Use base name
				log.Println("Loaded last map:", filepath)
				m.showToast("Loaded last map: "+filepath, ToastInfo)
				m.ValidateTileGrid()
			} else {
				log.Printf("Could not automatically load last map '%s': %v", filepath, err)
				m.showToast("Could not load last map: "+filepath, ToastWarning)
				m.currentFile = "" // Reset if load failed
			}
		}
	} else if !os.IsNotExist(err) {
		log.Printf("Warning: Failed to load config: %v", err)
	}
}

// --- Helper Functions (Assume these exist or need implementation) ---

// openLoadDialog() string
// openSaveDialog() string
// openCloseConfirmationDialog() bool
// openLoadResourceDialog() (name string, filepath string, isSheet bool, sheetMargin int32, gridSize int32, err string)

// Button / IconButton types and New* / is*Clicked methods would remain largely the same.
// Toast type and showToast would remain largely the same.
// renderGrid, renderUI, renderToast need modifications for new features (zoom, pan, layers, undo/redo buttons, status bar, viewport culling, etc.) - These are complex and not fully shown here, but the logic structure provides the hooks.
// floodFillSelection logic remains the same for now.

// Remaining stubs/placeholders for rendering and dialogs would need implementation.
