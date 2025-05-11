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
	for _, layer := range beam.OrderedLayers() {
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
				m.renderGridTile(pos, beam.Position{X: x, Y: y}, tile, layer)

				// Draw any NPC's on the map
				for _, npc := range m.tileGrid.NPCs {
					if npc.Pos.X == x && npc.Pos.Y == y {
						npcX := startX + (x-viewStartX)*m.uiState.tileSize
						npcY := startY + (y-viewStartY)*m.uiState.tileSize
						m.resources.RenderNPC(npc, rl.Rectangle{
							X:      float32(npcX),
							Y:      float32(npcY),
							Width:  float32(m.uiState.tileSize),
							Height: float32(m.uiState.tileSize),
						}, m.uiState.tileSize)
					}
				}

				// Draw any items on the map
				for _, item := range m.tileGrid.Items {
					m.resources.RenderItem(&item, rl.Rectangle{
						X:      float32(item.Position.X * m.uiState.tileSize),
						Y:      float32(item.Position.Y * m.uiState.tileSize),
						Width:  float32(m.uiState.tileSize),
						Height: float32(m.uiState.tileSize),
					}, m.uiState.tileSize)
				}
			}
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

func (m *MapMaker) renderGridTile(pos rl.Rectangle, pos2d beam.Position, tile beam.Tile, layer beam.Layer) {
	if len(tile.Textures) == 0 {
		return
	}

	for _, tex := range tile.Textures {
		if len(tex.Frames) == 0 {
			continue
		} else if tex.Layer != layer {
			continue
		}

		// If the texture isn't complex, we can just draw the frames on top of each other.
		if !tex.IsAnimated {
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
					Width:  pos.Width * float32(frame.ScaleX),
					Height: pos.Height * float32(frame.ScaleY),
				}

				if frame.MirrorX {
					info.Region.Width = -info.Region.Width
				}
				if frame.MirrorY {
					info.Region.Height = -info.Region.Height
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
				Width:  pos.Width * float32(frame.ScaleX),
				Height: pos.Height * float32(frame.ScaleY),
			}

			if frame.MirrorY {
				destRect.Width = -destRect.Width
			}
			if frame.MirrorX {
				destRect.Height = -destRect.Height
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

	// Draw special tile outlines
	if m.uiState.selectedTool != "gridlines" && tile.Type == beam.WallTile {
		rl.DrawRectangleLinesEx(pos, 2, rl.Brown)
	}

	if m.uiState.selectedTool != "gridlines" && pos2d.X != 0 && pos2d.Y != 0 {
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
	tileSmallerBtn, tileLargerBtn, widthSmallerBtn, widthLargerBtn, heightSmallerBtn, heightLargerBtn, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, resetBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn, gridlinesBtn, npcBtn := m.getUIButtons()

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
	m.drawToolIcons(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn, gridlinesBtn, npcBtn)

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

	// Render NPC editor if active
	if m.uiState.npcEditor != nil && m.uiState.npcEditor.visible {
		m.renderNPCEditor()
	}

	if m.showResourceViewer {
		m.renderResourceViewer()
	}

	if m.uiState.showNPCList {
		m.renderNPCList()
	}

	// Draw status bar
	rl.DrawRectangle(0, m.window.height-int32(m.uiState.statusBarHeight),
		m.window.width, int32(m.uiState.statusBarHeight), rl.RayWhite)
	rl.DrawLine(0, m.window.height-int32(m.uiState.statusBarHeight),
		m.window.width, m.window.height-int32(m.uiState.statusBarHeight), rl.LightGray)

}

func (m *MapMaker) drawToolIcons(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn, layersBtn, locationBtn, gridlinesBtn, npcBtn IconButton) {
	m.drawIconButton(paintbrushBtn, rl.LightGray)
	m.drawIconButton(paintbucketBtn, rl.LightGray)
	m.drawIconButton(eraseBtn, rl.LightGray)
	m.drawIconButton(selectBtn, rl.LightGray)
	m.drawIconButton(layersBtn, rl.LightGray)
	m.drawIconButton(locationBtn, rl.LightGray)
	m.drawIconButton(gridlinesBtn, rl.LightGray)
	m.drawIconButton(npcBtn, rl.LightGray)

	// Draw tools with selection highlight
	toolButtons := map[string]IconButton{
		"paintbrush":   paintbrushBtn,
		"paintbucket":  paintbucketBtn,
		"eraser":       eraseBtn,
		"pencileraser": eraseBtn,
		"select":       selectBtn,
		"selectall":    selectBtn,
		"layers":       layersBtn,
		"location":     locationBtn,
		"gridlines":    gridlinesBtn,
		"npc":          npcBtn,
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

	timeSinceOpen := rl.GetTime() - m.uiState.resourceViewerOpenTime
	canAcceptClicks := timeSinceOpen >= 0.5

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
		itemHeight := int32(60) // Increased height to accommodate additional info
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

			// Draw grid size and margin info
			gridInfo := fmt.Sprintf("Grid: %dx%d  Margin: %d",
				texInfo.GridSizeX, texInfo.GridSizeY, texInfo.Margin)
			rl.DrawText(gridInfo, int32(itemRect.X+10), int32(itemRect.Y+28), 14, rl.DarkGray)

			// Delete button
			deleteBtn := rl.Rectangle{
				X:      itemRect.X + itemRect.Width - 60,
				Y:      itemRect.Y + 10,
				Width:  50,
				Height: 26,
			}
			rl.DrawRectangleRec(deleteBtn, rl.Red)
			rl.DrawText("Delete", int32(deleteBtn.X+3), int32(deleteBtn.Y+5), 14, rl.White)

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

			if canAcceptClicks && rl.CheckCollisionPointRec(rl.GetMousePosition(), clickArea) &&
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

	// Calculate total content height first
	var totalHeight int32 = 60
	tempTile := m.tileGrid.Tiles[m.uiState.tileInfoPos[0].Y][m.uiState.tileInfoPos[0].X]
	for _, tex := range tempTile.Textures {
		totalHeight += 35
		for range tex.Frames {
			totalHeight += 55
		}
	}

	// Add scroll state if it doesn't exist
	if m.uiState.tileInfoScrollY == 0 {
		m.uiState.tileInfoScrollY = 0
	}

	// Handle mouse wheel for scrolling
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), rl.Rectangle{
		X:      float32(m.uiState.tileInfoPopupX),
		Y:      float32(m.uiState.tileInfoPopupY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}) {
		wheel := rl.GetMouseWheelMove()
		m.uiState.tileInfoScrollY -= int32(wheel * 20)
	}

	// Clamp scroll value
	maxScroll := totalHeight - int32(dialogHeight) + 40
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.uiState.tileInfoScrollY < 0 {
		m.uiState.tileInfoScrollY = 0
	}
	if m.uiState.tileInfoScrollY > maxScroll {
		m.uiState.tileInfoScrollY = maxScroll
	}

	// Draw popup background
	rl.DrawRectangle(m.uiState.tileInfoPopupX, m.uiState.tileInfoPopupY, int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(m.uiState.tileInfoPopupX),
		Y:      float32(m.uiState.tileInfoPopupY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}, 1, rl.Gray)

	// Begin scissor mode for content clipping
	rl.BeginScissorMode(
		m.uiState.tileInfoPopupX,
		m.uiState.tileInfoPopupY+30,
		int32(dialogWidth),
		int32(dialogHeight)-40,
	)

	// Draw content
	padding := int32(10)
	// Adjust initial Y position by scroll offset
	textY := m.uiState.tileInfoPopupY + padding - m.uiState.tileInfoScrollY

	// Draw tile type
	tile := m.tileGrid.Tiles[m.uiState.tileInfoPos[0].Y][m.uiState.tileInfoPos[0].X]
	rl.DrawText(fmt.Sprintf("Tile Type: %d", tile.Type), m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 25

	// Draw tile position - show "many" if multiple tiles selected
	posText := "Position: many"
	if len(m.uiState.tileInfoPos) == 1 {
		posText = fmt.Sprintf("Position: (%d, %d)", tile.Pos.X, tile.Pos.Y)
	}
	rl.DrawText(posText, m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 25

	// Draw textures
	rl.DrawText("Textures:", m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 20

	for texIndex, tex := range tile.Textures {
		// Draw complex text and edit button side by side
		complexText := fmt.Sprintf("- Complex: %t", tex.IsAnimated)
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
				tile:          &m.tileGrid.Tiles[m.uiState.tileInfoPos[0].Y][m.uiState.tileInfoPos[0].X],
				texIndex:      texIndex,
				frameIndex:    0,
				clearedInputs: make(map[string]bool),
			}

			// Set up editor fields based on whether texture is complex
			if tex.IsAnimated && len(tex.Frames) > 0 {
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
				m.uiState.advancedEditorOpenTime = rl.GetTime()
				m.uiState.activeInput = ""
			} else {
				// Use simple editor for non-complex textures
				editor.layer = tex.Layer
				if len(tex.Frames) > 0 {
					firstFrame := tex.Frames[0]
					editor.rotation = fmt.Sprintf("%.1f", firstFrame.Rotation)
					editor.scalex = fmt.Sprintf("%.2f", firstFrame.ScaleX)
					editor.scaley = fmt.Sprintf("%.2f", firstFrame.ScaleY)
					editor.offsetX = fmt.Sprintf("%.2f", firstFrame.OffsetX)
					editor.offsetY = fmt.Sprintf("%.2f", firstFrame.OffsetY)
					editor.mirrorX = firstFrame.MirrorX
					editor.mirrorY = firstFrame.MirrorY
					editor.tintR = fmt.Sprintf("%d", firstFrame.Tint.R)
					editor.tintG = fmt.Sprintf("%d", firstFrame.Tint.G)
					editor.tintB = fmt.Sprintf("%d", firstFrame.Tint.B)
					editor.tintA = fmt.Sprintf("%d", firstFrame.Tint.A)
				} else {
					editor.rotation = "0.0"
					editor.scalex = "1.0"
					editor.scaley = "1.0"
					editor.offsetX = "0.0"
					editor.offsetY = "0.0"
					editor.tintR = "255"
					editor.tintG = "255"
					editor.tintB = "255"
					editor.tintA = "255"
					editor.mirrorX = false
					editor.mirrorY = false
				}
				m.uiState.textureEditor = editor
			}
		}
		textY += 20

		for _, frame := range tex.Frames {
			warningText := ""
			textColor := rl.DarkGray

			if m.tileGrid.missingResourceTiles.Contains(pos[0], frame.Name) {
				warningText = " !"
				textColor = rl.Yellow
			}

			rl.DrawText(fmt.Sprintf("  - %s (%.1f°) Scale: %.2f Offset: (%.2f, %.2f)",
				frame.Name, frame.Rotation, frame.ScaleX, frame.OffsetX, frame.OffsetY),
				m.uiState.tileInfoPopupX+padding+5, textY, 12, textColor)
			textY += 15

			// Add tint information on next line, indented further
			rl.DrawText(fmt.Sprintf("    Tint: R:%d G:%d B:%d A:%d%s",
				frame.Tint.R, frame.Tint.G, frame.Tint.B, frame.Tint.A, warningText),
				m.uiState.tileInfoPopupX+padding+5, textY, 12, textColor)
			textY += 15

			// Add layer information
			rl.DrawText(fmt.Sprintf("Layer: %s", tex.Layer.String()),
				m.uiState.tileInfoPopupX+padding+25, textY, 12, textColor)
			textY += 25
		}
	}

	rl.EndScissorMode()

	// Draw scroll indicators if needed
	if maxScroll > 0 {
		scrollPct := float32(m.uiState.tileInfoScrollY) / float32(maxScroll)
		scrollBarHeight := float32(dialogHeight-40) * (float32(dialogHeight-40) / float32(totalHeight))
		scrollBarY := float32(m.uiState.tileInfoPopupY+30) + scrollPct*float32(int32(dialogHeight)-40-int32(scrollBarHeight))

		// Draw scroll bar
		rl.DrawRectangle(
			m.uiState.tileInfoPopupX+int32(dialogWidth)-8,
			int32(scrollBarY),
			5,
			int32(scrollBarHeight),
			rl.Gray)
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

type NPCEditorState struct {
	visible     bool
	spawnPos    beam.Position
	name        string
	health      string
	attack      string
	defense     string
	attackSpeed string
	attackRange string
	moveSpeed   string
	isHostile   bool
	aggroRange  string
	wanderRange string

	// Texture editing state
	editingDirection       beam.Direction
	textures               *beam.NPCTexture
	frameCountStr          string
	animationTimeStr       string
	selectedFrames         []string
	advSelectingFrameIndex int

	// String representations for input fields
	spawnXStr string
	spawnYStr string

	attackable bool
	impassable bool

	// Frame editing fields
	selectedFrameIndex int // Track which frame is selected for editing
	frameRotation      string
	frameScaleX        string
	frameScaleY        string
	frameOffsetX       string
	frameOffsetY       string
	frameMirrorX       bool
	frameMirrorY       bool
	frameTintR         string
	frameTintG         string
	frameTintB         string
	frameTintA         string
}

func (m *MapMaker) renderNPCEditor() {
	editor := m.uiState.npcEditor

	// Dialog dimensions and position
	dialogWidth := 800
	dialogHeight := 650
	dialogX := (rl.GetScreenWidth() - dialogWidth) / 2
	dialogY := (rl.GetScreenHeight() - dialogHeight) / 2

	// Draw semi-transparent background
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.Black, 0.7))

	// Draw main dialog box
	rl.DrawRectangle(int32(dialogX), int32(dialogY), int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(dialogX),
		Y:      float32(dialogY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}, 1, rl.Gray)

	// Title
	rl.DrawText("NPC Editor", int32(dialogX+20), int32(dialogY+20), 24, rl.Black)

	// Layout constants
	const (
		padding     = 20
		labelWidth  = 120
		inputWidth  = 150
		inputHeight = 30
		columnWidth = 350
	)

	// Start positions for the two columns
	leftX := dialogX + padding
	rightX := dialogX + columnWidth + padding
	startY := dialogY + 80

	// Helper function for input fields
	createNPCInput := func(label string, value *string, x, y int, numeric bool) {
		rl.DrawText(label, int32(x), int32(y+8), 16, rl.Black)
		inputRect := rl.Rectangle{
			X:      float32(x + labelWidth),
			Y:      float32(y),
			Width:  float32(inputWidth),
			Height: float32(inputHeight),
		}

		rl.DrawRectangleRec(inputRect, rl.LightGray)
		rl.DrawText(*value, int32(inputRect.X+5), int32(inputRect.Y+8), 16, rl.Black)

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), inputRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.uiState.activeNPCInput = label
		}

		if m.uiState.activeNPCInput == label {
			rl.DrawRectangleLinesEx(inputRect, 2, rl.Blue)

			key := rl.GetCharPressed()
			for key > 0 {
				if numeric {
					if (key >= '0' && key <= '9') || key == '.' {
						*value += string(key)
					}
				} else {
					if key >= 32 && key <= 126 {
						*value += string(key)
					}
				}
				key = rl.GetCharPressed()
			}
			if rl.IsKeyPressed(rl.KeyBackspace) && len(*value) > 0 {
				*value = (*value)[:len(*value)-1]
			}
		}

		if label == "Animation Time" {
			animTime, err := strconv.ParseFloat(*value, 64)
			if err == nil && animTime >= 0 {
				if editor.editingDirection == beam.DirUp {
					editor.textures.Up.AnimationTime = animTime
				} else if editor.editingDirection == beam.DirDown {
					editor.textures.Down.AnimationTime = animTime
				} else if editor.editingDirection == beam.DirLeft {
					editor.textures.Left.AnimationTime = animTime
				} else if editor.editingDirection == beam.DirRight {
					editor.textures.Right.AnimationTime = animTime
				}
			}
		}
	}

	// Left column - Basic attributes
	y := startY

	editor.spawnXStr = fmt.Sprintf("%d", editor.spawnPos.X)
	editor.spawnYStr = fmt.Sprintf("%d", editor.spawnPos.Y)
	createNPCInput("Name", &editor.name, leftX, y, false)
	y += inputHeight + padding
	createNPCInput("Health", &editor.health, leftX, y, true)
	y += inputHeight + padding
	createNPCInput("Attack", &editor.attack, leftX, y, true)
	y += inputHeight + padding
	createNPCInput("Defense", &editor.defense, leftX, y, true)
	y += inputHeight + padding
	createNPCInput("Attack Speed", &editor.attackSpeed, leftX, y, true)
	y += inputHeight + padding
	createNPCInput("Attack Range", &editor.attackRange, leftX, y, true)
	y += inputHeight + padding
	createNPCInput("Spawn X", &editor.spawnXStr, leftX, y, true) // Added Spawn X
	y += inputHeight + padding
	createNPCInput("Spawn Y", &editor.spawnYStr, leftX, y, true) // Added Spawn Y

	y += inputHeight + padding
	checkboxRect := rl.Rectangle{
		X:      float32(leftX + labelWidth),
		Y:      float32(y),
		Width:  float32(inputHeight),
		Height: float32(inputHeight),
	}
	rl.DrawRectangleRec(checkboxRect, rl.LightGray)
	if editor.attackable {
		rl.DrawRectangle(
			int32(checkboxRect.X+5),
			int32(checkboxRect.Y+5),
			int32(checkboxRect.Width-10),
			int32(checkboxRect.Height-10),
			rl.Black,
		)
	}
	rl.DrawText("Attackable", int32(leftX), int32(y+8), 16, rl.Black)

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), checkboxRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		editor.attackable = !editor.attackable
	}

	// Add impassable checkbox
	y += inputHeight + padding
	checkboxRect = rl.Rectangle{
		X:      float32(leftX + labelWidth),
		Y:      float32(y),
		Width:  float32(inputHeight),
		Height: float32(inputHeight),
	}
	rl.DrawRectangleRec(checkboxRect, rl.LightGray)
	if editor.impassable {
		rl.DrawRectangle(
			int32(checkboxRect.X+5),
			int32(checkboxRect.Y+5),
			int32(checkboxRect.Width-10),
			int32(checkboxRect.Height-10),
			rl.Black,
		)
	}
	rl.DrawText("Impassable", int32(leftX), int32(y+8), 16, rl.Black)

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), checkboxRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		editor.impassable = !editor.impassable
	}

	y += inputHeight + padding
	createNPCInput("Wander Range", &editor.wanderRange, leftX, y, true)

	// Right column - Movement and behavior
	y = startY
	createNPCInput("Move Speed", &editor.moveSpeed, rightX, y, true)
	y += inputHeight + padding
	createNPCInput("Aggro Range", &editor.aggroRange, rightX, y, true)
	y += inputHeight + padding

	// Hostile checkbox
	checkboxRect = rl.Rectangle{
		X:      float32(rightX + labelWidth),
		Y:      float32(y),
		Width:  float32(inputHeight),
		Height: float32(inputHeight),
	}
	rl.DrawRectangleRec(checkboxRect, rl.LightGray)
	if editor.isHostile {
		rl.DrawRectangle(
			int32(checkboxRect.X+5),
			int32(checkboxRect.Y+5),
			int32(checkboxRect.Width-10),
			int32(checkboxRect.Height-10),
			rl.Black,
		)
	}
	rl.DrawText("Hostile", int32(rightX), int32(y+8), 16, rl.Black)

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), checkboxRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		editor.isHostile = !editor.isHostile
	}

	// Direction selector
	y += inputHeight + padding*2
	rl.DrawText("Direction Textures", int32(rightX), int32(y), 16, rl.Black)
	y += 25

	dirBtnSize := int32(60)
	dirBtnSpacing := int32(5)
	dirStartX := rightX + (columnWidth-int(dirBtnSize*3+dirBtnSpacing*2))/2

	// Up button
	upBtn := rl.Rectangle{
		X:      float32(int32(dirStartX) + dirBtnSize + dirBtnSpacing),
		Y:      float32(y),
		Width:  float32(dirBtnSize),
		Height: float32(dirBtnSize),
	}

	// Left button
	leftBtn := rl.Rectangle{
		X:      float32(dirStartX),
		Y:      float32(int32(y) + dirBtnSize + dirBtnSpacing),
		Width:  float32(dirBtnSize),
		Height: float32(dirBtnSize),
	}

	// Down button
	downBtn := rl.Rectangle{
		X:      float32(int32(dirStartX) + dirBtnSize + dirBtnSpacing),
		Y:      float32(int32(y) + dirBtnSize + dirBtnSpacing),
		Width:  float32(dirBtnSize),
		Height: float32(dirBtnSize),
	}

	// Right button
	rightBtn := rl.Rectangle{
		X:      float32(int32(dirStartX) + (dirBtnSize+dirBtnSpacing)*2),
		Y:      float32(int32(y) + dirBtnSize + dirBtnSpacing),
		Width:  float32(dirBtnSize),
		Height: float32(dirBtnSize),
	}

	// Draw direction buttons with textures if set
	drawDirButton := func(btn rl.Rectangle, dir beam.Direction, label string) {
		isSelected := editor.editingDirection == dir
		btnColor := rl.LightGray
		if isSelected {
			btnColor = rl.Blue
		}

		rl.DrawRectangleRec(btn, btnColor)

		var tex *beam.AnimatedTexture
		switch dir {
		case beam.DirUp:
			tex = editor.textures.Up
		case beam.DirDown:
			tex = editor.textures.Down
		case beam.DirLeft:
			tex = editor.textures.Left
		case beam.DirRight:
			tex = editor.textures.Right
		}

		if len(tex.Frames) > 0 && tex.Frames[0].Name != "" {
			info, err := m.resources.GetTexture("default", tex.Frames[0].Name)
			if err == nil {
				scale := float32(dirBtnSize-10) / info.Region.Width
				if info.Region.Height*scale > float32(dirBtnSize-10) {
					scale = float32(dirBtnSize-10) / info.Region.Height
				}

				rl.DrawTexturePro(
					info.Texture,
					info.Region,
					rl.Rectangle{
						X:      btn.X + (btn.Width-info.Region.Width*scale)/2,
						Y:      btn.Y + (btn.Height-info.Region.Height*scale)/2,
						Width:  info.Region.Width * scale,
						Height: info.Region.Height * scale,
					},
					rl.Vector2{}, 0, rl.White,
				)
			}
		} else {
			rl.DrawText(label, int32(btn.X+(btn.Width-float32(rl.MeasureText(label, 16)))/2),
				int32(btn.Y+btn.Height/2-8), 16, rl.Black)
		}

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), btn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			editor.editingDirection = dir
			editor.selectedFrames = make([]string, 1)
			if len(tex.Frames) > 0 {
				editor.frameCountStr = fmt.Sprintf("%d", len(tex.Frames))
				editor.animationTimeStr = fmt.Sprintf("%.1f", tex.AnimationTime)
				editor.selectedFrames = make([]string, len(tex.Frames))
				for i, frame := range tex.Frames {
					editor.selectedFrames[i] = frame.Name
				}
			} else {
				editor.frameCountStr = "1"
				editor.animationTimeStr = "0.5"
			}
		}
	}

	drawDirButton(upBtn, beam.DirUp, "Up")
	drawDirButton(leftBtn, beam.DirLeft, "Left")
	drawDirButton(downBtn, beam.DirDown, "Down")
	drawDirButton(rightBtn, beam.DirRight, "Right")

	// Animation settings for selected direction
	y += int(dirBtnSize)*2 + int(dirBtnSpacing)*2 + padding

	createNPCInput("Frame Count", &editor.frameCountStr, rightX, y, true)
	y += inputHeight + padding
	createNPCInput("Animation Time", &editor.animationTimeStr, rightX, y, true)
	y += inputHeight + padding

	// Frame selection grid
	frameCount, _ := strconv.Atoi(editor.frameCountStr)
	if frameCount > 0 {
		rl.DrawText("Animation Frames:", int32(rightX), int32(y), 16, rl.Black)
		y += 25

		frameSize := int32(50)
		framePadding := int32(5)
		framesPerRow := (columnWidth - padding*2) / (int(frameSize) + int(framePadding))

		// Adjust selected frames array size if needed
		if len(editor.selectedFrames) != frameCount {
			newFrames := make([]string, frameCount)
			copy(newFrames, editor.selectedFrames)
			editor.selectedFrames = newFrames
		}

		for i := 0; i < frameCount; i++ {
			row := i / framesPerRow
			col := i % framesPerRow

			frameX := rightX + col*(int(frameSize)+int(framePadding))
			frameY := y + row*(int(frameSize)+int(framePadding))

			frameRect := rl.Rectangle{
				X:      float32(frameX),
				Y:      float32(frameY),
				Width:  float32(frameSize),
				Height: float32(frameSize),
			}

			rl.DrawRectangleRec(frameRect, rl.LightGray)

			if i < len(editor.selectedFrames) && editor.selectedFrames[i] != "" {
				info, err := m.resources.GetTexture("default", editor.selectedFrames[i])
				if err != nil {
					fmt.Println("Error getting texture:", err)
					continue
				}
				scale := float32(frameSize-10) / info.Region.Width
				if info.Region.Height*scale > float32(frameSize-10) {
					scale = float32(frameSize-10) / info.Region.Height
				}

				rl.DrawTexturePro(
					info.Texture,
					info.Region,
					rl.Rectangle{
						X:      frameRect.X + (frameRect.Width-info.Region.Width*scale)/2,
						Y:      frameRect.Y + (frameRect.Height-info.Region.Height*scale)/2,
						Width:  info.Region.Width * scale,
						Height: info.Region.Height * scale,
					},
					rl.Vector2{}, 0, rl.White,
				)

				if rl.CheckCollisionPointRec(rl.GetMousePosition(), frameRect) {
					if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
						editor.advSelectingFrameIndex = i
						m.showResourceViewer = true
						m.uiState.resourceViewerOpenTime = rl.GetTime()
					} else if rl.IsMouseButtonPressed(rl.MouseRightButton) {
						editor.selectedFrameIndex = i

						// Initialize frame settings with current values
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

						if currentTex != nil && i < len(currentTex.Frames) {
							frame := currentTex.Frames[i]
							editor.frameRotation = fmt.Sprintf("%.1f", frame.Rotation)
							editor.frameScaleX = fmt.Sprintf("%.2f", frame.ScaleX)
							editor.frameScaleY = fmt.Sprintf("%.2f", frame.ScaleY)
							editor.frameOffsetX = fmt.Sprintf("%.2f", frame.OffsetX)
							editor.frameOffsetY = fmt.Sprintf("%.2f", frame.OffsetY)
							editor.frameMirrorX = frame.MirrorX
							editor.frameMirrorY = frame.MirrorY
							editor.frameTintR = fmt.Sprintf("%d", frame.Tint.R)
							editor.frameTintG = fmt.Sprintf("%d", frame.Tint.G)
							editor.frameTintB = fmt.Sprintf("%d", frame.Tint.B)
							editor.frameTintA = fmt.Sprintf("%d", frame.Tint.A)
						} else {
							// Set default values
							editor.frameRotation = "0.0"
							editor.frameScaleX = "1.0"
							editor.frameScaleY = "1.0"
							editor.frameOffsetX = "0.0"
							editor.frameOffsetY = "0.0"
							editor.frameMirrorX = false
							editor.frameMirrorY = false
							editor.frameTintR = "255"
							editor.frameTintG = "255"
							editor.frameTintB = "255"
							editor.frameTintA = "255"
						}
					}
				}

			} else {
				rl.DrawText("+", int32(frameRect.X+frameRect.Width/2-5),
					int32(frameRect.Y+frameRect.Height/2-8), 16, rl.DarkGray)
			}

			if rl.CheckCollisionPointRec(rl.GetMousePosition(), frameRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				m.uiState.npcEditor.advSelectingFrameIndex = i
				m.showResourceViewer = true
				m.uiState.resourceViewerOpenTime = rl.GetTime()
			}
		}
	}

	if editor.selectedFrameIndex >= 0 {
		m.renderNPCFrameSettings(editor, dialogX, dialogY, dialogWidth, dialogHeight)
	}

	// Save/Cancel buttons
	saveBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - 200),
		Y:      float32(dialogY + dialogHeight - 40),
		Width:  80,
		Height: 30,
	}

	cancelBtn := rl.Rectangle{
		X:      float32(dialogX + dialogWidth - 100),
		Y:      float32(dialogY + dialogHeight - 40),
		Width:  80,
		Height: 30,
	}

	rl.DrawRectangleRec(saveBtn, rl.Green)
	rl.DrawRectangleRec(cancelBtn, rl.Red)
	rl.DrawText("Save", int32(saveBtn.X+25), int32(saveBtn.Y+8), 16, rl.White)
	rl.DrawText("Cancel", int32(cancelBtn.X+20), int32(cancelBtn.Y+8), 16, rl.White)

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), cancelBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.closeNPCEditor()
		return
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), saveBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Validate and save NPC data
		health, _ := strconv.Atoi(editor.health)
		attack, _ := strconv.Atoi(editor.attack)
		defense, _ := strconv.Atoi(editor.defense)
		attackSpeed, _ := strconv.ParseFloat(editor.attackSpeed, 64)
		attackRange, _ := strconv.ParseFloat(editor.attackRange, 64)
		moveSpeed, _ := strconv.ParseFloat(editor.moveSpeed, 64)
		aggroRange, _ := strconv.Atoi(editor.aggroRange)
		spawnX, _ := strconv.Atoi(editor.spawnXStr)
		spawnY, _ := strconv.Atoi(editor.spawnYStr)
		wanderRange, _ := strconv.Atoi(editor.wanderRange)

		// Create NPC data
		npcData := beam.NPCData{
			Name:            editor.name,
			Texture:         editor.textures,
			Health:          health,
			MaxHealth:       health,
			Attack:          attack,
			BaseAttack:      attack,
			Defense:         defense,
			BaseDefense:     defense,
			AttackSpeed:     attackSpeed,
			BaseAttackSpeed: attackSpeed,
			AttackRange:     attackRange,
			BaseAttackRange: attackRange,
			MoveSpeed:       moveSpeed,
			Direction:       beam.DirDown,
			Hostile:         editor.isHostile,
			AggroRange:      aggroRange,
			Attackable:      editor.attackable,
			Impassable:      editor.impassable,
			WanderRange:     wanderRange,
			SpawnPos:        beam.Position{X: spawnX, Y: spawnY}, // Set SpawnPos
		}

		// Validate all inputs
		if editor.name == "" || editor.health == "" || editor.attack == "" ||
			editor.defense == "" || editor.attackSpeed == "" || editor.attackRange == "" ||
			editor.moveSpeed == "" || editor.aggroRange == "" || editor.spawnXStr == "" || editor.spawnYStr == "" { // Added spawn pos validation
			rl.DrawText("Please fill in all fields.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		if health <= 0 || attack <= 0 || defense < 0 || attackSpeed <= 0 ||
			attackRange <= 0 || moveSpeed <= 0 || aggroRange < 0 || spawnX < 0 || spawnY < 0 { // Added spawn pos validation
			rl.DrawText("Values must be positive (except Defense/Aggro/Spawn).", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		if _, err := strconv.Atoi(editor.aggroRange); err != nil {
			rl.DrawText("Aggro Range must be an integer.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		if _, err := strconv.Atoi(editor.health); err != nil {
			rl.DrawText("Health must be an integer.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		if _, err := strconv.Atoi(editor.attack); err != nil {
			rl.DrawText("Attack must be an integer.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		if _, err := strconv.Atoi(editor.defense); err != nil {
			rl.DrawText("Defense must be an integer.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		if _, err := strconv.Atoi(editor.spawnXStr); err != nil { // Added spawn X validation
			rl.DrawText("Spawn X must be an integer.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		if _, err := strconv.Atoi(editor.spawnYStr); err != nil { // Added spawn Y validation
			rl.DrawText("Spawn Y must be an integer.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}
		// texture for every direction
		if len(editor.textures.Up.Frames) == 0 || len(editor.textures.Down.Frames) == 0 ||
			len(editor.textures.Left.Frames) == 0 || len(editor.textures.Right.Frames) == 0 {
			rl.DrawText("Please select textures for all directions.", int32(dialogX+20), int32(dialogY+dialogHeight-80), 16, rl.Red)
			return
		}

		// Save NPC data to the tile
		found := false
		for i, npc := range m.tileGrid.Map.NPCs {
			if npc.Data.Name == editor.name {
				m.tileGrid.Map.NPCs[i] = &beam.NPC{
					Data: npcData,
					Pos:  npcData.SpawnPos,
				}
				found = true
				break
			}
		}

		if !found {
			// Add new NPC to the map
			m.tileGrid.Map.NPCs = append(m.tileGrid.Map.NPCs, &beam.NPC{
				Data: npcData,
				Pos:  npcData.SpawnPos,
			})
		}
		m.closeNPCEditor()
	}
}

// renderNPCList renders the NPC list view
func (m *MapMaker) renderNPCList() {
	dialogWidth := 600
	dialogHeight := 400
	dialogX := (rl.GetScreenWidth() - dialogWidth) / 2
	dialogY := (rl.GetScreenHeight() - dialogHeight) / 2

	// Draw semi-transparent background
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.Black, 0.7))

	// Draw dialog background
	rl.DrawRectangle(int32(dialogX), int32(dialogY), int32(dialogWidth), int32(dialogHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(dialogX),
		Y:      float32(dialogY),
		Width:  float32(dialogWidth),
		Height: float32(dialogHeight),
	}, 1, rl.Gray)

	// Draw title
	rl.DrawText("NPC List", int32(dialogX+20), int32(dialogY+20), 24, rl.Black)

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
		m.uiState.showNPCList = false
	}

	// List content area
	contentY := dialogY + 60
	rowHeight := int32(40)
	padding := int32(10)

	// Draw headers
	rl.DrawText("Name", int32(dialogX+20), int32(contentY), 20, rl.DarkGray)
	rl.DrawText("Position", int32(dialogX+200), int32(contentY), 20, rl.DarkGray)
	rl.DrawText("Actions", int32(dialogX+400), int32(contentY), 20, rl.DarkGray)
	contentY += 30

	// Draw NPC rows
	for i, npc := range m.tileGrid.NPCs {
		y := int32(contentY) + int32(i)*rowHeight
		rowBg := rl.White
		if i%2 == 0 {
			rowBg = rl.LightGray
		}

		// Draw row background
		rl.DrawRectangle(
			int32(dialogX)+10,
			int32(y),
			int32(dialogWidth)-20,
			rowHeight-2,
			rowBg,
		)

		// Draw NPC info
		rl.DrawText(npc.Data.Name, int32(dialogX+20), int32(y+10), 16, rl.Black)
		rl.DrawText(fmt.Sprintf("(%d, %d)", npc.Pos.X, npc.Pos.Y), int32(dialogX+200), int32(y+10), 16, rl.Black)

		// Edit button
		editBtn := rl.Rectangle{
			X:      float32(dialogX + 400),
			Y:      float32(y + padding/2),
			Width:  60,
			Height: float32(rowHeight - padding),
		}
		rl.DrawRectangleRec(editBtn, rl.Blue)
		rl.DrawText("Edit", int32(editBtn.X+15), int32(editBtn.Y+5), 16, rl.White)

		// Delete button
		deleteBtn := rl.Rectangle{
			X:      float32(dialogX + 470),
			Y:      float32(y + padding/2),
			Width:  60,
			Height: float32(rowHeight - padding),
		}
		rl.DrawRectangleRec(deleteBtn, rl.Red)
		rl.DrawText("Delete", int32(deleteBtn.X+5), int32(deleteBtn.Y+5), 16, rl.White)

		// Handle button clicks
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), editBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.uiState.npcEditor = &NPCEditorState{
				visible:          true,
				spawnPos:         npc.Data.SpawnPos,
				name:             npc.Data.Name,
				health:           strconv.Itoa(npc.Data.Health),
				attack:           strconv.Itoa(npc.Data.Attack),
				defense:          strconv.Itoa(npc.Data.Defense),
				attackSpeed:      fmt.Sprintf("%.1f", npc.Data.AttackSpeed),
				attackRange:      fmt.Sprintf("%.1f", npc.Data.AttackRange),
				moveSpeed:        fmt.Sprintf("%.1f", npc.Data.MoveSpeed),
				aggroRange:       strconv.Itoa(npc.Data.AggroRange),
				isHostile:        npc.Data.Hostile,
				textures:         npc.Data.Texture,
				editingDirection: beam.DirDown,
				frameCountStr:    "1",
				animationTimeStr: "0.5",
				selectedFrames:   make([]string, 1),
				spawnXStr:        strconv.Itoa(npc.Data.SpawnPos.X), // Initialize spawnXStr
				spawnYStr:        strconv.Itoa(npc.Data.SpawnPos.Y), // Initialize spawnYStr
				attackable:       npc.Data.Attackable,
				impassable:       npc.Data.Impassable,
				wanderRange:      strconv.Itoa(npc.Data.WanderRange),
			}
			m.uiState.showNPCList = false
			m.uiState.npcEditor.selectedFrameIndex = -1
		}

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), deleteBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			// Remove the NPC
			m.tileGrid.NPCs = append(m.tileGrid.NPCs[:i], m.tileGrid.NPCs[i+1:]...)
		}
	}

	if len(m.tileGrid.NPCs) == 0 {
		rl.DrawText("No NPCs placed on the map", int32(dialogX+200), int32(contentY+20), 20, rl.Gray)
	}
}

func (m *MapMaker) renderNPCFrameSettings(editor *NPCEditorState, dialogX, dialogY, dialogWidth, dialogHeight int) {
	if editor.selectedFrameIndex < 0 || editor.selectedFrameIndex >= len(editor.selectedFrames) {
		return
	}

	// Get the current frame's texture
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

	if currentTex == nil || editor.selectedFrameIndex >= len(currentTex.Frames) {
		return
	}

	// Settings panel position and dimensions
	settingsPanelWidth := 250
	settingsPanelHeight := 400
	settingsX := dialogX + dialogWidth - settingsPanelWidth - 20
	settingsY := dialogY + 50

	// Draw settings panel background
	rl.DrawRectangle(int32(settingsX), int32(settingsY),
		int32(settingsPanelWidth), int32(settingsPanelHeight), rl.RayWhite)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      float32(settingsX),
		Y:      float32(settingsY),
		Width:  float32(settingsPanelWidth),
		Height: float32(settingsPanelHeight),
	}, 1, rl.Gray)

	// Title
	rl.DrawText(fmt.Sprintf("Frame %d Settings", editor.selectedFrameIndex+1),
		int32(settingsX+10), int32(settingsY+10), 16, rl.Black)

	// Helper function for input fields
	createFrameInput := func(label string, value *string, y int, numeric bool) {
		inputRect := rl.Rectangle{
			X:      float32(settingsX + 100),
			Y:      float32(y),
			Width:  100,
			Height: 25,
		}
		rl.DrawText(label, int32(settingsX+10), int32(y+5), 14, rl.Black)
		rl.DrawRectangleRec(inputRect, rl.LightGray)
		rl.DrawText(*value, int32(inputRect.X+5), int32(inputRect.Y+5), 14, rl.Black)

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), inputRect) &&
			rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			m.uiState.activeNPCInput = "frame_" + label
		}

		if m.uiState.activeNPCInput == "frame_"+label {
			rl.DrawRectangleLinesEx(inputRect, 2, rl.Blue)
			key := rl.GetCharPressed()
			for key > 0 {
				if numeric {
					if (key >= '0' && key <= '9') || key == '.' || key == '-' {
						*value += string(key)
					}
				} else {
					if key >= 32 && key <= 126 {
						*value += string(key)
					}
				}
				key = rl.GetCharPressed()
			}
			if rl.IsKeyPressed(rl.KeyBackspace) && len(*value) > 0 {
				*value = (*value)[:len(*value)-1]
			}
		}
	}

	// Input fields
	y := settingsY + 40
	spacing := 30
	createFrameInput("Rotation", &editor.frameRotation, y, true)
	y += spacing
	createFrameInput("Scale X", &editor.frameScaleX, y, true)
	y += spacing
	createFrameInput("Scale Y", &editor.frameScaleY, y, true)
	y += spacing
	createFrameInput("Offset X", &editor.frameOffsetX, y, true)
	y += spacing
	createFrameInput("Offset Y", &editor.frameOffsetY, y, true)
	y += spacing

	// Mirror toggles
	checkboxSize := int32(20)
	mirrorXBox := rl.Rectangle{
		X:      float32(settingsX + 100),
		Y:      float32(y),
		Width:  float32(checkboxSize),
		Height: float32(checkboxSize),
	}
	mirrorYBox := rl.Rectangle{
		X:      float32(settingsX + 180),
		Y:      float32(y),
		Width:  float32(checkboxSize),
		Height: float32(checkboxSize),
	}

	rl.DrawText("Mirror:", int32(settingsX+10), int32(y+2), 14, rl.Black)
	rl.DrawRectangleRec(mirrorXBox, rl.LightGray)
	rl.DrawRectangleRec(mirrorYBox, rl.LightGray)
	rl.DrawText("X", int32(mirrorXBox.X-15), int32(mirrorXBox.Y+2), 14, rl.Black)
	rl.DrawText("Y", int32(mirrorYBox.X-15), int32(mirrorYBox.Y+2), 14, rl.Black)

	if editor.frameMirrorX {
		rl.DrawRectangle(int32(mirrorXBox.X+4), int32(mirrorXBox.Y+4),
			int32(mirrorXBox.Width-8), int32(mirrorXBox.Height-8), rl.Black)
	}
	if editor.frameMirrorY {
		rl.DrawRectangle(int32(mirrorYBox.X+4), int32(mirrorYBox.Y+4),
			int32(mirrorYBox.Width-8), int32(mirrorYBox.Height-8), rl.Black)
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), mirrorXBox) &&
		rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		editor.frameMirrorX = !editor.frameMirrorX
	}
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), mirrorYBox) &&
		rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		editor.frameMirrorY = !editor.frameMirrorY
	}

	y += spacing + 10

	// Tint inputs
	rl.DrawText("Tint:", int32(settingsX+10), int32(y+2), 14, rl.Black)
	createFrameInput("R", &editor.frameTintR, y, true)
	createFrameInput("G", &editor.frameTintG, y+spacing, true)
	createFrameInput("B", &editor.frameTintB, y+spacing*2, true)
	createFrameInput("A", &editor.frameTintA, y+spacing*3, true)

	// Apply button
	applyBtn := rl.Rectangle{
		X:      float32(settingsX + settingsPanelWidth - 70),
		Y:      float32(settingsY + settingsPanelHeight - 35),
		Width:  60,
		Height: 25,
	}
	rl.DrawRectangleRec(applyBtn, rl.Green)
	rl.DrawText("Apply", int32(applyBtn.X+10), int32(applyBtn.Y+5), 14, rl.White)

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), applyBtn) &&
		rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Apply the changes to the current frame
		rotation, _ := strconv.ParseFloat(editor.frameRotation, 64)
		scaleX, _ := strconv.ParseFloat(editor.frameScaleX, 64)
		scaleY, _ := strconv.ParseFloat(editor.frameScaleY, 64)
		offsetX, _ := strconv.ParseFloat(editor.frameOffsetX, 64)
		offsetY, _ := strconv.ParseFloat(editor.frameOffsetY, 64)
		tintR, _ := strconv.Atoi(editor.frameTintR)
		tintG, _ := strconv.Atoi(editor.frameTintG)
		tintB, _ := strconv.Atoi(editor.frameTintB)
		tintA, _ := strconv.Atoi(editor.frameTintA)

		frame := &currentTex.Frames[editor.selectedFrameIndex]
		frame.Rotation = rotation
		frame.ScaleX = scaleX
		frame.ScaleY = scaleY
		frame.OffsetX = offsetX
		frame.OffsetY = offsetY
		frame.MirrorX = editor.frameMirrorX
		frame.MirrorY = editor.frameMirrorY
		frame.Tint = rl.Color{
			R: uint8(tintR),
			G: uint8(tintG),
			B: uint8(tintB),
			A: uint8(tintA),
		}

		editor.selectedFrameIndex = -1
	}
}

type TextureEditorState struct {
	tile          *beam.Tile
	visible       bool
	texIndex      int
	frameIndex    int
	rotation      string
	scalex        string
	scaley        string
	offsetX       string
	offsetY       string
	tintR         string
	tintG         string
	tintB         string
	tintA         string
	mirrorX       bool
	mirrorY       bool
	clearedInputs map[string]bool
	layer         beam.Layer

	// Advanced Editor State
	advAnimationTimeStr    string
	advFrameCountStr       string
	advSelectedFrames      []string // Stores texture names for each frame
	advSelectingFrameIndex int      // Index of the frame being selected via resource viewer, -1 if none
	selectedFrameIndex     int
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
	dialogHeight := 480
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
	padding := 10
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
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), inputRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) &&
			m.uiState.activeInput != "layer_dropdown" {
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

	// Helper function to create boolean input field
	createBoolInput := func(label string, value *bool, yPos int) {
		rl.DrawText(label, int32(dialogX+padding), int32(yPos+8), 16, rl.Black)
		checkboxRect := rl.Rectangle{
			X:      float32(dialogX + padding + labelWidth),
			Y:      float32(yPos),
			Width:  float32(inputHeight),
			Height: float32(inputHeight),
		}
		rl.DrawRectangleRec(checkboxRect, rl.LightGray)
		if *value {
			rl.DrawRectangle(
				int32(checkboxRect.X+5),
				int32(checkboxRect.Y+5),
				int32(checkboxRect.Width-10),
				int32(checkboxRect.Height-10),
				rl.Black,
			)
		}

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), checkboxRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			*value = !*value
		}
	}

	// Set the layer dropdown
	layerY := y
	y += inputHeight + padding

	// Create all input fields
	createInput("Rotation", &editor.rotation, y)
	y += inputHeight + padding
	createInput("Scale X", &editor.scalex, y)
	y += inputHeight + padding
	createInput("Scale Y", &editor.scaley, y)
	y += inputHeight + padding
	createInput("Offset X", &editor.offsetX, y)
	y += inputHeight + padding
	createInput("Offset Y", &editor.offsetY, y)
	y += inputHeight + padding
	createBoolInput("Mirror X", &editor.mirrorX, y)
	y += inputHeight + padding
	createBoolInput("Mirror Y", &editor.mirrorY, y)
	y += inputHeight + padding

	// Draw layer dropdown
	m.renderLayerDropdown(dialogX, padding, layerY, labelWidth, inputWidth+55, inputHeight, editor)

	// Continue with tint inputs
	tintLabel := "Tint RGBA:"
	rl.DrawText(tintLabel, int32(dialogX+padding)-5, int32(y+8), 16, rl.Black)

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
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), cancelBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) && !m.uiState.showAdvancedEditor {
		m.closeTextureEditor()
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), saveBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Update all selected tiles with new values
		for _, pos := range m.uiState.tileInfoPos {
			tile := &m.tileGrid.Tiles[pos.Y][pos.X]

			if editor.texIndex < len(tile.Textures) && editor.frameIndex < len(tile.Textures[editor.texIndex].Frames) {
				currTexture := tile.Textures[editor.texIndex]
				currTexture.Layer = editor.layer
				frame := &currTexture.Frames[editor.frameIndex]
				frame.Rotation, _ = strconv.ParseFloat(editor.rotation, 64)
				frame.ScaleX, _ = strconv.ParseFloat(editor.scalex, 64)
				frame.ScaleY, _ = strconv.ParseFloat(editor.scaley, 64)
				frame.OffsetX, _ = strconv.ParseFloat(editor.offsetX, 64)
				frame.OffsetY, _ = strconv.ParseFloat(editor.offsetY, 64)
				frame.MirrorX = editor.mirrorX
				frame.MirrorY = editor.mirrorY
				r, _ := strconv.Atoi(editor.tintR)
				g, _ := strconv.Atoi(editor.tintG)
				b, _ := strconv.Atoi(editor.tintB)
				a, _ := strconv.Atoi(editor.tintA)
				frame.Tint = rl.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
			}
		}
		m.closeTextureEditor()
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), advancedBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Initialize advanced editor state when opening it
		tex := editor.tile.Textures[editor.texIndex]
		if tex.IsAnimated && len(tex.Frames) > 0 {
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
			editor.selectedFrameIndex = -1               // Initialize to no selection
		}
		editor.advSelectingFrameIndex = -1
		m.uiState.showAdvancedEditor = true
		m.uiState.advancedEditorOpenTime = rl.GetTime()
		m.uiState.activeInput = ""
	}
	if m.uiState.showAdvancedEditor {
		m.renderAdvancedEditor()
	}
}

func (m *MapMaker) renderLayerDropdown(dialogX int, padding int, y int, labelWidth int, inputWidth int, inputHeight int, editor *TextureEditorState) {
	rl.DrawText("Layer", int32(dialogX+padding), int32(y+8), 16, rl.Black)
	dropdownRect := rl.Rectangle{
		X:      float32(dialogX + padding + labelWidth),
		Y:      float32(y),
		Width:  float32(inputWidth),
		Height: float32(inputHeight),
	}

	// Draw dropdown button
	rl.DrawRectangleRec(dropdownRect, rl.LightGray)
	rl.DrawRectangleLinesEx(dropdownRect, 1, rl.Gray)
	layerText := editor.layer.String()
	if layerText == "" {
		layerText = "Background" // Default value
	}
	rl.DrawText(layerText, int32(dropdownRect.X+5), int32(dropdownRect.Y+8), 16, rl.Black)

	// Draw dropdown arrow
	arrowSize := int32(8)
	rl.DrawTriangle(
		rl.Vector2{X: float32(dropdownRect.X + dropdownRect.Width - 15), Y: float32(dropdownRect.Y + 12)},
		rl.Vector2{X: float32(dropdownRect.X + dropdownRect.Width - 15 - float32(arrowSize)), Y: float32(dropdownRect.Y + 12)},
		rl.Vector2{X: float32(dropdownRect.X + dropdownRect.Width - 15 - float32(arrowSize/2)), Y: float32(dropdownRect.Y + 12 + float32(arrowSize))},
		rl.DarkGray,
	)

	// Handle dropdown click
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), dropdownRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		m.uiState.activeInput = "layer_dropdown"
	}

	// Show dropdown list if active
	if m.uiState.activeInput == "layer_dropdown" {
		layers := []beam.Layer{beam.BaseLayer, beam.BackgroundLayer, beam.ForegroundLayer}
		listRect := rl.Rectangle{
			X:      dropdownRect.X,
			Y:      dropdownRect.Y + dropdownRect.Height,
			Width:  dropdownRect.Width,
			Height: float32(len(layers) * inputHeight),
		}

		rl.DrawRectangleRec(listRect, rl.White)
		rl.DrawRectangleLinesEx(listRect, 1, rl.Gray)

		for i, layer := range layers {
			itemRect := rl.Rectangle{
				X:      listRect.X,
				Y:      listRect.Y + float32(i*inputHeight),
				Width:  listRect.Width,
				Height: float32(inputHeight),
			}

			// Highlight on hover
			if rl.CheckCollisionPointRec(rl.GetMousePosition(), itemRect) {
				rl.DrawRectangleRec(itemRect, rl.LightGray)
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					editor.layer = layer
					m.uiState.activeInput = ""
				}
			}

			rl.DrawText(layer.String(), int32(itemRect.X+5), int32(itemRect.Y+8), 16, rl.Black)
		}
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

	timeSinceOpen := rl.GetTime() - m.uiState.advancedEditorOpenTime
	canAcceptClicks := timeSinceOpen >= 0.5

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

	// Adjust selected frames slice size if frameCount changed
	if len(editor.advSelectedFrames) != frameCount && frameCount >= 0 {
		newFrames := make([]string, frameCount)
		copy(newFrames, editor.advSelectedFrames)
		editor.advSelectedFrames = newFrames
		if editor.selectedFrameIndex >= frameCount {
			editor.selectedFrameIndex = -1
		}
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
		if i == editor.selectedFrameIndex {
			rl.DrawRectangleLinesEx(frameRect, 2, rl.Blue)
		} else {
			rl.DrawRectangleLinesEx(frameRect, 1, rl.Gray)
		}

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
			if canAcceptClicks && rl.CheckCollisionPointRec(rl.GetMousePosition(), frameRect) && rl.IsMouseButtonPressed(rl.MouseRightButton) {
				editor.selectedFrameIndex = i
			}
		} else {
			// Draw placeholder for empty slot
			rl.DrawText("+", int32(int(frameRect.X)+framePreviewSize/2-5), int32(int(frameRect.Y)+framePreviewSize/2-10), 20, rl.DarkGray)
		}

		// Handle click to select texture for this frame
		if canAcceptClicks && rl.CheckCollisionPointRec(rl.GetMousePosition(), frameRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			editor.advSelectingFrameIndex = i
			m.showResourceViewer = true
			m.uiState.resourceViewerOpenTime = rl.GetTime()
		}
	}

	// Display frame settings when a frame is selected
	if editor.selectedFrameIndex >= 0 && editor.selectedFrameIndex < len(editor.advSelectedFrames) {
		settingsY := contentY + ((frameCount+framesPerRow-1)/framesPerRow)*(framePreviewSize+framePadding) + padding*2

		// Frame settings title
		rl.DrawText(fmt.Sprintf("Frame %d Settings:", editor.selectedFrameIndex+1),
			int32(dialogX+padding), int32(settingsY), 16, rl.Black)
		settingsY += 25

		// Get original frame settings if they exist
		var currFrame *beam.Texture
		if editor.texIndex < len(editor.tile.Textures) {
			tex := editor.tile.Textures[editor.texIndex]
			if editor.selectedFrameIndex < len(tex.Frames) {
				currFrame = &tex.Frames[editor.selectedFrameIndex]
			}
		}

		// Settings grid layout
		settingsPerRow := 2
		settingWidth := (dialogWidth - padding*3) / settingsPerRow
		settingHeight := 25

		// Helper to draw setting value
		drawSetting := func(label string, value string, col, row int) {
			x := dialogX + padding + col*settingWidth
			y := int32(settingsY) + int32(row*settingHeight)
			rl.DrawText(fmt.Sprintf("%s: %s", label, value), int32(x), y, 14, rl.DarkGray)
		}

		// Draw current settings
		if currFrame != nil {
			drawSetting("Rotation", fmt.Sprintf("%.1f°", currFrame.Rotation), 1, 3)
			drawSetting("Scale X", fmt.Sprintf("%.2f", currFrame.ScaleX), 0, 0)
			drawSetting("Scale Y", fmt.Sprintf("%.2f", currFrame.ScaleY), 1, 0)
			drawSetting("Offset X", fmt.Sprintf("%.2f", currFrame.OffsetX), 0, 1)
			drawSetting("Offset Y", fmt.Sprintf("%.2f", currFrame.OffsetY), 1, 1)
			drawSetting("Mirror X", fmt.Sprintf("%v", currFrame.MirrorX), 0, 2)
			drawSetting("Mirror Y", fmt.Sprintf("%v", currFrame.MirrorY), 1, 2)
			drawSetting("Tint", fmt.Sprintf("R:%d G:%d B:%d A:%d",
				currFrame.Tint.R, currFrame.Tint.G, currFrame.Tint.B, currFrame.Tint.A), 0, 3)
		} else {
			// Show default values for new frames
			drawSetting("Rotation", "0.0°", 1, 3)
			drawSetting("Scale X", "1.00", 0, 0)
			drawSetting("Scale Y", "1.00", 1, 0)
			drawSetting("Offset X", "0.00", 0, 1)
			drawSetting("Offset Y", "0.00", 1, 1)
			drawSetting("Mirror X", "false", 0, 2)
			drawSetting("Mirror Y", "false", 1, 2)
			drawSetting("Tint", "R:255 G:255 B:255 A:255", 0, 3)
		}

		// Add an edit button
		editBtn := rl.Rectangle{
			X:      float32(dialogX + dialogWidth - 100),
			Y:      float32(settingsY),
			Width:  90,
			Height: 25,
		}
		rl.DrawRectangleRec(editBtn, rl.Blue)
		rl.DrawText("Edit Frame", int32(editBtn.X+10), int32(editBtn.Y+5), 14, rl.White)

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), editBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			// Initialize simple editor with current frame values
			if currFrame != nil {
				editor.rotation = fmt.Sprintf("%.1f", currFrame.Rotation)
				editor.scalex = fmt.Sprintf("%.2f", currFrame.ScaleX)
				editor.scaley = fmt.Sprintf("%.2f", currFrame.ScaleY)
				editor.offsetX = fmt.Sprintf("%.2f", currFrame.OffsetX)
				editor.offsetY = fmt.Sprintf("%.2f", currFrame.OffsetY)
				editor.mirrorX = currFrame.MirrorX
				editor.mirrorY = currFrame.MirrorY
				editor.tintR = fmt.Sprintf("%d", currFrame.Tint.R)
				editor.tintG = fmt.Sprintf("%d", currFrame.Tint.G)
				editor.tintB = fmt.Sprintf("%d", currFrame.Tint.B)
				editor.tintA = fmt.Sprintf("%d", currFrame.Tint.A)
			} else {
				// Set default values
				editor.rotation = "0.0"
				editor.scalex = "1.0"
				editor.scaley = "1.0"
				editor.offsetX = "0.0"
				editor.offsetY = "0.0"
				editor.tintR = "255"
				editor.tintG = "255"
				editor.tintB = "255"
				editor.tintA = "255"
				editor.mirrorX = false
				editor.mirrorY = false
			}
			editor.frameIndex = editor.selectedFrameIndex
			m.uiState.showAdvancedEditor = false
			editor.visible = true
		}
	}

	// Save/Cancel buttons
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

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), exitBtn) && rl.IsMouseButtonPressed(rl.MouseLeftButton) && !m.showResourceViewer {
		m.closeAllEditors()
		m.uiState.activeInput = ""
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), saveBtnAdv) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Validate inputs
		animTime := 0.0
		var timeErr error
		if frameCount > 0 {
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
			// Apply changes to all selected tiles
			for _, pos := range m.uiState.tileInfoPos {
				tile := &m.tileGrid.Tiles[pos.Y][pos.X]
				if editor.texIndex < len(tile.Textures) {
					tex := tile.Textures[editor.texIndex]
					// Use properties from the *original* first frame if available, otherwise defaults
					originalRotation := 0.0
					originalScale := 1.0
					originalOffsetX := 0.0
					originalOffsetY := 0.0
					originalTint := rl.White
					originalMirrorX := false
					originalMirrorY := false
					if len(tile.Textures[editor.texIndex].Frames) > 0 {
						originalFrame := tile.Textures[editor.texIndex].Frames[0]
						originalRotation = originalFrame.Rotation
						originalScale = originalFrame.ScaleX
						originalOffsetX = originalFrame.OffsetX
						originalOffsetY = originalFrame.OffsetY
						originalTint = originalFrame.Tint
						originalMirrorY = originalFrame.MirrorY
						originalMirrorX = originalFrame.MirrorX
					}

					tex.IsAnimated = frameCount > 1
					tex.AnimationTime = animTime
					tex.CurrentFrame = 0
					tex.Frames = make([]beam.Texture, 0, frameCount)

					for i := 0; i < frameCount; i++ {
						newFrame := beam.Texture{
							Name:     editor.advSelectedFrames[i],
							Rotation: originalRotation,
							ScaleX:   originalScale,
							ScaleY:   originalScale,
							OffsetX:  originalOffsetX,
							OffsetY:  originalOffsetY,
							Tint:     originalTint,
							MirrorX:  originalMirrorX,
							MirrorY:  originalMirrorY,
						}
						tex.Frames = append(tex.Frames, newFrame)
					}
				}
			}

			m.showToast("Texture properties saved!", ToastSuccess)
			m.uiState.showAdvancedEditor = false
			m.uiState.textureEditor = nil
			m.uiState.activeInput = ""
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
	m.showResourceViewer = false // Close the resource viewer if it was open
}

func (m *MapMaker) closeAllEditors() {
	m.closeTextureEditor()
	m.showTileInfo = false
	m.uiState.activeInput = ""
}

func (m *MapMaker) closeNPCEditor() {
	m.uiState.npcEditor = nil
	m.showResourceViewer = false // Close the resource viewer if it was open
}
