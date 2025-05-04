// A sprite sheet viewer application that allows users to load and
// view sprite sheets with configurable grid size and margin settings.
// It supports dynamic reloading and provides a graphical interface
// for viewing individual sprites within the sheet.

package viewer

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam/resources"
)

type Viewer struct {
	Cfg     Config
	UIState *UIState
}

// UIState holds the application state and configuration.
type UIState struct {
	ShowFileDialog bool
	Margin         int32
	GridSize       int32
	CurrentFile    string
	RM             *resources.ResourceManager
	Sheet          *resources.SpriteSheet
	SpriteNames    []string
	ScrollOffset   float32
	LoadError      string
	DebugInfo      string
}

type Config struct {
	DisplaySize    int32
	Padding        int32
	StartX         int32
	StartY         int32
	ViewportHeight int32
	HeaderHeight   int32
}

func NewViewer() *Viewer {
	return &Viewer{
		Cfg:     InitConfig(),
		UIState: InitUI(),
	}
}

func InitConfig() Config {
	return Config{
		DisplaySize:    32,
		Padding:        10,
		StartX:         50,
		StartY:         80,
		ViewportHeight: 500,
		HeaderHeight:   40,
	}
}

func InitUI() *UIState {
	ui := &UIState{
		Margin:   1,
		GridSize: 16,
	}
	return ui
}

// reload attempts to load or reload the current sprite sheet with the specified
// margin and grid size settings. It updates the internal state with any errors
// or debug information.
func (s *UIState) reload() {
	if s.CurrentFile == "" {
		return
	}

	newSprites := []resources.Resource{
		{
			Name:        "s",
			Path:        s.CurrentFile,
			IsSheet:     true,
			SheetMargin: int32(s.Margin),
			GridSize:    int32(s.GridSize),
		},
	}

	if s.RM != nil {
		s.RM.Close()
	}

	s.RM = resources.NewResourceManagerWithGlobal(newSprites, nil)
	if s.RM == nil {
		s.LoadError = "Failed to create resource manager"
		return
	}

	if len(s.RM.Scenes) == 0 || len(s.RM.Scenes[0].SpriteSheets) == 0 {
		s.LoadError = "No sprites found in sheet"
		return
	}

	s.Sheet = s.RM.Scenes[0].SpriteSheets[0]
	if s.Sheet.Texture.ID == 0 {
		s.LoadError = "Invalid texture"
		return
	}

	s.updateSpriteNames()
	s.DebugInfo = fmt.Sprintf("Loaded %d sprites", len(s.SpriteNames))
	s.LoadError = ""
}

// handleInput processes keyboard and mouse input events.
func (s *UIState) HandleInput(showSettings *bool) {
	if rl.IsKeyPressed(rl.KeyO) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) {
		if file := openFileDialog(); file != "" {
			s.CurrentFile = file
			s.reload()
		}
	}
	if rl.IsKeyPressed(rl.KeyEscape) && *showSettings {
		*showSettings = false
	}
	s.ScrollOffset -= rl.GetMouseWheelMove() * 30
}

// handleScrolling manages scroll state based on content height and viewport
func (s *UIState) HandleScrolling(contentHeight float32, viewportHeight int32) {
	maxScroll := float32(0)
	if contentHeight > float32(viewportHeight) {
		maxScroll = contentHeight - float32(viewportHeight)
	}
	if s.ScrollOffset < 0 {
		s.ScrollOffset = 0
	}
	if s.ScrollOffset > maxScroll {
		s.ScrollOffset = maxScroll
	}
}

// renderSprites draws all visible sprites from the sprite sheet.
func (s *UIState) RenderSprites(cfg Config) {
	if s.Sheet == nil || s.Sheet.Texture.ID == 0 {
		if s.LoadError == "" {
			rl.DrawText("No spritesheet loaded. Press 'Open File' to select one.", 50, cfg.StartY, 20, rl.Gray)
		}
		return
	}

	spritesPerRow := (800 - cfg.StartX*2) / (cfg.DisplaySize + cfg.Padding)
	totalRows := len(s.SpriteNames) / int(spritesPerRow)
	if len(s.SpriteNames)%int(spritesPerRow) != 0 {
		totalRows++
	}

	contentHeight := float32(cfg.StartY) + float32(totalRows*(int(cfg.DisplaySize)+int(cfg.Padding)+20))
	s.HandleScrolling(contentHeight, cfg.ViewportHeight)

	x, y := cfg.StartX, cfg.StartY
	for _, name := range s.SpriteNames {
		yPos := float32(y) - s.ScrollOffset

		if yPos+float32(cfg.DisplaySize) < 0 || yPos > float32(600) {
			x += cfg.DisplaySize + cfg.Padding
			if x > 700 {
				x = cfg.StartX
				y += cfg.DisplaySize + cfg.Padding + 20
			}
			continue
		}

		rect := s.Sheet.Sprites[name]

		source := rl.Rectangle{
			X:      float32(rect.X),
			Y:      float32(rect.Y),
			Width:  float32(rect.Width),
			Height: float32(rect.Height),
		}

		dest := rl.Rectangle{
			X:      float32(x),
			Y:      yPos,
			Width:  float32(cfg.DisplaySize),
			Height: float32(cfg.DisplaySize),
		}
		rl.DrawTexturePro(s.Sheet.Texture, source, dest, rl.Vector2{}, 0, rl.White)

		rl.DrawRectangleLinesEx(dest, 1, rl.Gray)
		rl.DrawText(name, int32(x), int32(int32(yPos)+cfg.DisplaySize+2), 10, rl.DarkGray)

		x += cfg.DisplaySize + cfg.Padding
		if x > 700 {
			x = cfg.StartX
			y += cfg.DisplaySize + cfg.Padding + 20
		}
	}

	if contentHeight > float32(cfg.ViewportHeight) {
		if s.ScrollOffset > 0 {
			rl.DrawTriangle(
				rl.Vector2{X: 780, Y: 50},
				rl.Vector2{X: 790, Y: 60},
				rl.Vector2{X: 770, Y: 60},
				rl.Gray)
		}
		if s.ScrollOffset < contentHeight-float32(cfg.ViewportHeight) {
			rl.DrawTriangle(
				rl.Vector2{X: 780, Y: float32(cfg.ViewportHeight + cfg.StartY - 10)},
				rl.Vector2{X: 770, Y: float32(cfg.ViewportHeight + cfg.StartY - 20)},
				rl.Vector2{X: 790, Y: float32(cfg.ViewportHeight + cfg.StartY - 20)},
				rl.Gray)
		}
	}
}

// renderUI draws the application interface including header, buttons, and settings panel.
func (s *UIState) RenderUI(cfg Config, showSettings *bool) {
	rl.DrawRectangle(0, 0, 800, cfg.HeaderHeight, rl.RayWhite)
	rl.DrawLine(0, cfg.HeaderHeight, 800, cfg.HeaderHeight, rl.LightGray)
	rl.DrawText("Sprite Sheet Viewer", 10, 10, 20, rl.Black)

	if drawButton(rl.Rectangle{X: 600, Y: 8, Width: 80, Height: 25}, "Settings") {
		*showSettings = !*showSettings
	}

	if drawButton(rl.Rectangle{X: 690, Y: 8, Width: 80, Height: 25}, "Open File") {
		if file := openFileDialog(); file != "" {
			s.CurrentFile = file
			s.reload()
		}
	}

	if s.DebugInfo != "" {
		rl.DrawText(s.DebugInfo, 450, 15, 10, rl.DarkGray)
	}

	if s.LoadError != "" {
		rl.DrawText(s.LoadError, 50, cfg.StartY, 20, rl.Red)
	}

	if *showSettings {
		panelHeight := int32(90)
		panelWidth := int32(300)

		settingsRect := rl.Rectangle{X: 400 - float32(panelWidth/2), Y: float32(cfg.HeaderHeight + 5)}

		rl.DrawRectangle(
			int32(settingsRect.X),
			int32(settingsRect.Y),
			panelWidth,
			panelHeight,
			rl.ColorAlpha(rl.LightGray, 0.95),
		)

		rl.DrawRectangleLinesEx(
			rl.Rectangle{
				X:      settingsRect.X,
				Y:      settingsRect.Y,
				Width:  float32(panelWidth),
				Height: float32(panelHeight),
			},
			1,
			rl.Black,
		)

		oldMargin := s.Margin
		oldGridSize := s.GridSize

		titleText := "Settings"
		titleWidth := rl.MeasureText(titleText, 15)
		rl.DrawText(titleText,
			int32(settingsRect.X+float32(panelWidth/2)-float32(titleWidth)/2),
			int32(settingsRect.Y+5),
			15,
			rl.Black)

		inputWidth := float32(60)
		inputHeight := float32(20)
		spacing := float32(40)

		totalWidth := inputWidth*2 + spacing
		startX := settingsRect.X + (float32(panelWidth)-totalWidth)/2

		marginInput := rl.Rectangle{
			X:      startX,
			Y:      settingsRect.Y + 45,
			Width:  inputWidth,
			Height: inputHeight,
		}

		gridInput := rl.Rectangle{
			X:      startX + inputWidth + spacing,
			Y:      settingsRect.Y + 45,
			Width:  inputWidth,
			Height: inputHeight,
		}

		s.Margin = drawInputField(marginInput, "Margin", s.Margin, 0, 24)
		s.GridSize = drawInputField(gridInput, "Grid Size", s.GridSize, 1, 128)

		helpText := "Use Up/Down keys when selected"
		helpWidth := rl.MeasureText(helpText, 10)
		helpX := settingsRect.X + float32(panelWidth/2) - float32(helpWidth)/2
		rl.DrawText(helpText, int32(helpX), int32(marginInput.Y+30), 10, rl.DarkGray)

		if oldMargin != s.Margin || oldGridSize != s.GridSize {
			s.reload()
		}
	}
}

// RenderViewerUI draws the application interface including header, buttons, and settings panel with confirm/cancel buttons.
func (s *UIState) RenderViewerUI(cfg Config, showSettings *bool) error {
	rl.DrawRectangle(0, 0, 800, cfg.HeaderHeight, rl.RayWhite)
	rl.DrawLine(0, cfg.HeaderHeight, 800, cfg.HeaderHeight, rl.LightGray)
	rl.DrawText("Sprite Sheet Viewer", 10, 10, 20, rl.Black)

	if drawButton(rl.Rectangle{X: 520, Y: 8, Width: 80, Height: 25}, "Settings") {
		*showSettings = !*showSettings
	}

	if drawButton(rl.Rectangle{X: 610, Y: 8, Width: 80, Height: 25}, "Confirm") {
		return fmt.Errorf("done")
	}

	if drawButton(rl.Rectangle{X: 700, Y: 8, Width: 80, Height: 25}, "Cancel") {
		return fmt.Errorf("cancelled")
	}

	if s.DebugInfo != "" {
		rl.DrawText(s.DebugInfo, 400, 15, 10, rl.DarkGray)
	}

	if s.LoadError != "" {
		rl.DrawText(s.LoadError, 50, cfg.StartY, 20, rl.Red)
	}

	if *showSettings {
		panelHeight := int32(90)
		panelWidth := int32(300)

		settingsRect := rl.Rectangle{X: 400 - float32(panelWidth/2), Y: float32(cfg.HeaderHeight + 5)}

		rl.DrawRectangle(
			int32(settingsRect.X),
			int32(settingsRect.Y),
			panelWidth,
			panelHeight,
			rl.ColorAlpha(rl.LightGray, 0.95),
		)

		rl.DrawRectangleLinesEx(
			rl.Rectangle{
				X:      settingsRect.X,
				Y:      settingsRect.Y,
				Width:  float32(panelWidth),
				Height: float32(panelHeight),
			},
			1,
			rl.Black,
		)

		oldMargin := s.Margin
		oldGridSize := s.GridSize

		titleText := "Settings"
		titleWidth := rl.MeasureText(titleText, 15)
		rl.DrawText(titleText,
			int32(settingsRect.X+float32(panelWidth/2)-float32(titleWidth)/2),
			int32(settingsRect.Y+5),
			15,
			rl.Black)

		inputWidth := float32(60)
		inputHeight := float32(20)
		spacing := float32(40)

		totalWidth := inputWidth*2 + spacing
		startX := settingsRect.X + (float32(panelWidth)-totalWidth)/2

		marginInput := rl.Rectangle{
			X:      startX,
			Y:      settingsRect.Y + 45,
			Width:  inputWidth,
			Height: inputHeight,
		}

		gridInput := rl.Rectangle{
			X:      startX + inputWidth + spacing,
			Y:      settingsRect.Y + 45,
			Width:  inputWidth,
			Height: inputHeight,
		}

		s.Margin = drawInputField(marginInput, "Margin", s.Margin, 0, 24)
		s.GridSize = drawInputField(gridInput, "Grid Size", s.GridSize, 1, 128)

		helpText := "Use Up/Down keys when selected"
		helpWidth := rl.MeasureText(helpText, 10)
		helpX := settingsRect.X + float32(panelWidth/2) - float32(helpWidth)/2
		rl.DrawText(helpText, int32(helpX), int32(marginInput.Y+30), 10, rl.DarkGray)

		if oldMargin != s.Margin || oldGridSize != s.GridSize {
			s.reload()
		}
	}
	return nil
}

// updateSpriteNames refreshes the sorted list of sprite names from the current sheet.
func (s *UIState) updateSpriteNames() {
	s.SpriteNames = nil
	for name := range s.Sheet.Sprites {
		s.SpriteNames = append(s.SpriteNames, name)
	}
	sort.Slice(s.SpriteNames, func(i, j int) bool {
		return naturalSort(s.SpriteNames[i], s.SpriteNames[j])
	})
}

func openFileDialog() string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e", `POSIX path of (choose file with prompt "Choose a sprite sheet:" of type {"png","jpg","jpeg"})`)
	case "linux":
		cmd = exec.Command("zenity", "--file-selection", "--file-filter=Images (*.png *.jpg *.jpeg)")
	default:
		return ""
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func drawButton(bounds rl.Rectangle, text string) bool {
	mousePoint := rl.GetMousePosition()
	btnState := rl.ColorAlpha(rl.Gray, 0.6)
	isHovered := rl.CheckCollisionPointRec(mousePoint, bounds)
	isClicked := isHovered && rl.IsMouseButtonPressed(rl.MouseLeftButton)

	if isHovered {
		btnState = rl.ColorAlpha(rl.DarkGray, 0.6)
	}

	rl.DrawRectangleRec(bounds, btnState)
	rl.DrawText(text, int32(bounds.X+bounds.Width/2-float32(rl.MeasureText(text, 10))/2),
		int32(bounds.Y+bounds.Height/2-5), 10, rl.Black)

	return isClicked
}

func drawInputField(bounds rl.Rectangle, label string, value int32, min, max int32) int32 {
	rl.DrawText(label, int32(bounds.X), int32(bounds.Y-15), 10, rl.Black)
	rl.DrawRectangleRec(bounds, rl.White)
	rl.DrawRectangleLinesEx(bounds, 1, rl.Gray)

	valueText := strconv.Itoa(int(value))
	textX := int32(bounds.X + 5)
	textY := int32(bounds.Y + bounds.Height/2 - 5)
	rl.DrawText(valueText, textX, textY, 10, rl.Black)

	mousePoint := rl.GetMousePosition()
	if rl.CheckCollisionPointRec(mousePoint, bounds) {
		if rl.IsKeyPressed(rl.KeyUp) {
			value = int32(rl.Clamp(float32(value+1), float32(min), float32(max)))
		} else if rl.IsKeyPressed(rl.KeyDown) {
			value = int32(rl.Clamp(float32(value-1), float32(min), float32(max)))
		}
	}

	return value
}

func naturalSort(a, b string) bool {
	aParts := strings.Split(a, "_")
	bParts := strings.Split(b, "_")

	minLen := len(aParts)
	if len(bParts) < minLen {
		minLen = len(bParts)
	}

	for i := 0; i < minLen; i++ {
		aNum, aErr := strconv.Atoi(aParts[i])
		bNum, bErr := strconv.Atoi(bParts[i])

		if aErr == nil && bErr == nil {
			if aNum != bNum {
				return aNum < bNum
			}
		} else {
			if aParts[i] != bParts[i] {
				return aParts[i] < bParts[i]
			}
		}
	}
	return len(aParts) < len(bParts)
}

// LoadSpritesheet loads a spritesheet with the specified configuration and returns the viewer instance
func ViewSpritesheet(res resources.Resource) (int32, int32, error) {
	viewer := NewViewer()
	viewer.UIState.CurrentFile = res.Path
	viewer.UIState.Margin = res.SheetMargin
	viewer.UIState.GridSize = res.GridSize
	viewer.UIState.reload()

	showSettings := false
	rl.SetExitKey(0)

	for !rl.WindowShouldClose() {
		viewer.UIState.HandleInput(&showSettings)
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		viewer.UIState.RenderSprites(viewer.Cfg)

		if err := viewer.UIState.RenderViewerUI(viewer.Cfg, &showSettings); err != nil {
			if err.Error() == "done" {
				break
			} else if err.Error() == "cancelled" {
				rl.EndDrawing()
				return 0, 0, err
			}
		}
		rl.EndDrawing()
	}

	return viewer.UIState.GridSize, viewer.UIState.Margin, nil
}
