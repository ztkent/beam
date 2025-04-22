package resources

import (
	"fmt"
	"path/filepath"
	"sort"
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

type ResourceState struct {
	Scenes []SceneState `json:"scenes"`
}

type SceneState struct {
	Name         string     `json:"name"`
	Textures     []Resource `json:"textures"`
	SpriteSheets []Resource `json:"spriteSheets"`
	Font         *Resource  `json:"font,omitempty"`
	Loaded       bool       `json:"loaded"`
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
func NewResourceManager() *ResourceManager {
	rm := &ResourceManager{
		Scenes: make([]Scene, 0),
	}
	rm.AddScene("default", nil, nil)
	rm.init()
	return rm
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

			spriteSheet := &SpriteSheet{
				Name:     def.Name,
				Path:     def.Path,
				Sprites:  make(map[string]Rectangle),
				GridSize: gridSize,
				Margin:   def.SheetMargin,
				Loaded:   false,
			}

			// Automatically load all sprites in the sheet. Assign names based on their path & position.
			if len(def.SheetData) == 0 {
				tempTexture := rl.LoadTexture(def.Path)
				fileName := strings.TrimSuffix(filepath.Base(def.Path), filepath.Ext(def.Path))
				def.SheetData = ScanSpriteSheet(def.Name, fileName, tempTexture, gridSize, def.SheetMargin)
				rl.UnloadTexture(tempTexture)
			}

			// Initialize sprite regions
			for spriteName, pos := range def.SheetData {
				spriteSheet.Sprites[spriteName] = Rectangle{
					X:      pos[0] * (gridSize + def.SheetMargin),
					Y:      pos[1] * (gridSize + def.SheetMargin),
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

func ScanSpriteSheet(name string, fileName string, texture rl.Texture2D, spriteSize, margin int32) map[string][]int32 {
	sheetName := fileName
	if name != "" {
		sheetName = name
	}
	sheetData := make(map[string][]int32)
	cols := (texture.Width)/(spriteSize+margin) + 1
	rows := (texture.Height)/(spriteSize+margin) + 1
	for row := int32(0); row < rows; row++ {
		for col := int32(0); col < cols; col++ {
			spriteName := fmt.Sprintf("%s_%d_%d", sheetName, row, col)
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
	Name    string
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
						Name:    tex.Name,
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
			if tex, rect, found := rm.getSpriteFromSheets(&view, textureName); found {
				return TextureInfo{
					Name:    textureName,
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

func (rm *ResourceManager) GetAllTextures(sceneName string, ignoreSheetTextures bool) ([]TextureInfo, error) {
	for _, scene := range rm.Scenes {
		if scene.Name == sceneName {
			var textures []TextureInfo

			// Add regular textures
			for _, tex := range scene.Textures {
				if tex.Loaded {
					textures = append(textures, TextureInfo{
						Name:    tex.Name,
						Texture: tex.Texture,
						Region: rl.Rectangle{
							X:      0,
							Y:      0,
							Width:  float32(tex.Texture.Width),
							Height: float32(tex.Texture.Height),
						},
						IsSheet: false,
					})
				}
			}

			// Add sprite sheet entries
			if !ignoreSheetTextures {
				for _, sheet := range scene.SpriteSheets {
					if sheet.Loaded {
						for name, region := range sheet.Sprites {
							textures = append(textures, TextureInfo{
								Name:    name,
								Texture: sheet.Texture,
								Region: rl.Rectangle{
									X:      float32(region.X),
									Y:      float32(region.Y),
									Width:  float32(region.Width),
									Height: float32(region.Height),
								},
								IsSheet: true,
							})
						}
					}
				}
			}
			sort.Slice(textures, func(i, j int) bool {
				return textures[i].Name < textures[j].Name
			})
			return textures, nil
		}
	}
	return nil, fmt.Errorf("scene not found: %s", sceneName)
}

func (rm *ResourceManager) GetAllSpritesheets(sceneName string) ([]SpriteSheet, error) {
	for _, scene := range rm.Scenes {
		if scene.Name == sceneName {
			var sheets []SpriteSheet

			// Add sprite sheet entries
			for _, sheet := range scene.SpriteSheets {
				if sheet.Loaded {
					sheets = append(sheets, *sheet)
				}
			}

			// Sort by name
			sort.Slice(sheets, func(i, j int) bool {
				return sheets[i].Name < sheets[j].Name
			})
			return sheets, nil
		}
	}
	return nil, fmt.Errorf("scene not found: %s", sceneName)
}

func (rm *ResourceManager) GetFont(viewName string) (rl.Font, error) {
	for _, view := range rm.Scenes {
		if view.Name == viewName && view.Font != nil && view.Font.Loaded {
			return view.Font.Font, nil
		}
	}
	return rl.Font{}, fmt.Errorf("font not found or not loaded in view: %s", viewName)
}

func (rm *ResourceManager) getSpriteFromSheets(view *Scene, spriteName string) (rl.Texture2D, Rectangle, bool) {
	for _, sheet := range view.SpriteSheets {
		if sheet.Loaded {
			if region, ok := sheet.Sprites[spriteName]; ok {
				return sheet.Texture, region, true
			}
		}
	}
	return rl.Texture2D{}, Rectangle{}, false
}

func (rm *ResourceManager) AddResource(sceneName string, resource Resource) error {
	for i := range rm.Scenes {
		if rm.Scenes[i].Name == sceneName {
			view := &rm.Scenes[i]

			// Check for resource name conflicts, require a name if one if isn't provided
			if resource.Name == "" {
				return fmt.Errorf("resource name is required")
			}
			if resource.IsSheet {
				for _, sheet := range view.SpriteSheets {
					if sheet.Name == resource.Name {
						return fmt.Errorf("SpriteSheet name conflict: %s. Name already exists.", resource.Name)
					}
				}
			} else {
				for _, tex := range view.Textures {
					if tex.Name == resource.Name {
						return fmt.Errorf("Texture name conflict: %s. Name already exists.", resource.Name)
					}
				}
			}

			if resource.IsSheet {
				gridSize := resource.GridSize
				if gridSize == 0 {
					gridSize = DefaultGridSize
				}
				spriteSheet := &SpriteSheet{
					Name:     resource.Name,
					Path:     resource.Path,
					Sprites:  make(map[string]Rectangle),
					GridSize: gridSize,
					Margin:   resource.SheetMargin,
					Loaded:   false,
				}

				if len(resource.SheetData) == 0 {
					tempTexture := rl.LoadTexture(resource.Path)
					fileName := strings.TrimSuffix(filepath.Base(resource.Path), filepath.Ext(resource.Path))
					resource.SheetData = ScanSpriteSheet(resource.Name, fileName, tempTexture, gridSize, resource.SheetMargin)
					rl.UnloadTexture(tempTexture)
				}

				for spriteName, pos := range resource.SheetData {
					spriteSheet.Sprites[spriteName] = Rectangle{
						X:      pos[0] * (gridSize + resource.SheetMargin),
						Y:      pos[1] * (gridSize + resource.SheetMargin),
						Width:  gridSize,
						Height: gridSize,
					}
				}
				view.SpriteSheets = append(view.SpriteSheets, spriteSheet)

				// Load the sheet if the scene is currently loaded
				if view.Loaded {
					spriteSheet.Texture = rl.LoadTexture(spriteSheet.Path)
					spriteSheet.Loaded = true
				}
			} else {
				texture := Texture{
					Name:   resource.Name,
					Path:   resource.Path,
					Loaded: false,
				}
				view.Textures = append(view.Textures, texture)

				// Load the texture if the scene is currently loaded
				if view.Loaded {
					texture.Texture = rl.LoadTexture(texture.Path)
					texture.Loaded = true
					view.Textures[len(view.Textures)-1] = texture
				}
			}
			return nil
		}
	}
	return fmt.Errorf("scene not found: %s", sceneName)
}

func (rm *ResourceManager) RemoveResource(sceneName string, resourceName string) error {
	for i := range rm.Scenes {
		if rm.Scenes[i].Name == sceneName {
			view := &rm.Scenes[i]

			// Check and remove from regular textures
			for j := range view.Textures {
				if view.Textures[j].Name == resourceName {
					if view.Textures[j].Loaded {
						rl.UnloadTexture(view.Textures[j].Texture)
					}
					view.Textures = append(view.Textures[:j], view.Textures[j+1:]...)
					return nil
				}
			}

			// Check and remove from sprite sheets
			for j := range view.SpriteSheets {
				if view.SpriteSheets[j].Name == resourceName {
					if view.SpriteSheets[j].Loaded {
						rl.UnloadTexture(view.SpriteSheets[j].Texture)
					}
					view.SpriteSheets = append(view.SpriteSheets[:j], view.SpriteSheets[j+1:]...)
					return nil
				}
			}

			return fmt.Errorf("resource not found: %s", resourceName)
		}
	}
	return fmt.Errorf("scene not found: %s", sceneName)
}

func (rm *ResourceManager) SaveState() ResourceState {
	state := ResourceState{
		Scenes: make([]SceneState, len(rm.Scenes)),
	}

	for i, scene := range rm.Scenes {
		sceneState := SceneState{
			Name:   scene.Name,
			Loaded: scene.Loaded,
		}

		// Save textures
		for _, tex := range scene.Textures {
			sceneState.Textures = append(sceneState.Textures, Resource{
				Name: tex.Name,
				Path: tex.Path,
			})
		}

		// Save sprite sheets
		for _, sheet := range scene.SpriteSheets {
			sheetData := make(map[string][]int32)
			for name, rect := range sheet.Sprites {
				sheetData[name] = []int32{rect.X / (sheet.GridSize + sheet.Margin), rect.Y / (sheet.GridSize + sheet.Margin)}
			}
			sceneState.SpriteSheets = append(sceneState.SpriteSheets, Resource{
				Name:        sheet.Name,
				Path:        sheet.Path,
				IsSheet:     true,
				SheetData:   sheetData,
				SheetMargin: sheet.Margin,
				GridSize:    sheet.GridSize,
			})
		}

		// Save font if present
		if scene.Font != nil {
			sceneState.Font = &Resource{
				Name: scene.Font.Name,
				Path: scene.Font.Path,
			}
		}

		state.Scenes[i] = sceneState
	}

	return state
}

func InitFromState(state ResourceState) *ResourceManager {
	rm := &ResourceManager{
		Scenes: make([]Scene, 0),
	}

	for _, sceneState := range state.Scenes {
		var textureDefs []Resource

		// Convert texture definitions
		for _, tex := range sceneState.Textures {
			textureDefs = append(textureDefs, Resource{
				Name: tex.Name,
				Path: tex.Path,
			})
		}

		// Convert sprite sheet definitions
		for _, sheet := range sceneState.SpriteSheets {
			textureDefs = append(textureDefs, Resource{
				Name:        sheet.Name,
				Path:        sheet.Path,
				IsSheet:     true,
				SheetData:   sheet.SheetData,
				SheetMargin: sheet.SheetMargin,
				GridSize:    sheet.GridSize,
			})
		}

		// Convert font definition
		var fontDef *Resource
		if sceneState.Font != nil {
			fontDef = &Resource{
				Name: sceneState.Font.Name,
				Path: sceneState.Font.Path,
			}
		}

		rm.AddScene(sceneState.Name, textureDefs, fontDef)
		if sceneState.Loaded {
			rm.LoadView(sceneState.Name)
		}
	}

	return rm
}
