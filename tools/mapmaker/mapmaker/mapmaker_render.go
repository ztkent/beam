package mapmaker

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam"
)

func (m *MapMaker) renderGrid() {
	startX := m.tileGrid.offset.X
	startY := m.tileGrid.offset.Y

	// Draw grid lines
	for i := 0; i <= m.tileGrid.Width; i++ {
		x := startX + i*m.uiState.tileSize
		rl.DrawLine(int32(x), int32(startY), int32(x), int32(startY+m.tileGrid.Height*m.uiState.tileSize), rl.LightGray)
	}
	for i := 0; i <= m.tileGrid.Height; i++ {
		y := startY + i*m.uiState.tileSize
		rl.DrawLine(int32(startX), int32(y), int32(startX+m.tileGrid.Width*m.uiState.tileSize), int32(y), rl.LightGray)
	}

	// Draw grid tiles
	for y := 0; y < m.tileGrid.Height; y++ {
		for x := 0; x < m.tileGrid.Width; x++ {
			pos := rl.Rectangle{
				X:      float32(startX + x*m.uiState.tileSize),
				Y:      float32(startY + y*m.uiState.tileSize),
				Width:  float32(m.uiState.tileSize),
				Height: float32(m.uiState.tileSize),
			}

			// Draw all textures at this location, in order
			for i, textureName := range m.tileGrid.Textures[y][x] {
				rotateFloor := float64(m.tileGrid.TextureRotations[y][x][i])
				tile := beam.Position{X: x, Y: y}
				m.renderGridTile(pos, tile, textureName, rotateFloor)
			}
		}
	}

	// Draw grid lines over tiles
	for i := 0; i <= m.tileGrid.Width; i++ {
		x := startX + i*m.uiState.tileSize
		rl.DrawLine(int32(x), int32(startY), int32(x), int32(startY+m.tileGrid.Height*m.uiState.tileSize), rl.LightGray)
	}
	for i := 0; i <= m.tileGrid.Height; i++ {
		y := startY + i*m.uiState.tileSize
		rl.DrawLine(int32(startX), int32(y), int32(startX+m.tileGrid.Width*m.uiState.tileSize), int32(y), rl.LightGray)
	}

	// Draw selection highlight if there's a selection
	if m.tileGrid.hasSelection {
		for _, tile := range m.tileGrid.selectedTiles {
			highlightX := startX + tile.X*m.uiState.tileSize
			highlightY := startY + tile.Y*m.uiState.tileSize
			// Draw highlight rectangle with thicker lines
			rl.DrawRectangleLinesEx(rl.Rectangle{
				X:      float32(highlightX),
				Y:      float32(highlightY),
				Width:  float32(m.uiState.tileSize),
				Height: float32(m.uiState.tileSize),
			}, 2, rl.Black)
		}
	}

	// Draw grid dimensions in bottom right
	dimensions := fmt.Sprintf("%dx%d", m.tileGrid.Width, m.tileGrid.Height)
	textWidth := int(rl.MeasureText(dimensions, 20))
	textX := startX + m.tileGrid.Width*m.uiState.tileSize - textWidth
	textY := startY + m.tileGrid.Height*m.uiState.tileSize + 5
	rl.DrawText(dimensions, int32(textX), int32(textY), 20, rl.DarkGray)
}

func (m *MapMaker) renderGridTile(pos rl.Rectangle, tile beam.Position, textureName string, rotate float64) {
	if textureName == "" {
		return
	} else if m.tileGrid.missingResourceTiles.Contains(tile, textureName) {
		// Draw yellow outline for missing resource
		rl.DrawRectangleLinesEx(pos, 2, rl.Yellow)
		return
	}

	// Center the texture in the tile
	origin := rl.Vector2{
		X: float32(m.uiState.tileSize) / 2,
		Y: float32(m.uiState.tileSize) / 2,
	}

	info, err := m.resources.GetTexture("default", textureName)
	if err != nil {
		fmt.Println("Error getting texture:", err)
		return
	}

	// Adjust destination rectangle to use center-based rotation
	destRect := rl.Rectangle{
		X:      pos.X + pos.Width/2,
		Y:      pos.Y + pos.Height/2,
		Width:  pos.Width,
		Height: pos.Height,
	}

	rl.DrawTexturePro(
		info.Texture,
		info.Region,
		destRect,
		origin,
		float32(rotate),
		rl.White,
	)
}

func (m *MapMaker) renderUI() {
	// Draw header background
	rl.DrawRectangle(0, 0, m.window.width, int32(m.uiState.menuBarHeight), rl.RayWhite)
	rl.DrawLine(0, int32(m.uiState.menuBarHeight-1), m.window.width, int32(m.uiState.menuBarHeight-1), rl.LightGray)

	// Draw section dividers
	rl.DrawLine(150, 5, 150, int32(m.uiState.menuBarHeight-5), rl.LightGray)
	rl.DrawLine(m.window.width-180, 5, m.window.width-180, int32(m.uiState.menuBarHeight-5), rl.LightGray)

	// Get all buttons
	tileSmallerBtn, tileLargerBtn, resolutionBtn, loadBtn, saveBtn, loadResourceBtn, viewResourcesBtn, resetBtn, paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn := m.getUIButtons()

	// Draw buttons
	m.drawButton(tileSmallerBtn, rl.White)
	m.drawButton(tileLargerBtn, rl.White)
	rl.DrawText(fmt.Sprintf("%dpx", m.uiState.tileSize), 48, 12, 12, rl.DarkGray)
	m.drawButton(resolutionBtn, rl.White)

	// Draw new grid control buttons
	m.drawToolIcons(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn)

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

func (m *MapMaker) drawToolIcons(paintbrushBtn, paintbucketBtn, eraseBtn, selectBtn IconButton) {
	m.drawIconButton(paintbrushBtn, rl.LightGray)
	m.drawIconButton(paintbucketBtn, rl.LightGray)
	m.drawIconButton(eraseBtn, rl.LightGray)
	m.drawIconButton(selectBtn, rl.LightGray)

	// Draw tools with selection highlight
	toolButtons := map[string]IconButton{
		"paintbrush":   paintbrushBtn,
		"paintbucket":  paintbucketBtn,
		"eraser":       eraseBtn,
		"pencileraser": eraseBtn,
		"select":       selectBtn,
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
		X:      float32(dialogX + dialogWidth - 120),
		Y:      float32(dialogY + 10),
		Width:  70,
		Height: 30,
	}
	rl.DrawRectangleRec(manageBtn, rl.LightGray)
	rl.DrawText("Manage", int32(manageBtn.X+10), int32(manageBtn.Y+8), 16, rl.Black)

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
			rl.DrawText("Delete", int32(deleteBtn.X+6), int32(deleteBtn.Y+6), 14, rl.White)
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
	dialogWidth := 300
	dialogHeight := 200

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
	tileType := m.tileGrid.Tiles[pos.Y][pos.X]
	rl.DrawText(fmt.Sprintf("Tile Type: %d", tileType), m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 25

	// Draw textures
	rl.DrawText("Textures:", m.uiState.tileInfoPopupX+padding, textY, 16, rl.Black)
	textY += 20

	textures := m.tileGrid.Textures[pos.Y][pos.X]
	rotations := m.tileGrid.TextureRotations[pos.Y][pos.X]

	for i, tex := range textures {
		rotation := rotations[i]
		warningText := ""
		textColor := rl.DarkGray

		if m.tileGrid.missingResourceTiles.Contains(pos, tex) {
			warningText = " !"
			textColor = rl.Yellow
		}

		rl.DrawText(fmt.Sprintf("- %s (%.1f°)%s", tex, rotation, warningText),
			m.uiState.tileInfoPopupX+padding+10, textY, 14, textColor)
		textY += 20
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
