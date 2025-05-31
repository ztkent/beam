package chat

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/beam/controls"
)

type DialogState int

const (
	DialogHidden DialogState = iota
	DialogVisible
	DialogWaiting
	DialogFinished
)

type Dialog struct {
	Text     string
	Duration time.Duration // How long to show before auto-continuing
}

type Chat struct {
	CurrentDialog int
	State         DialogState
	StartTime     time.Time
	Font          rl.Font
	Dialogs       []Dialog
}

func NewChat() *Chat {
	chat := &Chat{
		CurrentDialog: 0,
		State:         DialogHidden,
		Dialogs:       DefaultDialog(),
		Font:          rl.GetFontDefault(),
	}
	return chat
}

func NewChatWithDialogs(dialogs []Dialog) *Chat {
	chat := &Chat{
		CurrentDialog: 0,
		State:         DialogHidden,
		Dialogs:       dialogs,
		Font:          rl.GetFontDefault(),
	}
	return chat
}

func DefaultDialog() []Dialog {
	return []Dialog{
		{
			Text:     "Welcome to the dungeon...",
			Duration: time.Second * 2,
		},
		{
			Text:     "Be careful, dangers lurk in every corner.",
			Duration: time.Second * 2,
		},
		{
			Text:     "Press E to interact with objects you find.",
			Duration: time.Second * 2,
		},
	}
}

func (c *Chat) Update(cm *controls.ControlsManager) {
	switch c.State {
	case DialogVisible:
		// Check if duration has passed
		if time.Since(c.StartTime) >= c.Dialogs[c.CurrentDialog].Duration {
			c.State = DialogWaiting
		}
		// Check for continue input
		if cm.IsActionPressed(controls.ActionConfirm) {
			c.NextDialog()
		}

	case DialogWaiting:
		if cm.IsActionPressed(controls.ActionConfirm) {
			c.NextDialog()
		}
	}
}

func (c *Chat) Show() {
	c.State = DialogVisible
	c.StartTime = time.Now()
}

func (c *Chat) Hide() {
	c.State = DialogHidden
}

func (c *Chat) NextDialog() {
	c.CurrentDialog++
	if c.CurrentDialog >= len(c.Dialogs) {
		c.State = DialogFinished
		c.Hide()
		return
	}
	c.StartTime = time.Now()
	c.State = DialogVisible
}

func (c *Chat) Draw() {
	if c.State == DialogHidden || c.State == DialogFinished {
		return
	}

	dialog := c.Dialogs[c.CurrentDialog]

	// Get screen dimensions
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	// Calculate text dimensions
	textSize := rl.MeasureTextEx(c.Font, dialog.Text, 20, 1)

	// Define dialog box dimensions
	padding := float32(20)
	boxWidth := textSize.X + (padding * 2)
	boxHeight := float32(80)

	// Calculate centered position
	boxX := (screenWidth - boxWidth) / 2
	boxY := screenHeight - boxHeight - padding

	// Draw dialog box background
	rl.DrawRectangle(
		int32(boxX),
		int32(boxY),
		int32(boxWidth),
		int32(boxHeight),
		rl.NewColor(0, 0, 0, 200),
	)

	// Draw border
	rl.DrawRectangleLinesEx(
		rl.Rectangle{
			X:      boxX,
			Y:      boxY,
			Width:  boxWidth,
			Height: boxHeight,
		},
		2,
		rl.White,
	)

	// Draw text
	rl.DrawTextEx(
		c.Font,
		dialog.Text,
		rl.Vector2{
			X: boxX + padding,
			Y: boxY + (boxHeight-textSize.Y)/2,
		},
		20,
		1,
		rl.White,
	)

	// Draw continue prompt if in waiting state
	if c.State == DialogWaiting {
		promptText := "Press SPACE to continue"
		promptSize := rl.MeasureTextEx(c.Font, promptText, 16, 1)
		rl.DrawTextEx(
			c.Font,
			promptText,
			rl.Vector2{
				X: boxX + boxWidth - promptSize.X - padding,
				Y: boxY + boxHeight - promptSize.Y - 5,
			},
			16,
			1,
			rl.NewColor(200, 200, 200, 255),
		)
	}
}
