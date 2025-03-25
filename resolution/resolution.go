package resolution

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Resolution int

const (
	Resolution1280x800 Resolution = iota
	Resolution1280x720
	Resolution1920x1080
)

var ResolutionOptions = map[Resolution][2]int32{
	Resolution1280x800:  {1280, 800},  // 16:10
	Resolution1280x720:  {1280, 720},  // 16:9
	Resolution1920x1080: {1920, 1080}, // 16:9
}

func InitScreenAtResolution(res Resolution, title string) {
	resOptions := ResolutionOptions[res]
	rl.InitWindow(resOptions[0], resOptions[1], title)
	rl.SetTargetFPS(60)
	return
}

func GetResolution() Resolution {
	switch rl.GetScreenHeight() {
	case 720:
		return Resolution1280x720
	case 1080:
		return Resolution1920x1080
	default:
		return Resolution1280x800
	}
}

func ResolutionFromDimensions(width, height int32) Resolution {
	switch height {
	case 720:
		return Resolution1280x720
	case 1080:
		return Resolution1920x1080
	default:
		return Resolution1280x800
	}
}
