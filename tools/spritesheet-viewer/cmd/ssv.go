package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam/tools/spritesheet-viewer/viewer"
)

func main() {
	rl.InitWindow(800, 600, "Sprite Sheet Viewer")
	rl.SetTargetFPS(60)

	viewer := viewer.NewViewer()
	defer rl.CloseWindow()
	defer viewer.UIState.RM.Close()

	showSettings := false
	rl.SetExitKey(0)

	for !rl.WindowShouldClose() {
		viewer.UIState.HandleInput(&showSettings)
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		viewer.UIState.RenderSprites(viewer.Cfg)
		viewer.UIState.RenderUI(viewer.Cfg, &showSettings)
		rl.EndDrawing()
	}
}
