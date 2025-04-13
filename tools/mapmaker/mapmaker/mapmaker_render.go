package mapmaker

import (
	"fmt"

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
				rl.DrawRectangleLinesEx(rl.Rectangle{
					X:      float32(highlightX),
					Y:      float32(highlightY),
					Width:  float32(m.uiState.tileSize),
					Height: float32(m.uiState.tileSize),
				}, 2, rl.Black)
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

			// Apply scale, tint, and offset to the destination rectangle
			if frame.Scale == 0 {
				frame.Scale = 1
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
	rl.DrawText(fmt.Sprintf("%dpx", m.uiState.tileSize), 48, 68, 12, rl.DarkGray)

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

	if m.showResourceViewer {
		m.renderResourceViewer()
	}

	if m.showTileInfo {
		m.renderTileInfoPopup()
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

	// Draw dialog background and frame
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.Black, 0.7))

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
		return
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
			if m.uiState.activeTexture != nil && m.uiState.activeTexture.Name == texInfo.Name {
				rl.DrawRectangleLinesEx(clickArea, 2, rl.Blue)
			}

			if rl.CheckCollisionPointRec(rl.GetMousePosition(), clickArea) &&
				rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				tex, err := m.resources.GetTexture("default", texInfo.Name)
				if err != nil {
					fmt.Println("Error getting texture:", err)
				} else {
					m.handleTextureSelect(&tex)
					m.showResourceViewer = false
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

	for _, tex := range tile.Textures {
		rl.DrawText(fmt.Sprintf("- Complex: %t", tex.IsComplex), m.uiState.tileInfoPopupX+padding+10, textY, 14, rl.DarkGray)
		textY += 20

		for _, frame := range tex.Frames {
			warningText := ""
			textColor := rl.DarkGray

			if m.tileGrid.missingResourceTiles.Contains(pos, frame.Name) {
				warningText = " !"
				textColor = rl.Yellow
			}

			rl.DrawText(fmt.Sprintf("  - %s (%.1f°) Scale: %.2f Offset: (%.2f, %.2f)%s",
				frame.Name, frame.Rotation, frame.Scale, frame.OffsetX, frame.OffsetY, warningText),
				m.uiState.tileInfoPopupX+padding+5, textY, 12, textColor)
			textY += 20
		}
	}

	// Draw close button
	closeBtn := rl.Rectangle{
		X:      float32(m.uiState.tileInfoPopupX + int32(dialogWidth) - 30),
		Y:      float32(m.uiState.tileInfoPopupY + 5),
		Width:  25,
		Height: 25,
	}
	rl.DrawRectangleRec(closeBtn, rl.LightGray)
	rl.DrawText("×", int32(closeBtn.X+7), int32(closeBtn.Y+2), 20, rl.DarkGray)
}
