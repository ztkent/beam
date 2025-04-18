package mapmaker

import (
	"fmt"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
)

func (m *MapMaker) renderGrid() {
	startX := m.tileGrid.offset.X
	startY := m.tileGrid.offset.Y

	// Calculate max visible tiles based on default size to maintain consistent viewport size
	maxVisibleWidth := MaxDisplayWidth * DefaultTileSize / m.uiState.tileSize
	maxVisibleHeight := MaxDisplayHeight * DefaultTileSize / m.uiState.tileSize

	// Calculate visible range based on viewport and adjusted max dimensions
	viewStartX := m.tileGrid.viewportOffset.X
	viewStartY := m.tileGrid.viewportOffset.Y
	viewEndX := min(viewStartX+maxVisibleWidth, m.tileGrid.Width)
	viewEndY := min(viewStartY+maxVisibleHeight, m.tileGrid.Height)

	// Draw grid lines for visible area
	visibleWidth := viewEndX - viewStartX
	visibleHeight := viewEndY - viewStartY

	// Draw horizontal grid lines
	for i := 0; i <= visibleWidth; i++ {
		x := startX + i*m.uiState.tileSize
		rl.DrawLine(int32(x), int32(startY), int32(x), int32(startY+visibleHeight*m.uiState.tileSize), rl.LightGray)
	}

	// Draw vertical grid lines
	for i := 0; i <= visibleHeight; i++ {
		y := startY + i*m.uiState.tileSize
		rl.DrawLine(int32(startX), int32(y), int32(startX+visibleWidth*m.uiState.tileSize), int32(y), rl.LightGray)
	}

	// Draw grid tiles within viewport
	for y := viewStartY; y < viewEndY; y++ {
		for x := viewStartX; x < viewEndX; x++ {
			// Calculate screen position for this tile
			screenX := startX + (x-viewStartX)*m.uiState.tileSize
			screenY := startY + (y-viewStartY)*m.uiState.tileSize

			pos := rl.Rectangle{
				X:      float32(screenX),
				Y:      float32(screenY),
				Width:  float32(m.uiState.tileSize),
				Height: float32(m.uiState.tileSize),
			}

			// Render tile at this location
			tile := m.tileGrid.Tiles[y][x]
			m.renderGridTile(pos, beam.Position{X: x, Y: y}, tile)
		}
	}

	// Draw viewport controls if any part of the grid is not visible
	if m.tileGrid.Width > maxVisibleWidth || m.tileGrid.Height > maxVisibleHeight {
		m.renderViewportControls()
	}

	// Draw selection highlight if there's a selection
	if m.tileGrid.hasSelection {
		for _, tile := range m.tileGrid.selectedTiles {
			// Only draw highlight if tile is in viewport
			if tile.X >= viewStartX && tile.X < viewEndX && tile.Y >= viewStartY && tile.Y < viewEndY {
				highlightX := startX + (tile.X-viewStartX)*m.uiState.tileSize
				highlightY := startY + (tile.Y-viewStartY)*m.uiState.tileSize

				// Highlight red if its an eraser
				color := rl.Black
				if m.uiState.selectedTool == "eraser" || m.uiState.selectedTool == "pencileraser" {
					color = rl.Red
				}

				rl.DrawRectangleLinesEx(rl.Rectangle{
					X:      float32(highlightX),
					Y:      float32(highlightY),
					Width:  float32(m.uiState.tileSize),
					Height: float32(m.uiState.tileSize),
				}, 2, color)
			}
		}
	}

	// Draw grid dimensions in bottom right
	dimensions := fmt.Sprintf("%dx%d", m.tileGrid.Width, m.tileGrid.Height)
	textWidth := int(rl.MeasureText(dimensions, 20))
	textX := startX + visibleWidth*m.uiState.tileSize - textWidth
	textY := startY + visibleHeight*m.uiState.tileSize + 5
	rl.DrawText(dimensions, int32(textX), int32(textY), 20, rl.DarkGray)
}

func (m *MapMaker) renderViewportControls() {
	btnSize := int32(24)
	gutterPadding := int32(15)
	btnSpacing := int32(2)
	verticalOffset := int(35)

	baseX := int32(gutterPadding)
	baseY := int32(m.tileGrid.offset.Y + (m.tileGrid.viewportHeight*m.uiState.tileSize)/2 + verticalOffset)

	maxVisibleWidth := MaxDisplayWidth * DefaultTileSize / m.uiState.tileSize
	maxVisibleHeight := MaxDisplayHeight * DefaultTileSize / m.uiState.tileSize

	remainingUp := m.tileGrid.viewportOffset.Y
	remainingDown := m.tileGrid.Height - (m.tileGrid.viewportOffset.Y + maxVisibleHeight)
	remainingLeft := m.tileGrid.viewportOffset.X
	remainingRight := m.tileGrid.Width - (m.tileGrid.viewportOffset.X + maxVisibleWidth)

	// Up button
	upBtn := rl.Rectangle{
		X:      float32(baseX + btnSize/2),
		Y:      float32(baseY - btnSize - btnSpacing),
		Width:  float32(btnSize),
		Height: float32(btnSize),
	}
	if remainingUp > 0 {
		rl.DrawTexturePro(
			m.uiState.uiTextures["up"],
			rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["up"].Width), Height: float32(m.uiState.uiTextures["up"].Height)},
			upBtn,
			rl.Vector2{X: 0, Y: 0},
			0,
			rl.White,
		)
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), upBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.tileGrid.viewportOffset.Y--
		}
	}

	// Down button
	downBtn := rl.Rectangle{
		X:      float32(baseX + btnSize/2),
		Y:      float32(baseY + btnSpacing),
		Width:  float32(btnSize),
		Height: float32(btnSize),
	}
	if remainingDown > 0 {
		rl.DrawTexturePro(
			m.uiState.uiTextures["down"],
			rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["down"].Width), Height: float32(m.uiState.uiTextures["down"].Height)},
			downBtn,
			rl.Vector2{X: 0, Y: 0},
			0,
			rl.White,
		)
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), downBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.tileGrid.viewportOffset.Y++
		}
	}

	// Left button
	leftBtn := rl.Rectangle{
		X:      float32(baseX),
		Y:      float32(baseY - btnSize/2),
		Width:  float32(btnSize),
		Height: float32(btnSize),
	}
	if remainingLeft > 0 {
		rl.DrawTexturePro(
			m.uiState.uiTextures["left"],
			rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["left"].Width), Height: float32(m.uiState.uiTextures["left"].Height)},
			leftBtn,
			rl.Vector2{X: 0, Y: 0},
			0,
			rl.White,
		)
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), leftBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.tileGrid.viewportOffset.X--
		}
	}

	// Right button
	rightBtn := rl.Rectangle{
		X:      float32(baseX + btnSize),
		Y:      float32(baseY - btnSize/2),
		Width:  float32(btnSize),
		Height: float32(btnSize),
	}
	if remainingRight > 0 {
		rl.DrawTexturePro(
			m.uiState.uiTextures["right"],
			rl.Rectangle{X: 0, Y: 0, Width: float32(m.uiState.uiTextures["right"].Width), Height: float32(m.uiState.uiTextures["right"].Height)},
			rightBtn,
			rl.Vector2{X: 0, Y: 0},
			0,
			rl.White,
		)
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), rightBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.tileGrid.viewportOffset.X++
		}
	}
}

func (m *MapMaker) renderGridTile(pos rl.Rectangle, pos2d beam.Position, tile beam.Tile) {
	if len(tile.Textures) == 0 {
		return
	}

	for _, tex := range tile.Textures {
		if len(tex.Frames) == 0 {
			continue
		}

		// If the texture isn't complex, we can just draw the frames on top of each other.
		if !tex.IsComplex {
			for _, frame := range tex.Frames {
				if frame.Name == "" {
					continue
				} else if m.tileGrid.missingResourceTiles.Contains(pos2d, frame.Name) {
					// Draw yellow outline for missing resource
					rl.DrawRectangleLinesEx(pos, 2, rl.Yellow)
					continue
				}

				// Center the texture in the tile
				origin := rl.Vector2{
					X: float32(m.uiState.tileSize) / 2,
					Y: float32(m.uiState.tileSize) / 2,
				}

				info, err := m.resources.GetTexture("default", frame.Name)
				if err != nil {
					fmt.Println("Error getting texture:", err)
					continue
				}

				// Adjust destination rectangle to use center-based rotation with scale and offset
				destRect := rl.Rectangle{
					X:      pos.X + pos.Width/2 + float32(frame.OffsetX*float64(m.uiState.tileSize)),
					Y:      pos.Y + pos.Height/2 + float32(frame.OffsetY*float64(m.uiState.tileSize)),
					Width:  pos.Width * float32(frame.Scale),
					Height: pos.Height * float32(frame.Scale),
				}

				// Add the tint
				if frame.Tint == (rl.Color{}) {
					frame.Tint = rl.White
				}

				rl.DrawTexturePro(
					info.Texture,
					info.Region,
					destRect,
					origin,
					float32(frame.Rotation),
					frame.Tint,
				)
			}
		} else {
			// If the texture is complex, we need draw the current frame for the animation time.
			frame := tex.GetCurrentFrame(rl.GetTime())
			origin := rl.Vector2{
				X: float32(m.uiState.tileSize) / 2,
				Y: float32(m.uiState.tileSize) / 2,
			}
			info, err := m.resources.GetTexture("default", frame.Name)
			if err != nil {
				fmt.Println("Error getting texture:", err)
				continue
			}
			destRect := rl.Rectangle{
				X:      pos.X + pos.Width/2 + float32(frame.OffsetX*float64(m.uiState.tileSize)),
				Y:      pos.Y + pos.Height/2 + float32(frame.OffsetY*float64(m.uiState.tileSize)),
				Width:  pos.Width * float32(frame.Scale),
				Height: pos.Height * float32(frame.Scale),
			}
			rl.DrawTexturePro(
				info.Texture,
				info.Region,
				destRect,
				origin,
				float32(frame.Rotation),
				frame.Tint,
			)
		}
	}

	if tile.Type == beam.WallTile {
		rl.DrawRectangleLinesEx(pos, 2, rl.Brown)
	}

	// Draw special tile outlines
	if pos2d.X != 0 && pos2d.Y != 0 {
		switch {
		case pos2d.X == m.tileGrid.Start.X && pos2d.Y == m.tileGrid.Start.Y:
			rl.DrawRectangleLinesEx(pos, 2, rl.Green)
		case pos2d.X == m.tileGrid.Exit.X && pos2d.Y == m.tileGrid.Exit.Y:
			rl.DrawRectangleLinesEx(pos, 2, rl.Red)
		case pos2d.X == m.tileGrid.Respawn.X && pos2d.Y == m.tileGrid.Respawn.Y:
			rl.DrawRectangleLinesEx(pos, 2, rl.Blue)
		}

		for _, entry := range m.tileGrid.DungeonEntry {
			if pos2d.X == entry.X && pos2d.Y == entry.Y {
				rl.DrawRectangleLinesEx(pos, 2, rl.Purple)
			}
		}
	}
}

func (m *MapMaker) renderUI() {
	// Draw header background
	rl.DrawRectangle(0, 0, m.window.width, int32(m.uiState.menuBarHeight), rl.RayWhite)
	rl.DrawLine(0, int32(m.uiState.menuBarHeight-1), m.window.width, int32(m.uiState.menuBarHeight-1), rl.LightGray)

	// Draw section dividers
	rl.DrawLine(150, 5, 150, int32(m.uiState.menuBarHeight-5), rl.LightGray)
	rl.DrawLine(m.window.width-180, 5, m.window.width-180, int32(m.uiState.menuBarHeight-5), rl.LightGray)

	// Get all buttons
	tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, resetBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn := m.getUIButtons()

	// Draw size control buttons
	m.drawButton(widthSmallerBtn, rl.White)
	m.drawButton(widthLargerBtn, rl.White)
	rl.DrawText(fmt.Sprintf("W:%d", m.uiState.gridWidth), 48, 12, 12, rl.DarkGray)

	m.drawButton(heightSmallerBtn, rl.White)
	m.drawButton(heightLargerBtn, rl.White)
	rl.DrawText(fmt.Sprintf("H:%d", m.uiState.gridHeight), 48, 37, 12, rl.DarkGray)

	m.drawButton(tileSmallerBtn, rl.White)
	m.drawButton(tileLargerBtn, rl.White)
	rl.DrawText(fmt.Sprintf("%dpx", m.uiState.tileSize), 48, 62, 12, rl.DarkGray)

	// Draw new grid control buttons
	m.drawToolIcons(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn)

	// Draw other icon buttons
	m.drawIconButton(saveBtn, rl.LightGray)
	m.drawIconButton(loadBtn, rl.LightGray)
	m.drawIconButton(loadResourceBtn, rl.LightGray)
	m.drawIconButton(viewResourcesBtn, rl.LightGray)
	m.drawIconButton(resetBtn, rl.LightGray)

	// Draw active texture preview box
	m.renderActiveTexturePreview()

	if m.showTileInfo {
		m.renderTileInfoPopup()
	}

	if m.showResourceViewer {
		m.renderResourceViewer()
	}

	// Draw status bar
	rl.DrawRectangle(0, m.window.height-int32(m.uiState.statusBarHeight),
		m.window.width, int32(m.uiState.statusBarHeight), rl.RayWhite)
	rl.DrawLine(0, m.window.height-int32(m.uiState.statusBarHeight),
		m.window.width, m.window.height-int32(m.uiState.statusBarHeight), rl.LightGray)

}

func (m *MapMaker) drawToolIcons(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn IconButton) {
	m.drawIconButton(paintbrushBtn, rl.LightGray)
	m.drawIconButton(paintbucketBtn, rl.LightGray)
	m.drawIconButton(eraseBtn, rl.LightGray)
	m.drawIconButton(selectBtn, rl.LightGray)
	m.drawIconButton(layersBtn, rl.LightGray)
	m.drawIconButton(locationBtn, rl.LightGray)

	// Draw tools with selection highlight
	toolButtons := map[string]IconButton{
		"paintbrush":   paintbrushBtn,
		"paintbucket":  paintbucketBtn,
		"eraser":       eraseBtn,
		"pencileraser": eraseBtn,
		"select":       selectBtn,
		"layers":       layersBtn,
		"location":     locationBtn,
	}
	for toolName, btn := range toolButtons {
		if m.uiState.selectedTool == toolName {
			// Draw selected border
			rl.DrawRectangleLinesEx(rl.Rectangle{
				X:      btn.rect.X - 1,
				Y:      btn.rect.Y - 1,
				Width:  btn.rect.Width + 2,
				Height: btn.rect.Height + 2,
			}, 3, rl.Blue)
		}
		m.drawIconButton(btn, rl.LightGray)
	}
}

func (m *MapMaker) renderActiveTexturePreview() {
	previewBox := rl.Rectangle{
		X:      float32(m.window.width - 335),
		Y:      15,
		Width:  30,
		Height: 30,
	}
	rl.DrawRectangleRec(previewBox, rl.LightGray)
	rl.DrawRectangleLinesEx(previewBox, 1, rl.Gray)

	if m.uiState.activeTexture != nil {
		if m.uiState.activeTexture.IsSheet {
			rl.DrawTexturePro(
				m.uiState.activeTexture.Texture,
				m.uiState.activeTexture.Region,
				rl.Rectangle{
					X:      previewBox.X,
					Y:      previewBox.Y,
					Width:  previewBox.Width,
					Height: previewBox.Height,
				},
				rl.Vector2{X: 0, Y: 0},
				0,
				rl.White,
			)
		} else {
			scale := float32(30) / float32(m.uiState.activeTexture.Texture.Width)
			if float32(m.uiState.activeTexture.Texture.Height)*scale > 30 {
				scale = float32(30) / float32(m.uiState.activeTexture.Texture.Height)
			}
			width := float32(m.uiState.activeTexture.Texture.Width) * scale
			height := float32(m.uiState.activeTexture.Texture.Height) * scale
			offsetX := (30 - width) / 2
			offsetY := (30 - height) / 2

			rl.DrawTexturePro(
				m.uiState.activeTexture.Texture,
				rl.Rectangle{
					X:      0,
					Y:      0,
					Width:  float32(m.uiState.activeTexture.Texture.Width),
					Height: float32(m.uiState.activeTexture.Texture.Height),
				},
				rl.Rectangle{
					X:      previewBox.X + offsetX,
					Y:      previewBox.Y + offsetY,
					Width:  width,
					Height: height,
				},
				rl.Vector2{X: 0, Y: 0},
				0,
				rl.White,
			)
		}
	}

	// Handle clicks on the preview box
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), previewBox) {
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.showRecentTextures = !m.showRecentTextures
		}

		tooltipText := "Click to view recent textures"
		if m.uiState.activeTexture != nil {
			tooltipText = m.uiState.activeTexture.Name
		}
		tooltipWidth := rl.MeasureText(tooltipText, 10)
		tooltipX := int32(previewBox.X + (previewBox.Width-float32(tooltipWidth))/2)
		tooltipY := int32(previewBox.Y + previewBox.Height + 3)
		rl.DrawText(tooltipText, tooltipX, tooltipY, 10, rl.DarkGray)
	}

	// Render recent textures popup if active
	if m.showRecentTextures && len(m.uiState.recentTextures) > 0 {
		popupWidth := int32(200)
		itemHeight := int32(40)
		padding := int32(5)
		popupHeight := int32(len(m.uiState.recentTextures))*itemHeight + padding*2

		// Position popup below the preview box
		popupX := int32(previewBox.X)
		popupY := int32(previewBox.Y + previewBox.Height + 20)

		// Draw popup background
		rl.DrawRectangle(popupX, popupY, popupWidth, popupHeight, rl.RayWhite)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      float32(popupX),
			Y:      float32(popupY),
			Width:  float32(popupWidth),
			Height: float32(popupHeight),
		}, 1, rl.Gray)

		// Draw recent textures
		i := 0
		for _, texName := range m.uiState.recentTextures {
			if _, err := m.resources.GetTexture("default", texName); err != nil {
				continue
			}
			itemY := popupY + padding + int32(i)*itemHeight
			itemRect := rl.Rectangle{
				X:      float32(popupX + padding),
				Y:      float32(itemY),
				Width:  float32(popupWidth - padding*2),
				Height: float32(itemHeight - padding),
			}

			// Draw highlight on hover
			mousePos := rl.GetMousePosition()
			if rl.CheckCollisionPointRec(mousePos, itemRect) {
				rl.DrawRectangleRec(itemRect, rl.LightGray)
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					if tex, err := m.resources.GetTexture("default", texName); err == nil {
						m.handleTextureSelect(&tex)
						m.showRecentTextures = false
					}
				}
			}

			// Draw texture preview
			if tex, err := m.resources.GetTexture("default", texName); err == nil {
				previewSize := float32(itemHeight - padding*2)
				rl.DrawTexturePro(
					tex.Texture,
					tex.Region,
					rl.Rectangle{
						X:      float32(popupX + padding),
						Y:      float32(itemY),
						Width:  previewSize,
						Height: previewSize,
					},
					rl.Vector2{X: 0, Y: 0},
					0,
					rl.White,
				)

				// Draw texture name
				rl.DrawText(texName,
					int32(popupX+padding*2+int32(previewSize)),
					int32(itemY+(itemHeight-padding)/2-5),
					10,
					rl.Black)
			}
			i++
		}

		// Close popup when clicking outside
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePos := rl.GetMousePosition()
			popupRect := rl.Rectangle{
				X:      float32(popupX),
				Y:      float32(popupY),
				Width:  float32(popupWidth),
				Height: float32(popupHeight),
			}
			if !rl.CheckCollisionPointRec(mousePos, popupRect) && !rl.CheckCollisionPointRec(mousePos, previewBox) {
				m.showRecentTextures = false
			}
		}
	}
}

func (m *MapMaker) renderResourceViewer() {
	dialogWidth := 600
	dialogHeight := 500
	dialogX := (rl.GetScreenWidth() - dialogWidth) / 2
	dialogY := (rl.GetScreenHeight() - dialogHeight) / 2

	// Draw main dialog box with layered borders for a cleaner look
	rl.DrawRectangle(int32(dialogX), int32(dialogY), int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(dialogX) - 1,
		Y:      float32(dialogY) - 1,
		Width:  float32(dialogWidth) + 2,
		Height: float32(dialogHeight) + 2,
	}, 2, rl.Fade(rl.Black, 0.2))
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(dialogX),
		Y:      float32(dialogY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}, 1, rl.Gray)

	// Title section and heading buttons
	titleHeight := 50
	rl.DrawText("Loaded Resources", int32(dialogX+20), int32(dialogY+20), 20, rl.Black)

	// Add manage button
	manageBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - 115),
		Y:      float32(dialogY + 10),
		Width:  70,
		Height: 30,
	}
	rl.DrawRectangleRec(manageBtn, rl.LightGray)
	rl.DrawText("Manage", int32(manageBtn.X+8), int32(manageBtn.Y+8), 16, rl.Black)

	// Close button
	closeBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - 40),
		Y:      float32(dialogY + 10),
		Width:  30,
		Height: 30,
	}
	rl.DrawRectangleRec(closeBtn, rl.LightGray)
	rl.DrawText("X", int32(closeBtn.X+10), int32(closeBtn.Y+5), 20, rl.Black)

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), closeBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.showResourceViewer = false
		if m.uiState.textureEditor != nil && m.uiState.textureEditor.advSelectingFrameIndex != -1 {
			m.uiState.textureEditor.advSelectingFrameIndex = -1
		}
	}

	// Toggle manage mode when manage button is clicked
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), manageBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.uiState.resourceManageMode = !m.uiState.resourceManageMode
	}

	// Setup scrollable content area
	contentArea := rl.Rectangle{
		X:      float32(dialogX),
		Y:      float32(dialogY + titleHeight),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight - titleHeight),
	}

	// Handle mouse wheel for scrolling
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), contentArea) {
		wheel := rl.GetMouseWheelMove()
		m.uiState.resourceViewerScroll -= int(wheel * 30)
	}

	// Grid layout settings with adjusted padding
	const (
		previewSize    = 32
		padding        = 12
		leftMargin     = 25
		bottomMargin   = 20
		itemsPerRow    = 10
		itemTotalWidth = previewSize + padding*2
	)

	// Calculate content bounds
	ss, _ := m.resources.GetAllSpritesheets("default")
	textures, _ := m.resources.GetAllTextures("default", false)

	totalRows := (len(textures) + len(ss) + itemsPerRow - 1) / itemsPerRow
	contentHeight := totalRows*int(itemTotalWidth) + int(bottomMargin)

	// Clamp scroll
	maxScroll := int(0)
	if contentHeight > int(contentArea.Height) {
		maxScroll = contentHeight - int(contentArea.Height)
	}
	if m.uiState.resourceViewerScroll > maxScroll {
		m.uiState.resourceViewerScroll = maxScroll
	}
	if m.uiState.resourceViewerScroll < 0 {
		m.uiState.resourceViewerScroll = 0
	}

	// Begin scissors mode for content clipping
	rl.BeginScissorMode(
		int32(contentArea.X),
		int32(contentArea.Y),
		int32(contentArea.Width),
		int32(contentArea.Height-20),
	)

	if m.uiState.resourceManageMode {
		// Draw manage mode view
		itemHeight := int32(40)
		padding := int32(10)
		for i, texInfo := range ss {
			y := int32(dialogY+titleHeight+i*int(itemHeight)) - int32(m.uiState.resourceViewerScroll)
			itemRect := rl.Rectangle{
				X:      float32(int32(dialogX) + padding),
				Y:      float32(y),
				Width:  float32(int32(dialogWidth) - padding*3),
				Height: float32(itemHeight - padding),
			}
			// Skip if item is outside visible area
			if y+itemHeight < int32(dialogY+titleHeight) || y > int32(dialogY+dialogHeight) {
				continue
			}
			// Draw item background
			rl.DrawRectangleRec(itemRect, rl.LightGray)
			// Draw texture name
			rl.DrawText(texInfo.Name, int32(itemRect.X+10), int32(itemRect.Y+8), 16, rl.Black)
			// Draw delete button
			deleteBtn := rl.Rectangle{
				X:      itemRect.X + itemRect.Width - 60,
				Y:      itemRect.Y + 2,
				Width:  50,
				Height: 26,
			}
			rl.DrawRectangleRec(deleteBtn, rl.Red)
			rl.DrawText("Delete", int32(deleteBtn.X+5), int32(deleteBtn.Y+6), 14, rl.White)

			// Handle delete button click
			if rl.CheckCollisionPointRec(rl.GetMousePosition(), deleteBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				err := m.resources.RemoveResource("default", texInfo.Name)
				if err != nil {
					fmt.Println("Error removing resource:", err)
				}
				m.ValidateTileGrid()
			}
		}
	} else {
		// Draw normal grid view
		startY := dialogY + titleHeight - m.uiState.resourceViewerScroll
		startX := dialogX + leftMargin

		for i, texInfo := range textures {
			col := i % itemsPerRow
			row := i / itemsPerRow
			x := startX + col*int(itemTotalWidth)
			y := startY + row*itemTotalWidth

			// Draw texture preview background
			rl.DrawRectangle(int32(x), int32(y), previewSize, previewSize, rl.LightGray)

			// Draw the texture
			if texInfo.IsSheet {
				rl.DrawTexturePro(
					texInfo.Texture,
					texInfo.Region,
					rl.Rectangle{
						X:      float32(x),
						Y:      float32(y),
						Width:  float32(previewSize),
						Height: float32(previewSize),
					},
					rl.Vector2{X: 0, Y: 0},
					0,
					rl.White,
				)
			} else {
				scale := float32(previewSize) / float32(texInfo.Texture.Width)
				if float32(texInfo.Texture.Height)*scale > float32(previewSize) {
					scale = float32(previewSize) / float32(texInfo.Texture.Height)
				}

				width := float32(texInfo.Texture.Width) * scale
				height := float32(texInfo.Texture.Height) * scale
				offsetX := (float32(previewSize) - width) / 2
				offsetY := (float32(previewSize) - height) / 2

				rl.DrawTexturePro(
					texInfo.Texture,
					rl.Rectangle{
						X:      0,
						Y:      0,
						Width:  float32(texInfo.Texture.Width),
						Height: float32(texInfo.Texture.Height),
					},
					rl.Rectangle{
						X:      float32(x) + offsetX,
						Y:      float32(y) + offsetY,
						Width:  width,
						Height: height,
					},
					rl.Vector2{X: 0, Y: 0},
					0,
					rl.White,
				)
			}

			// After drawing the texture preview, add click handling
			clickArea := rl.Rectangle{
				X:      float32(x),
				Y:      float32(y),
				Width:  previewSize,
				Height: previewSize,
			}

			// Highlight active texture
			if m.uiState.activeTexture != nil && m.uiState.activeTexture.Name == texInfo.Name &&
				(m.uiState.textureEditor == nil || m.uiState.textureEditor.advSelectingFrameIndex == -1) {
				rl.DrawRectangleLinesEx(clickArea, 2, rl.Blue)
			}

			if rl.CheckCollisionPointRec(rl.GetMousePosition(), clickArea) &&
				rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				tex, err := m.resources.GetTexture("default", texInfo.Name)
				if err != nil {
					fmt.Println("Error getting texture:", err)
				} else {
					m.handleTextureSelect(&tex)
				}
			}
		}
	}

	rl.EndScissorMode()
}

func (m *MapMaker) renderTileInfoPopup() {
	pos := m.uiState.tileInfoPos
	dialogWidth := 350
	dialogHeight := 300

	// Handle dragging
	mousePos := rl.GetMousePosition()
	dragArea := rl.Rectangle{
		X:      float32(m.uiState.tileInfoPopupX),
		Y:      float32(m.uiState.tileInfoPopupY),
		Width:  float32(dialogWidth),
		Height: 30,
	}

	if rl.CheckCollisionPointRec(mousePos, dragArea) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.uiState.isDraggingPopup = true
	}

	if m.uiState.isDraggingPopup && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		m.uiState.tileInfoPopupX = int32(mousePos.X - dragArea.Width/2)
		m.uiState.tileInfoPopupY = int32(mousePos.Y - dragArea.Height/2)
	}

	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		m.uiState.isDraggingPopup = false
	}

	// Ensure popup stays within window bounds
	if m.uiState.tileInfoPopupX+int32(dialogWidth) > m.window.width {
		m.uiState.tileInfoPopupX = m.window.width - int32(dialogWidth)
	}
	if m.uiState.tileInfoPopupX < 0 {
		m.uiState.tileInfoPopupX = 0
	}
	if m.uiState.tileInfoPopupY+int32(dialogHeight) > m.window.height {
		m.uiState.tileInfoPopupY = m.window.height - int32(dialogHeight)
	}
	if m.uiState.tileInfoPopupY < 0 {
		m.uiState.tileInfoPopupY = 0
	}

	// Draw popup background
	rl.DrawRectangle(m.uiState.tileInfoPopupX, m.uiState.tileInfoPopupY, int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(m.uiState.tileInfoPopupX),
		Y:      float32(m.uiState.tileInfoPopupY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}, 1, rl.Gray)

	// Draw content
	padding := int32(10)
	textY := m.uiState.tileInfoPopupY + padding

	// Draw tile type
	tile := m.tileGrid.Tiles[pos.Y][pos.X]
	rl.DrawText(fmt.Sprintf("Tile Type: %d", tile.Type), m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 25

	// Draw tile position
	rl.DrawText(fmt.Sprintf("Position: (%d, %d)", tile.Pos.X, tile.Pos.Y), m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 25

	// Draw textures
	rl.DrawText("Textures:", m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 20

	for texIndex, tex := range tile.Textures {
		// Draw complex text and edit button side by side
		complexText := fmt.Sprintf("- Complex: %t", tex.IsComplex)
		rl.DrawText(complexText, m.uiState.tileInfoPopupX+padding+10, textY, 14, rl.DarkGray)

		// Create edit button beside the complex text
		editBtn := rl.Rectangle{
			X:      float32(m.uiState.tileInfoPopupX + padding + 10 + rl.MeasureText(complexText, 14) + 10),
			Y:      float32(textY),
			Width:  30,
			Height: 15,
		}
		rl.DrawRectangleRec(editBtn, rl.LightGray)
		rl.DrawText("Edit", int32(editBtn.X+5), int32(editBtn.Y+2), 10, rl.Black)

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), editBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.closeAllEditors()
			m.showTileInfo = true // Keep tile info open

			// Initialize base editor state
			editor := &TextureEditorState{
				visible:       true,
				tile:          &m.tileGrid.Tiles[pos.Y][pos.X],
				texIndex:      texIndex,
				frameIndex:    0,
				clearedInputs: make(map[string]bool),
			}

			// Set up editor fields based on whether texture is complex
			if tex.IsComplex && len(tex.Frames) > 0 {
				// Go straight to advanced editor for complex textures
				editor.advAnimationTimeStr = fmt.Sprintf("%.2f", tex.AnimationTime)
				editor.advFrameCountStr = fmt.Sprintf("%d", len(tex.Frames))
				editor.advSelectedFrames = make([]string, len(tex.Frames))
				for i, frame := range tex.Frames {
					editor.advSelectedFrames[i] = frame.Name
				}
				editor.advSelectingFrameIndex = -1
				m.uiState.textureEditor = editor
				m.uiState.showAdvancedEditor = true
				m.uiState.activeInput = ""
			} else {
				// Use simple editor for non-complex textures
				if len(tex.Frames) > 0 {
					firstFrame := tex.Frames[0]
					editor.rotation = fmt.Sprintf("%.1f", firstFrame.Rotation)
					editor.scale = fmt.Sprintf("%.2f", firstFrame.Scale)
					editor.offsetX = fmt.Sprintf("%.2f", firstFrame.OffsetX)
					editor.offsetY = fmt.Sprintf("%.2f", firstFrame.OffsetY)
					editor.tintR = fmt.Sprintf("%d", firstFrame.Tint.R)
					editor.tintG = fmt.Sprintf("%d", firstFrame.Tint.G)
					editor.tintB = fmt.Sprintf("%d", firstFrame.Tint.B)
					editor.tintA = fmt.Sprintf("%d", firstFrame.Tint.A)
				} else {
					editor.rotation = "0.0"
					editor.scale = "1.0"
					editor.offsetX = "0.0"
					editor.offsetY = "0.0"
					editor.tintR = "255"
					editor.tintG = "255"
					editor.tintB = "255"
					editor.tintA = "255"
				}
				m.uiState.textureEditor = editor
			}
		}
		textY += 20

		for _, frame := range tex.Frames {
			warningText := ""
			textColor := rl.DarkGray

			if m.tileGrid.missingResourceTiles.Contains(pos, frame.Name) {
				warningText = " !"
				textColor = rl.Yellow
			}

			rl.DrawText(fmt.Sprintf("  - %s (%.1f°) Scale: %.2f Offset: (%.2f, %.2f)",
				frame.Name, frame.Rotation, frame.Scale, frame.OffsetX, frame.OffsetY),
				m.uiState.tileInfoPopupX+padding+5, textY, 12, textColor)
			textY += 15

			// Add tint information on next line, indented further
			rl.DrawText(fmt.Sprintf("    Tint: R:%d G:%d B:%d A:%d%s",
				frame.Tint.R, frame.Tint.G, frame.Tint.B, frame.Tint.A, warningText),
				m.uiState.tileInfoPopupX+padding+5, textY, 12, textColor)
			textY += 25
		}
	}
	// Draw close button with corrected position
	closeBtn := rl.Rectangle{
		X:      float32(m.uiState.tileInfoPopupX + int32(dialogWidth) - 30),
		Y:      float32(m.uiState.tileInfoPopupY + 5),
		Width:  25,
		Height: 25,
	}

	// Draw the X button
	rl.DrawRectangleRec(closeBtn, rl.LightGray)
	rl.DrawText("×", int32(closeBtn.X+7), int32(closeBtn.Y+2), 20, rl.DarkGray)

	// Check for clicks on close button when not dragging
	if !m.uiState.isDraggingPopup && rl.CheckCollisionPointRec(rl.GetMousePosition(), closeBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.showTileInfo = false
	}

	// Render texture editor if active
	if m.uiState.textureEditor != nil && m.uiState.textureEditor.visible {
		m.renderTextureEditor()
	}
}

type TextureEditorState struct {
	tile          *beam.Tile
	visible       bool
	texIndex      int
	frameIndex    int
	rotation      string
	scale         string
	offsetX       string
	offsetY       string
	tintR         string
	tintG         string
	tintB         string
	tintA         string
	clearedInputs map[string]bool

	// Advanced Editor State
	advAnimationTimeStr    string
	advFrameCountStr       string
	advSelectedFrames      []string // Stores texture names for each frame
	advSelectingFrameIndex int      // Index of the frame being selected via resource viewer, -1 if none
}

func (m *MapMaker) renderTextureEditor() {
	editor := m.uiState.textureEditor
	if editor == nil {
		return
	}
	if editor.clearedInputs == nil {
		editor.clearedInputs = make(map[string]bool)
	}

	dialogWidth := 300
	dialogHeight := 350
	dialogX := (rl.GetScreenWidth() - dialogWidth) / 2
	dialogY := (rl.GetScreenHeight() - dialogHeight) / 2

	// Draw dialog background
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.Black, 0.7))
	rl.DrawRectangle(int32(dialogX), int32(dialogY), int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(dialogX),
		Y:      float32(dialogY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}, 1, rl.Gray)

	// Title
	rl.DrawText("Edit Texture Properties", int32(dialogX+10), int32(dialogY+10), 20, rl.Black)

	// Input fields
	y := dialogY + 50
	padding := 20
	labelWidth := 80
	inputWidth := 100
	inputHeight := 30

	// Helper function to create input field
	createInput := func(label string, value *string, yPos int) {
		rl.DrawText(label, int32(dialogX+padding), int32(yPos+8), 16, rl.Black)
		inputRect := rl.Rectangle{
			X:      float32(dialogX + padding + labelWidth),
			Y:      float32(yPos),
			Width:  float32(inputWidth),
			Height: float32(inputHeight),
		}
		rl.DrawRectangleRec(inputRect, rl.LightGray)
		rl.DrawText(*value, int32(inputRect.X+5), int32(inputRect.Y+8), 16, rl.Black)

		// Handle input focus and text input
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), inputRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.uiState.activeInput = label
		}
		if m.uiState.activeInput == label {
			rl.DrawRectangleLinesEx(inputRect, 2, rl.Blue)

			// Clear value on first keypress if not already cleared
			key := rl.GetCharPressed()
			if !editor.clearedInputs[label] && key > 0 {
				*value = ""
				editor.clearedInputs[label] = true
			}

			for key > 0 {
				if key >= 32 && key <= 126 {
					*value += string(key)
				}
				key = rl.GetCharPressed()
			}
			if rl.IsKeyPressed(rl.KeyBackspace) && len(*value) > 0 {
				*value = (*value)[:len(*value)-1]
			}
		}
	}

	// Create all input fields
	createInput("Rotation", &editor.rotation, y)
	y += inputHeight + padding
	createInput("Scale", &editor.scale, y)
	y += inputHeight + padding
	createInput("Offset X", &editor.offsetX, y)
	y += inputHeight + padding
	createInput("Offset Y", &editor.offsetY, y)
	y += inputHeight + padding

	// Consolidated tint inputs
	tintLabel := "Tint RGBA:"
	rl.DrawText(tintLabel, int32(dialogX+padding)-10, int32(y+8), 16, rl.Black)

	tintWidth := 45
	tintSpacing := 5
	tintX := dialogX + padding + labelWidth

	// Draw tint input boxes in a row
	drawTintInput := func(value *string, x float32) rl.Rectangle {
		rect := rl.Rectangle{
			X:      x,
			Y:      float32(y),
			Width:  float32(tintWidth),
			Height: float32(inputHeight),
		}
		rl.DrawRectangleRec(rect, rl.LightGray)
		rl.DrawText(*value, int32(rect.X+5), int32(rect.Y+8), 16, rl.Black)
		return rect
	}

	rRect := drawTintInput(&editor.tintR, float32(tintX))
	gRect := drawTintInput(&editor.tintG, float32(tintX+tintWidth+tintSpacing))
	bRect := drawTintInput(&editor.tintB, float32(tintX+2*(tintWidth+tintSpacing)))
	aRect := drawTintInput(&editor.tintA, float32(tintX+3*(tintWidth+tintSpacing)))

	// Handle input focus for tint fields
	for idx, rect := range []rl.Rectangle{rRect, gRect, bRect, aRect} {
		label := []string{"TintR", "TintG", "TintB", "TintA"}[idx]
		value := []*string{&editor.tintR, &editor.tintG, &editor.tintB, &editor.tintA}[idx]

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), rect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.uiState.activeInput = label
		}
		if m.uiState.activeInput == label {
			rl.DrawRectangleLinesEx(rect, 2, rl.Blue)

			// Clear value on first keypress if not already cleared
			key := rl.GetCharPressed()
			if !editor.clearedInputs[label] && key > 0 {
				*value = ""
				editor.clearedInputs[label] = true
			}

			for key > 0 {
				if key >= 32 && key <= 126 {
					*value += string(key)
				}
				key = rl.GetCharPressed()
			}
			if rl.IsKeyPressed(rl.KeyBackspace) && len(*value) > 0 {
				*value = (*value)[:len(*value)-1]
			}
		}
	}

	// Save/Cancel buttons
	btnWidth := 80
	btnHeight := 30
	saveBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - btnWidth*3 - padding*2 - 8),
		Y:      float32(dialogY + dialogHeight - btnHeight - padding),
		Width:  float32(btnWidth),
		Height: float32(btnHeight),
	}
	cancelBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - btnWidth*2 - padding - 8),
		Y:      float32(dialogY + dialogHeight - btnHeight - padding),
		Width:  float32(btnWidth),
		Height: float32(btnHeight),
	}
	advancedBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - btnWidth - 8),
		Y:      float32(dialogY + dialogHeight - btnHeight - padding),
		Width:  float32(btnWidth),
		Height: float32(btnHeight),
	}

	rl.DrawRectangleRec(saveBtn, rl.LightGray)
	rl.DrawRectangleRec(cancelBtn, rl.LightGray)
	rl.DrawRectangleRec(advancedBtn, rl.LightGray)
	rl.DrawText("Save", int32(saveBtn.X+20), int32(saveBtn.Y+8), 16, rl.Black)
	rl.DrawText("Cancel", int32(cancelBtn.X+15), int32(cancelBtn.Y+8), 16, rl.Black)
	rl.DrawText("Advanced", int32(advancedBtn.X+4), int32(advancedBtn.Y+8), 16, rl.Black)

	// Handle button clicks
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), cancelBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.closeTextureEditor()
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), saveBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Update the texture frame with new values
		// Ensure the frame exists before trying to access it
		if editor.texIndex < len(editor.tile.Textures) && editor.frameIndex < len(editor.tile.Textures[editor.texIndex].Frames) {
			frame := &editor.tile.Textures[editor.texIndex].Frames[editor.frameIndex]
			frame.Rotation, _ = strconv.ParseFloat(editor.rotation, 64)
			frame.Scale, _ = strconv.ParseFloat(editor.scale, 64)
			frame.OffsetX, _ = strconv.ParseFloat(editor.offsetX, 64)
			frame.OffsetY, _ = strconv.ParseFloat(editor.offsetY, 64)
			r, _ := strconv.Atoi(editor.tintR)
			g, _ := strconv.Atoi(editor.tintG)
			b, _ := strconv.Atoi(editor.tintB)
			a, _ := strconv.Atoi(editor.tintA)
			frame.Tint = rl.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
		} else {
			fmt.Println("Error: Texture or frame index out of bounds during save.")
		}
		m.closeTextureEditor()
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), advancedBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Initialize advanced editor state when opening it
		tex := editor.tile.Textures[editor.texIndex]
		if tex.IsComplex && len(tex.Frames) > 0 {
			editor.advAnimationTimeStr = fmt.Sprintf("%.2f", tex.AnimationTime)
			editor.advFrameCountStr = fmt.Sprintf("%d", len(tex.Frames))
			editor.advSelectedFrames = make([]string, len(tex.Frames))
			for i, frame := range tex.Frames {
				editor.advSelectedFrames[i] = frame.Name
			}
		} else {
			editor.advAnimationTimeStr = "0.5"           // Default animation time
			editor.advFrameCountStr = "2"                // Default frame count
			editor.advSelectedFrames = make([]string, 2) // Initialize based on default count
		}
		editor.advSelectingFrameIndex = -1
		m.uiState.showAdvancedEditor = true
		m.uiState.activeInput = ""
	}
	if m.uiState.showAdvancedEditor {
		m.renderAdvancedEditor()
	}
}

func (m *MapMaker) renderAdvancedEditor() {
	editor := m.uiState.textureEditor
	if editor == nil {
		return
	}

	dialogWidth := 600
	dialogHeight := 500
	dialogX := (rl.GetScreenWidth() - dialogWidth) / 2
	dialogY := (rl.GetScreenHeight() - dialogHeight) / 2

	// Draw dialog background
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.Black, 0.7))
	rl.DrawRectangle(int32(dialogX), int32(dialogY), int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(dialogX),
		Y:      float32(dialogY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}, 1, rl.Gray)

	// Title
	rl.DrawText("Complex Texture Editor", int32(dialogX+180), int32(dialogY+15), 20, rl.Black)

	// Back button
	backBtn := rl.Rectangle{
		X:      float32(dialogX + 10),
		Y:      float32(dialogY + 10),
		Width:  80,
		Height: 30,
	}
	rl.DrawRectangleRec(backBtn, rl.LightGray)
	rl.DrawText("Back", int32(backBtn.X+20), int32(backBtn.Y+8), 16, rl.Black)

	// Exit button
	exitBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - 90),
		Y:      float32(dialogY + 10),
		Width:  80,
		Height: 30,
	}
	rl.DrawRectangleRec(exitBtn, rl.Red)
	rl.DrawText("Exit", int32(exitBtn.X+25), int32(exitBtn.Y+8), 16, rl.White)

	contentY := dialogY + 60
	padding := 20
	labelWidth := 120
	inputWidth := 80
	inputHeight := 30

	// Helper function for input fields in this context
	createAdvInput := func(label string, value *string, yPos int, inputID string) {
		rl.DrawText(label, int32(dialogX+padding), int32(yPos+8), 16, rl.Black)
		inputRect := rl.Rectangle{
			X:      float32(dialogX + padding + labelWidth),
			Y:      float32(yPos),
			Width:  float32(inputWidth),
			Height: float32(inputHeight),
		}
		rl.DrawRectangleRec(inputRect, rl.LightGray)
		rl.DrawText(*value, int32(inputRect.X+5), int32(inputRect.Y+8), 16, rl.Black)

		// Handle input focus and text input
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), inputRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.uiState.activeInput = inputID
		}
		if m.uiState.activeInput == inputID {
			rl.DrawRectangleLinesEx(inputRect, 2, rl.Blue)

			key := rl.GetCharPressed()
			for key > 0 {
				if (key >= '0' && key <= '9') || (inputID == "advAnimTime" && key == '.') {
					*value += string(key)
				}
				key = rl.GetCharPressed()
			}
			if rl.IsKeyPressed(rl.KeyBackspace) && len(*value) > 0 {
				*value = (*value)[:len(*value)-1]
			}
		}
	}

	// Animation Time Input - Disable if frameCount is 1
	animTimeLabelColor := rl.Black
	animTimeInputColor := rl.LightGray
	animTimeTextColor := rl.Black
	frameCount, _ := strconv.Atoi(editor.advFrameCountStr)
	isAnimTimeDisabled := frameCount == 1

	if isAnimTimeDisabled {
		animTimeLabelColor = rl.Gray
		animTimeInputColor = rl.DarkGray
		animTimeTextColor = rl.Gray
		if m.uiState.activeInput == "advAnimTime" {
			m.uiState.activeInput = ""
		}
	}

	rl.DrawText("Anim Time (s):", int32(dialogX+padding), int32(contentY+8), 16, animTimeLabelColor)
	animTimeInputRect := rl.Rectangle{
		X:      float32(dialogX + padding + labelWidth),
		Y:      float32(contentY),
		Width:  float32(inputWidth),
		Height: float32(inputHeight),
	}
	rl.DrawRectangleRec(animTimeInputRect, animTimeInputColor)
	rl.DrawText(editor.advAnimationTimeStr, int32(animTimeInputRect.X+5), int32(animTimeInputRect.Y+8), 16, animTimeTextColor)

	if !isAnimTimeDisabled {
		// Handle input focus and text input only if not disabled
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), animTimeInputRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.uiState.activeInput = "advAnimTime"
		}
		if m.uiState.activeInput == "advAnimTime" {
			rl.DrawRectangleLinesEx(animTimeInputRect, 2, rl.Blue)

			key := rl.GetCharPressed()
			for key > 0 {
				if (key >= '0' && key <= '9') || key == '.' {
					editor.advAnimationTimeStr += string(key)
				}
				key = rl.GetCharPressed()
			}
			if rl.IsKeyPressed(rl.KeyBackspace) && len(editor.advAnimationTimeStr) > 0 {
				editor.advAnimationTimeStr = editor.advAnimationTimeStr[:len(editor.advAnimationTimeStr)-1]
			}
		}
	}
	contentY += inputHeight + padding

	// Frame Count Input
	createAdvInput("Frame Count:", &editor.advFrameCountStr, contentY, "advFrameCount")
	contentY += inputHeight + padding

	// Re-parse frame count and validate
	frameCount, err := strconv.Atoi(editor.advFrameCountStr)
	if err != nil || frameCount <= 0 {
		frameCount = 0
		if editor.advFrameCountStr != "" {
			rl.DrawText("Invalid frame count", int32(dialogX+padding+labelWidth+inputWidth+10), int32(contentY-inputHeight-padding+8), 16, rl.Red)
		}
	}

	// Adjust selectedFrames slice size if frameCount changed
	if len(editor.advSelectedFrames) != frameCount && frameCount >= 0 {
		newFrames := make([]string, frameCount)
		copy(newFrames, editor.advSelectedFrames)
		editor.advSelectedFrames = newFrames
	}

	// Frame Selection Area
	rl.DrawText("Animation Frames:", int32(dialogX+padding), int32(contentY), 16, rl.Black)
	contentY += 30

	framePreviewSize := 40
	framePadding := 10
	framesPerRow := (dialogWidth - padding*2) / (framePreviewSize + framePadding)
	frameStartX := dialogX + padding

	for i := 0; i < frameCount; i++ {
		row := i / framesPerRow
		col := i % framesPerRow
		frameX := frameStartX + col*(framePreviewSize+framePadding)
		frameY := contentY + row*(framePreviewSize+framePadding)

		frameRect := rl.Rectangle{
			X:      float32(frameX),
			Y:      float32(frameY),
			Width:  float32(framePreviewSize),
			Height: float32(framePreviewSize),
		}

		// Draw preview box
		rl.DrawRectangleRec(frameRect, rl.LightGray)
		rl.DrawRectangleLinesEx(frameRect, 1, rl.Gray)

		// Draw selected texture preview if available
		if i < len(editor.advSelectedFrames) && editor.advSelectedFrames[i] != "" {
			texName := editor.advSelectedFrames[i]
			texInfo, err := m.resources.GetTexture("default", texName)
			if err == nil {
				// Draw texture centered in the box
				scale := float32(framePreviewSize) / texInfo.Region.Width
				if texInfo.Region.Height*scale > float32(framePreviewSize) {
					scale = float32(framePreviewSize) / texInfo.Region.Height
				}
				drawWidth := texInfo.Region.Width * scale
				drawHeight := texInfo.Region.Height * scale
				drawX := frameRect.X + (frameRect.Width-drawWidth)/2
				drawY := frameRect.Y + (frameRect.Height-drawHeight)/2

				rl.DrawTexturePro(
					texInfo.Texture,
					texInfo.Region,
					rl.Rectangle{X: drawX, Y: drawY, Width: drawWidth, Height: drawHeight},
					rl.Vector2{}, 0, rl.White,
				)
			} else {
				// Draw placeholder if texture is missing but selected
				rl.DrawText("?", int32(int(frameRect.X)+framePreviewSize/2-5), int32(int(frameRect.Y)+framePreviewSize/2-10), 20, rl.Red)
			}
		} else {
			// Draw placeholder for empty slot
			rl.DrawText("+", int32(int(frameRect.X)+framePreviewSize/2-5), int32(int(frameRect.Y)+framePreviewSize/2-10), 20, rl.DarkGray)
		}

		// Handle click to select texture for this frame
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), frameRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			editor.advSelectingFrameIndex = i
			m.showResourceViewer = true // Open resource viewer to select texture
			m.uiState.activeInput = ""  // Deactivate text inputs
		}
	}

	saveBtnAdv := rl.Rectangle{
		X:      float32(dialogX + dialogWidth/2 - 40),
		Y:      float32(dialogY + dialogHeight - 40),
		Width:  80,
		Height: 30,
	}
	rl.DrawRectangleRec(saveBtnAdv, rl.Green)
	rl.DrawText("Save", int32(saveBtnAdv.X+20), int32(saveBtnAdv.Y+8), 16, rl.White)

	// Handle button clicks
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), backBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.closeTextureEditor()
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), exitBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.closeAllEditors()
		m.uiState.activeInput = ""
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), saveBtnAdv) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Validate inputs
		animTime := 0.0 // Default value
		var timeErr error
		if frameCount > 0 { // Only parse animTime if frameCount is greater than 0
			animTime, timeErr = strconv.ParseFloat(editor.advAnimationTimeStr, 64)
		}

		allFramesSelected := true
		if frameCount <= 0 {
			allFramesSelected = false
		}
		for i := 0; i < frameCount; i++ {
			if i >= len(editor.advSelectedFrames) || editor.advSelectedFrames[i] == "" {
				allFramesSelected = false
				break
			}
		}

		if (frameCount == 1 || (timeErr == nil && animTime > 0)) && allFramesSelected {
			// Apply changes to the TileTexture
			if editor.texIndex < len(editor.tile.Textures) {
				tex := editor.tile.Textures[editor.texIndex]
				// Use properties from the *original* first frame if available, otherwise defaults
				originalRotation := 0.0
				originalScale := 1.0
				originalOffsetX := 0.0
				originalOffsetY := 0.0
				originalTint := rl.White
				if len(editor.tile.Textures[editor.texIndex].Frames) > 0 {
					originalFrame := editor.tile.Textures[editor.texIndex].Frames[0]
					originalRotation = originalFrame.Rotation
					originalScale = originalFrame.Scale
					originalOffsetX = originalFrame.OffsetX
					originalOffsetY = originalFrame.OffsetY
					originalTint = originalFrame.Tint
				}

				tex.IsComplex = frameCount > 1
				tex.AnimationTime = animTime                              // Will be 0 if frameCount is 1
				tex.CurrentFrame = 0                                      // Reset animation state
				tex.Frames = make([]beam.TileTextureFrame, 0, frameCount) // Clear existing frames

				for i := 0; i < frameCount; i++ {
					newFrame := beam.TileTextureFrame{
						Name:     editor.advSelectedFrames[i],
						Rotation: originalRotation, // Apply original/default properties
						Scale:    originalScale,
						OffsetX:  originalOffsetX,
						OffsetY:  originalOffsetY,
						Tint:     originalTint,
					}
					tex.Frames = append(tex.Frames, newFrame)
				}

				m.showToast("Texture properties saved!", ToastSuccess)
				m.uiState.showAdvancedEditor = false // Close advanced editor on successful save
				m.uiState.textureEditor = nil        // Close simple texture editor as well
				m.uiState.activeInput = ""
			} else {
				m.showToast("Error: Texture index out of bounds.", ToastError)
			}
		} else {
			// Show error message
			errMsg := "Invalid input:"
			if frameCount != 1 && (timeErr != nil || animTime <= 0) {
				errMsg += " Invalid time."
			}
			if frameCount <= 0 {
				errMsg += " Frame count > 0."
			}
			if !allFramesSelected {
				errMsg += " Select all frames."
			}
			m.showToast(errMsg, ToastError)
		}
	}
}

func (m *MapMaker) closeTextureEditor() {
	m.uiState.textureEditor = nil
	m.uiState.showAdvancedEditor = false
	m.uiState.activeInput = ""
}

func (m *MapMaker) closeAllEditors() {
	m.closeTextureEditor()
	m.showTileInfo = false
	m.uiState.activeInput = ""
}
