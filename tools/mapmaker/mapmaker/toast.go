package mapmaker

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
)

type Toast struct {
	message   string
	toastType ToastType
	created   time.Time
	duration  time.Duration
}

func NewToast(message string, toastType ToastType) Toast {
	return Toast{
		message:   message,
		toastType: toastType,
		created:   time.Now(),
		duration:  3 * time.Second,
	}
}

func (t *Toast) isExpired() bool {
	return time.Since(t.created) > t.duration
}

func (t *Toast) getColor() rl.Color {
	switch t.toastType {
	case ToastSuccess:
		return rl.Color{R: 76, G: 175, B: 80, A: 255} // Green
	case ToastError:
		return rl.Color{R: 244, G: 67, B: 54, A: 255} // Red
	default:
		return rl.Color{R: 33, G: 150, B: 243, A: 255} // Blue
	}
}

func (m *MapMaker) showToast(message string, toastType ToastType) {
	t := NewToast(message, toastType)
	m.uiState.toast = &t
}

func (m *MapMaker) renderToast() {
	if m.uiState.toast == nil || m.uiState.toast.isExpired() {
		return
	}

	padding := float32(20)
	fontSize := int32(16)
	textWidth := float32(rl.MeasureText(m.uiState.toast.message, fontSize))
	toastWidth := textWidth + padding*2
	toastHeight := float32(40)

	// Position at bottom center of screen
	toastX := (float32(m.window.width) - toastWidth) / 2
	toastY := toastHeight + padding

	// Calculate fade out for last 0.5 seconds
	alpha := uint8(255)
	timeLeft := m.uiState.toast.duration - time.Since(m.uiState.toast.created)
	if timeLeft < 500*time.Millisecond {
		alpha = uint8(float64(255) * (float64(timeLeft) / float64(500*time.Millisecond)))
	}

	// Draw background with alpha
	bgColor := m.uiState.toast.getColor()
	bgColor.A = alpha
	rl.DrawRectangleRounded(
		rl.Rectangle{X: toastX, Y: toastY, Width: toastWidth, Height: toastHeight},
		0.3,
		8,
		bgColor,
	)

	// Draw text with alpha
	textColor := rl.White
	textColor.A = alpha
	textX := toastX + padding
	textY := toastY + (toastHeight-float32(fontSize))/2
	rl.DrawText(m.uiState.toast.message, int32(textX), int32(textY), fontSize, textColor)
}
