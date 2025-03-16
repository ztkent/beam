package resources

import (
	"fmt"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	DefaultGridSize int32 = 16
	DefaultMargin   int32 = 1
)

type ResourceManager struct {
	Scenes []Scene
}

type Scene struct {
	Name         string
	Textures     []Texture
	SpriteSheets []*SpriteSheet
	Font         *Font
	Loaded       bool
}

type Font struct {
	Name   string
	Path   string
	Font   rl.Font
	Loaded bool
}

type Texture struct {
	Name    string
	Path    string
	Texture rl.Texture2D
	Loaded  bool
}

type SpriteSheet struct {
	Name     string
	Path     string
	Texture  rl.Texture2D
	Sprites  map[string]Rectangle
	GridSize int32
	Margin   int32
	Loaded   bool
}

type Rectangle struct {
	X      int32
	Y      int32
	Width  int32
	Height int32
}

type Resource struct {
	Name        string
	Path        string
	IsSheet     bool
	SheetData   map[string][]int32 // map[spriteName][x,y] coordinates in grid units
	SheetMargin int32
	GridSize    int32
}

// NewResourceManagerWithGlobal creates a resource manager with a default view that's always loaded
func NewResourceManagerWithGlobal(defaultTextures []Resource, defaultFont *Resource) *ResourceManager {
	rm := &ResourceManager{
		Scenes: make([]Scene, 0),
	}

	// Add default view
	if defaultFont != nil {
		rm.AddScene("default", defaultTextures, defaultFont)
	} else {
		rm.AddScene("default", defaultTextures, nil)
	}
	rm.init()
	return rm
}

// NewResourceManager creates a resource manager without any default resources
func NewResourceManager(screenWidth, screenHeight int32) *ResourceManager {
	return &ResourceManager{
		Scenes: make([]Scene, 0),
	}
}

func (rm *ResourceManager) init() {
	rm.LoadView("default")
}

func (rm *ResourceManager) Close() {
	if rm == nil {
		return
	}

	for _, scene := range rm.Scenes {
		if scene.Loaded {
			if scene.Font != nil && scene.Font.Loaded {
				rl.UnloadFont(scene.Font.Font)
			}
			for _, tex := range scene.Textures {
				if tex.Loaded {
					rl.UnloadTexture(tex.Texture)
				}
			}
			for _, sheet := range scene.SpriteSheets {
				if sheet.Loaded {
					rl.UnloadTexture(sheet.Texture)
				}
			}
		}
	}
}

func (rm *ResourceManager) AddScene(sceneName string, textureDefs []Resource, fontDef *Resource) error {
	// Check for duplicate view
	for _, scene := range rm.Scenes {
		if scene.Name == sceneName {
			return fmt.Errorf("view already exists: %s", sceneName)
		}
	}

	textures := make([]Texture, 0)
	spriteSheets := make([]*SpriteSheet, 0)

	for _, def := range textureDefs {
		if def.IsSheet {
			gridSize := def.GridSize
			if gridSize == 0 {
				gridSize = DefaultGridSize
			}
			margin := def.SheetMargin
			if margin == 0 {
				margin = DefaultMargin
			}

			spriteSheet := &SpriteSheet{
				Name:     def.Name,
				Path:     def.Path,
				Sprites:  make(map[string]Rectangle),
				GridSize: gridSize,
				Margin:   margin,
				Loaded:   false,
			}

			// Automatically load all sprites in the sheet. Assign names based on their path & position.
			if len(def.SheetData) == 0 {
				tempTexture := rl.LoadTexture(def.Path)
				fileName := strings.TrimSuffix(filepath.Base(def.Path), filepath.Ext(def.Path))
				def.SheetData = ScanSpriteSheet(fileName, tempTexture, gridSize, margin)
				rl.UnloadTexture(tempTexture)
			}

			// Initialize sprite regions
			for spriteName, pos := range def.SheetData {
				spriteSheet.Sprites[spriteName] = Rectangle{
					X:      pos[0] * (gridSize + margin),
					Y:      pos[1] * (gridSize + margin),
					Width:  gridSize,
					Height: gridSize,
				}
			}
			spriteSheets = append(spriteSheets, spriteSheet)
		} else {
			textures = append(textures, Texture{
				Name:   def.Name,
				Path:   def.Path,
				Loaded: false,
			})
		}
	}

	var font *Font
	if fontDef != nil {
		font = &Font{
			Name:   fontDef.Name,
			Path:   fontDef.Path,
			Loaded: false,
		}
	}

	rm.Scenes = append(rm.Scenes, Scene{
		Name:         sceneName,
		Textures:     textures,
		SpriteSheets: spriteSheets,
		Font:         font,
		Loaded:       false,
	})

	return nil
}

func ScanSpriteSheet(fileName string, texture rl.Texture2D, spriteSize, margin int32) map[string][]int32 {
	sheetData := make(map[string][]int32)
	cols := (texture.Width) / (spriteSize + margin)
	rows := (texture.Height) / (spriteSize + margin)
	for row := int32(0); row < rows; row++ {
		for col := int32(0); col < cols; col++ {
			spriteName := fmt.Sprintf("%s_%d_%d", fileName, row, col)
			sheetData[spriteName] = []int32{col, row}
		}
	}
	return sheetData
}

func (rm *ResourceManager) LoadView(viewName string) error {
	for i := range rm.Scenes {
		if rm.Scenes[i].Name == viewName {
			view := &rm.Scenes[i]

			// Load sprite sheets if present
			for _, sheet := range view.SpriteSheets {
				if !sheet.Loaded {
					sheet.Texture = rl.LoadTexture(sheet.Path)
					sheet.Loaded = true
				}
			}

			// Load font if specified
			if view.Font != nil && !view.Font.Loaded {
				view.Font.Font = rl.LoadFont(view.Font.Path)
				view.Font.Loaded = true
			}

			// Load textures
			for j := range view.Textures {
				tex := &view.Textures[j]
				if !tex.Loaded {
					tex.Texture = rl.LoadTexture(tex.Path)
					tex.Loaded = true
				}
			}

			view.Loaded = true
			return nil
		}
	}
	return fmt.Errorf("view not found: %s", viewName)
}

func (rm *ResourceManager) UnloadView(viewName string) error {
	for i := range rm.Scenes {
		if rm.Scenes[i].Name == viewName {
			view := &rm.Scenes[i]

			for _, sheet := range view.SpriteSheets {
				if sheet.Loaded {
					rl.UnloadTexture(sheet.Texture)
					sheet.Loaded = false
				}
			}

			if view.Font != nil && view.Font.Loaded {
				rl.UnloadFont(view.Font.Font)
				view.Font.Loaded = false
			}

			for j := range view.Textures {
				tex := &view.Textures[j]
				if tex.Loaded {
					rl.UnloadTexture(tex.Texture)
					tex.Loaded = false
				}
			}

			view.Loaded = false
			return nil
		}
	}
	return fmt.Errorf("view not found: %s", viewName)
}

func (rm *ResourceManager) RemoveView(viewName string) error {
	for i, view := range rm.Scenes {
		if view.Name == viewName {
			// Unload resources first
			rm.UnloadView(viewName)
			// Remove view from slice
			rm.Scenes = append(rm.Scenes[:i], rm.Scenes[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("view not found: %s", viewName)
}

type TextureInfo struct {
	Texture rl.Texture2D
	Region  rl.Rectangle
	IsSheet bool
}

func (rm *ResourceManager) GetTexture(viewName, textureName string) (TextureInfo, error) {
	for _, view := range rm.Scenes {
		if view.Name == viewName {
			// Check regular textures
			for _, tex := range view.Textures {
				if tex.Name == textureName && tex.Loaded {
					return TextureInfo{
						Texture: tex.Texture,
						Region: rl.Rectangle{
							X:      0,
							Y:      0,
							Width:  float32(tex.Texture.Width),
							Height: float32(tex.Texture.Height),
						},
						IsSheet: false,
					}, nil
				}
			}

			// If not found, check sprite sheets
			if tex, rect, found := rm.GetSpriteFromSheets(&view, textureName); found {
				return TextureInfo{
					Texture: tex,
					Region: rl.Rectangle{
						X:      float32(rect.X),
						Y:      float32(rect.Y),
						Width:  float32(rect.Width),
						Height: float32(rect.Height),
					},
					IsSheet: true,
				}, nil
			}

			return TextureInfo{}, fmt.Errorf("texture not found: %s", textureName)
		}
	}
	return TextureInfo{}, fmt.Errorf("view not found: %s", viewName)
}

func (rm *ResourceManager) GetFont(viewName string) (rl.Font, error) {
	for _, view := range rm.Scenes {
		if view.Name == viewName && view.Font != nil && view.Font.Loaded {
			return view.Font.Font, nil
		}
	}
	return rl.Font{}, fmt.Errorf("font not found or not loaded in view: %s", viewName)
}

func (rm *ResourceManager) GetSpriteFromSheets(view *Scene, spriteName string) (rl.Texture2D, Rectangle, bool) {
	for _, sheet := range view.SpriteSheets {
		if sheet.Loaded {
			if region, ok := sheet.Sprites[spriteName]; ok {
				return sheet.Texture, region, true
			}
		}
	}
	return rl.Texture2D{}, Rectangle{}, false
}

func (rm *ResourceManager) GetSprite(viewName, spriteName string) (rl.Texture2D, Rectangle, error) {
	for _, view := range rm.Scenes {
		if view.Name == viewName {
			if tex, rect, found := rm.GetSpriteFromSheets(&view, spriteName); found {
				return tex, rect, nil
			}
			return rl.Texture2D{}, Rectangle{}, fmt.Errorf("sprite not found: %s", spriteName)
		}
	}
	return rl.Texture2D{}, Rectangle{}, fmt.Errorf("view not found: %s", viewName)
}
