package mapmaker

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
)

// Option Buttons
type Button struct {
	rect    rl.Rectangle
	text    string
	clicked bool
}

func (m *MapMaker) NewButton(x, y, width, height float32, text string) Button {
	return Button{
		rect: rl.Rectangle{X: x, Y: y, Width: width, Height: height},
		text: text,
	}
}

func (m *MapMaker) drawButton(btn Button, color rl.Color) {
	rl.DrawRectangleRec(btn.rect, color)
	rl.DrawRectangleLinesEx(btn.rect, 1, rl.DarkGray)

	fontSize := int32(12)
	textWidth := rl.MeasureText(btn.text, fontSize)
	textHeight := fontSize
	// Center text both horizontally and vertically
	textX := btn.rect.X + (btn.rect.Width-float32(textWidth))/2
	textY := btn.rect.Y + (btn.rect.Height-float32(textHeight))/2

	rl.DrawText(btn.text, int32(textX), int32(textY), fontSize, rl.DarkGray)
}

func (m *MapMaker) isButtonClicked(btn Button) bool {
	return rl.CheckCollisionPointRec(rl.GetMousePosition(), btn.rect) && rl.IsMouseButtonPressed(rl.MouseLeftButton)
}

// IconButton adds image icon support
type IconButton struct {
	rect    rl.Rectangle
	texture rl.Texture2D // Change icon from string to Texture2D
	srcRect rl.Rectangle // Source rectangle for the icon texture
	tooltip string
}

func (m *MapMaker) NewIconButton(x, y, width, height float32, texture rl.Texture2D, srcRect rl.Rectangle, tooltip string) IconButton {
	return IconButton{
		rect:    rl.Rectangle{X: x, Y: y, Width: width, Height: height},
		texture: texture,
		srcRect: srcRect,
		tooltip: tooltip,
	}
}

func (m *MapMaker) drawIconButton(btn IconButton, color rl.Color) {
	rl.DrawRectangleRec(btn.rect, color)
	rl.DrawRectangleLinesEx(btn.rect, 1, rl.DarkGray)

	// Draw the icon texture centered in the button
	iconSize := float32(24) // Default icon size
	iconX := btn.rect.X + (btn.rect.Width-iconSize)/2
	iconY := btn.rect.Y + (btn.rect.Height-iconSize)/2

	rl.DrawTexturePro(
		btn.texture,
		btn.srcRect,
		rl.Rectangle{
			X:      iconX,
			Y:      iconY,
			Width:  iconSize,
			Height: iconSize,
		},
		rl.Vector2{X: 0, Y: 0},
		0,
		rl.White,
	)

	// Draw tooltip on hover
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), btn.rect) {
		tooltipWidth := rl.MeasureText(btn.tooltip, 10)
		rl.DrawText(btn.tooltip,
			int32(btn.rect.X+(btn.rect.Width-float32(tooltipWidth))/2),
			int32(btn.rect.Y+btn.rect.Height+3),
			10,
			rl.DarkGray)
	}
}

func (m *MapMaker) isIconButtonClicked(btn IconButton) bool {
	return rl.CheckCollisionPointRec(rl.GetMousePosition(), btn.rect) && rl.IsMouseButtonPressed(rl.MouseLeftButton)
}

func (m *MapMaker) floodFillSelection(startX, startY int) beam.Positions {
	result := make(beam.Positions, 0)
	if startX < 0 || startX >= m.tileGrid.Width || startY < 0 || startY >= m.tileGrid.Height {
		return result
	}

	// Get the source tile's texture pattern
	sourceTile := m.tileGrid.Tiles[startY][startX]

	// Create a visited map
	visited := make(map[string]bool)

	// Stack for flood fill
	stack := beam.Positions{{X: startX, Y: startY}}

	// Check if two tiles have the same texture pattern
	matchesPattern := func(x, y int) bool {
		if x < 0 || x >= m.tileGrid.Width || y < 0 || y >= m.tileGrid.Height {
			return false
		}

		targetTile := m.tileGrid.Tiles[y][x]
		if len(targetTile.Textures) != len(sourceTile.Textures) {
			return false
		} else if targetTile.Type != sourceTile.Type {
			return false
		}

		// Compare each texture in the pattern
		for i, tex := range targetTile.Textures {
			sourceTex := sourceTile.Textures[i]

			// Check if both textures are complex or simple
			if tex.IsAnimated != sourceTex.IsAnimated {
				return false
			}

			// Compare frames if complex
			if tex.IsAnimated {
				if len(tex.Frames) != len(sourceTex.Frames) {
					return false
				}
				for j, frame := range tex.Frames {
					sourceFrame := sourceTex.Frames[j]
					if frame.Name != sourceFrame.Name ||
						frame.Rotation != sourceFrame.Rotation ||
						frame.Scale != sourceFrame.Scale ||
						frame.OffsetX != sourceFrame.OffsetX ||
						frame.OffsetY != sourceFrame.OffsetY ||
						frame.Tint != sourceFrame.Tint {
						return false
					}
				}
			} else {
				// Compare the first frame for simple textures
				if len(tex.Frames) == 0 || len(sourceTex.Frames) == 0 {
					return false
				}
				frame := tex.Frames[0]
				sourceFrame := sourceTex.Frames[0]
				if frame.Name != sourceFrame.Name ||
					frame.Rotation != sourceFrame.Rotation ||
					frame.Scale != sourceFrame.Scale ||
					frame.OffsetX != sourceFrame.OffsetX ||
					frame.OffsetY != sourceFrame.OffsetY ||
					frame.Tint != sourceFrame.Tint {
					return false
				}
			}
		}
		return true
	}

	// Process the stack
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		key := fmt.Sprintf("%d,%d", current.X, current.Y)
		if visited[key] {
			continue
		}

		visited[key] = true
		result = append(result, current)

		// Check all 4 directions
		directions := beam.Positions{
			{X: current.X + 1, Y: current.Y}, // right
			{X: current.X - 1, Y: current.Y}, // left
			{X: current.X, Y: current.Y + 1}, // down
			{X: current.X, Y: current.Y - 1}, // up
		}

		for _, dir := range directions {
			if dir.X >= 0 && dir.X < m.tileGrid.Width &&
				dir.Y >= 0 && dir.Y < m.tileGrid.Height {
				key := fmt.Sprintf("%d,%d", dir.X, dir.Y)
				if !visited[key] && matchesPattern(dir.X, dir.Y) {
					stack = append(stack, dir)
				}
			}
		}
	}

	return result
}

func openCloseConfirmationDialog() bool {
	dialogWidth := int32(300)
	dialogHeight := int32(150)

	for {
		if rl.WindowShouldClose() {
			return false
		}

		dialogX := (rl.GetScreenWidth() - int(dialogWidth)) / 2
		dialogY := (rl.GetScreenHeight() - int(dialogHeight)) / 2
		mousePos := rl.GetMousePosition()

		rl.BeginDrawing()
		// Draw background overlay
		rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.DarkGray, 0.3))

		// Draw dialog box
		rl.DrawRectangle(int32(dialogX), int32(dialogY), dialogWidth, dialogHeight, rl.RayWhite)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      float32(dialogX),
			Y:      float32(dialogY),
			Width:  float32(dialogWidth),
			Height: float32(dialogHeight),
		}, 2, rl.Gray)

		// Draw title and message
		rl.DrawText("Close Map", int32(dialogX+20), int32(dialogY+20), 20, rl.Black)
		rl.DrawText("Are you sure you want to close?", int32(dialogX+20), int32(dialogY+50), 16, rl.DarkGray)

		// Draw buttons
		cancelBtn := rl.Rectangle{
			X:      float32(dialogX + 20),
			Y:      float32(dialogY + int(dialogHeight) - 50),
			Width:  120,
			Height: 30,
		}
		confirmBtn := rl.Rectangle{
			X:      float32(dialogX + int(dialogWidth) - 140),
			Y:      float32(dialogY + int(dialogHeight) - 50),
			Width:  120,
			Height: 30,
		}

		rl.DrawRectangleRec(cancelBtn, rl.LightGray)
		rl.DrawRectangleRec(confirmBtn, rl.Red)

		// Center text in buttons
		cancelText := "Cancel"
		confirmText := "Close"
		cancelTextWidth := rl.MeasureText(cancelText, 16)
		confirmTextWidth := rl.MeasureText(confirmText, 16)

		rl.DrawText(cancelText,
			int32(cancelBtn.X+(cancelBtn.Width-float32(cancelTextWidth))/2),
			int32(cancelBtn.Y+(cancelBtn.Height-16)/2),
			16, rl.Black)
		rl.DrawText(confirmText,
			int32(confirmBtn.X+(confirmBtn.Width-float32(confirmTextWidth))/2),
			int32(confirmBtn.Y+(confirmBtn.Height-16)/2),
			16, rl.White)

		if rl.CheckCollisionPointRec(mousePos, cancelBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			rl.EndDrawing()
			return false
		}

		if rl.CheckCollisionPointRec(mousePos, confirmBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			rl.EndDrawing()
			return true
		}

		rl.EndDrawing()
	}
}

func openLoadResourceDialog() (string, string, bool, int32, int32, string) {
	dialog := ResourceDialog{
		visible:     true,
		sheetMargin: 0,
		gridSize:    16,
	}

	dialogWidth := int32(400)
	dialogHeight := int32(300)

	for dialog.visible {
		if rl.WindowShouldClose() {
			return "", "", false, 0, 0, "Cancelled"
		}

		dialogX := (rl.GetScreenWidth() - int(dialogWidth)) / 2
		dialogY := (rl.GetScreenHeight() - int(dialogHeight)) / 2
		mousePos := rl.GetMousePosition()

		rl.BeginDrawing()
		// Draw background overlay
		rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.DarkGray, 0.3))
		rl.DrawRectangle(int32(dialogX), int32(dialogY), int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      float32(dialogX),
			Y:      float32(dialogY),
			Width:  float32(dialogWidth),
			Height: float32(dialogHeight),
		}, 2, rl.Gray)

		// Title
		rl.DrawText("Add Resource", int32(dialogX+20), int32(dialogY+20), 20, rl.Black)

		// Name input field
		rl.DrawText("Name:", int32(dialogX+20), int32(dialogY+60), 16, rl.Black)
		rl.DrawRectangle(int32(dialogX+20), int32(dialogY+80), int32(dialogWidth-40), 30, rl.LightGray)
		rl.DrawText(dialog.name, int32(dialogX+25), int32(dialogY+87), 16, rl.Black)

		// File selection button
		fileButtonBounds := rl.Rectangle{
			X:      float32(dialogX + 20),
			Y:      float32(dialogY + 140),
			Width:  float32(dialogWidth - 40),
			Height: 30,
		}
		rl.DrawRectangleRec(fileButtonBounds, rl.LightGray)
		rl.DrawText("Select File", int32(fileButtonBounds.X+10), int32(fileButtonBounds.Y+8), 16, rl.Black)

		if rl.CheckCollisionPointRec(mousePos, fileButtonBounds) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			selectedPath := openFileDialog()
			if selectedPath != "" {
				dialog.path = selectedPath
			}
		}

		// Display selected file path
		if dialog.path != "" {
			pathDisplay := dialog.path
			if len(pathDisplay) > 40 {
				pathDisplay = "..." + pathDisplay[len(pathDisplay)-37:]
			}
			rl.DrawText(pathDisplay, int32(dialogX+25), int32(dialogY+180), 16, rl.Black)
		}

		// Sprite sheet checkbox
		checkboxBounds := rl.Rectangle{
			X:      float32(dialogX + 150),
			Y:      float32(dialogY + 200),
			Width:  20,
			Height: 20,
		}
		rl.DrawRectangleRec(checkboxBounds, rl.White)
		rl.DrawRectangleLinesEx(checkboxBounds, 1, rl.Gray)
		if dialog.isSheet {
			rl.DrawRectangle(int32(checkboxBounds.X+4), int32(checkboxBounds.Y+4), 12, 12, rl.Blue)
		}
		rl.DrawText("Is Sprite Sheet:", int32(dialogX+20), int32(dialogY+200), 16, rl.Black)

		if rl.CheckCollisionPointRec(mousePos, checkboxBounds) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			dialog.isSheet = !dialog.isSheet
		}

		// Grid size controls
		if dialog.isSheet {
			rl.DrawText(fmt.Sprintf("Grid Size: %d", dialog.gridSize), int32(dialogX+20), int32(dialogY+230), 16, rl.Black)

			minusButtonBounds := rl.Rectangle{
				X:      float32(dialogX + 120),
				Y:      float32(dialogY + 230),
				Width:  20,
				Height: 20,
			}
			plusButtonBounds := rl.Rectangle{
				X:      float32(dialogX + 150),
				Y:      float32(dialogY + 230),
				Width:  20,
				Height: 20,
			}

			rl.DrawRectangleRec(minusButtonBounds, rl.LightGray)
			rl.DrawRectangleRec(plusButtonBounds, rl.LightGray)
			rl.DrawText("-", int32(minusButtonBounds.X+7), int32(minusButtonBounds.Y+2), 16, rl.Black)
			rl.DrawText("+", int32(plusButtonBounds.X+6), int32(plusButtonBounds.Y+2), 16, rl.Black)

			if rl.CheckCollisionPointRec(mousePos, minusButtonBounds) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				if dialog.gridSize > 8 {
					dialog.gridSize--
				}
			}
			if rl.CheckCollisionPointRec(mousePos, plusButtonBounds) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				if dialog.gridSize < 128 {
					dialog.gridSize++
				}
			}
		}

		// Confirm/Cancel buttons
		cancelButtonBounds := rl.Rectangle{
			X:      float32(int32(dialogX) + dialogWidth - 180),
			Y:      float32(int32(dialogY) + dialogHeight - 40),
			Width:  70,
			Height: 30,
		}
		confirmButtonBounds := rl.Rectangle{
			X:      float32(int32(dialogX) + dialogWidth - 100),
			Y:      float32(int32(dialogY) + dialogHeight - 40),
			Width:  70,
			Height: 30,
		}

		rl.DrawRectangleRec(cancelButtonBounds, rl.LightGray)
		rl.DrawRectangleRec(confirmButtonBounds, rl.LightGray)
		rl.DrawText("Cancel", int32(cancelButtonBounds.X+10), int32(cancelButtonBounds.Y+8), 16, rl.Black)
		rl.DrawText("Confirm", int32(confirmButtonBounds.X+8), int32(confirmButtonBounds.Y+8), 16, rl.Black)

		if rl.CheckCollisionPointRec(mousePos, cancelButtonBounds) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			dialog.visible = false
			rl.EndDrawing()
			return "", "", false, 0, 0, "Cancelled"
		}

		if rl.CheckCollisionPointRec(mousePos, confirmButtonBounds) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			if dialog.name != "" && dialog.path != "" {
				dialog.visible = false
				rl.EndDrawing()
				return dialog.name, dialog.path, dialog.isSheet, dialog.sheetMargin, dialog.gridSize, ""
			}
		}

		// Handle text input for name
		key := rl.GetCharPressed()
		for key > 0 {
			if len(dialog.name) < 30 {
				dialog.name += string(key)
			}
			key = rl.GetCharPressed()
		}
		if rl.IsKeyPressed(rl.KeyBackspace) && len(dialog.name) > 0 {
			dialog.name = dialog.name[:len(dialog.name)-1]
		}

		rl.EndDrawing()
	}

	return "", "", false, 0, 0, "Cancelled"
}

func openFileDialog() string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e", `POSIX path of (choose file with prompt "Choose a sprite sheet or texture:" of type {"png","jpg","jpeg"})`)
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
