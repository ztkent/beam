package controls

import (
	"encoding/json"
	"fmt"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InputType represents the type of input device
type InputType int

const (
	InputKeyboard InputType = iota
	InputGamepad
	InputMouse
)

// Action represents a game action that can be mapped to inputs
type Action string

const (
	// Movement actions
	ActionMoveUp    Action = "move_up"
	ActionMoveDown  Action = "move_down"
	ActionMoveLeft  Action = "move_left"
	ActionMoveRight Action = "move_right"

	// Game actions
	ActionAttack    Action = "attack"
	ActionInteract  Action = "interact"
	ActionEquip     Action = "equip"
	ActionInventory Action = "inventory"
	ActionPause     Action = "pause"

	// UI actions
	ActionConfirm        Action = "confirm"
	ActionCancel         Action = "cancel"
	ActionToggleHUD      Action = "toggle_hud"
	ActionToggleStats    Action = "toggle_stats"
	ActionToggleControls Action = "toggle_controls"

	// Menu navigation
	ActionMenuUp    Action = "menu_up"
	ActionMenuDown  Action = "menu_down"
	ActionMenuLeft  Action = "menu_left"
	ActionMenuRight Action = "menu_right"
)

// InputBinding represents a binding between an action and an input
type InputBinding struct {
	Type     InputType      `json:"type"`
	Key      int32          `json:"key,omitempty"`      // For keyboard
	Button   rl.MouseButton `json:"button,omitempty"`   // For gamepad/mouse
	Axis     int32          `json:"axis,omitempty"`     // For gamepad analog sticks
	Positive bool           `json:"positive,omitempty"` // For axis direction
	Gamepad  int32          `json:"gamepad,omitempty"`  // Gamepad index
}

// ControlScheme holds all input mappings
type ControlScheme struct {
	Name     string                    `json:"name"`
	Bindings map[Action][]InputBinding `json:"bindings"`
}

// ControlsManager manages input handling and mapping
type ControlsManager struct {
	schemes      map[string]*ControlScheme
	activeScheme string
	gamepadIndex int32
	deadzone     float32
	configPath   string

	// State tracking for edge detection
	previousKeyState    map[int32]bool
	previousButtonState map[int32]bool
	previousMouseState  map[int32]bool
}

// NewControlsManager creates a new controls manager with default schemes
func NewControlsManager(configPath string) *ControlsManager {
	cm := &ControlsManager{
		schemes:             make(map[string]*ControlScheme),
		activeScheme:        "keyboard",
		gamepadIndex:        0,
		deadzone:            0.1,
		configPath:          configPath,
		previousKeyState:    make(map[int32]bool),
		previousButtonState: make(map[int32]bool),
		previousMouseState:  make(map[int32]bool),
	}

	// Create default control schemes
	cm.createDefaultSchemes()

	// Try to load from config
	cm.LoadConfig()

	return cm
}

// createDefaultSchemes sets up default keyboard and gamepad control schemes
func (cm *ControlsManager) createDefaultSchemes() {
	// Default keyboard scheme
	keyboardScheme := &ControlScheme{
		Name:     "Keyboard & Mouse",
		Bindings: make(map[Action][]InputBinding),
	}

	// Movement
	keyboardScheme.Bindings[ActionMoveUp] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyW},
		{Type: InputKeyboard, Key: rl.KeyUp},
	}
	keyboardScheme.Bindings[ActionMoveDown] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyS},
		{Type: InputKeyboard, Key: rl.KeyDown},
	}
	keyboardScheme.Bindings[ActionMoveLeft] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyA},
		{Type: InputKeyboard, Key: rl.KeyLeft},
	}
	keyboardScheme.Bindings[ActionMoveRight] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyD},
		{Type: InputKeyboard, Key: rl.KeyRight},
	}

	// Game actions
	keyboardScheme.Bindings[ActionAttack] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeySpace},
		{Type: InputMouse, Button: rl.MouseButtonLeft},
	}
	keyboardScheme.Bindings[ActionInteract] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyQ},
	}
	keyboardScheme.Bindings[ActionEquip] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyE},
	}
	keyboardScheme.Bindings[ActionInventory] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyI},
		{Type: InputKeyboard, Key: rl.KeyTab},
	}
	keyboardScheme.Bindings[ActionPause] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyEscape},
		{Type: InputKeyboard, Key: rl.KeyP},
	}

	// UI actions
	keyboardScheme.Bindings[ActionConfirm] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyEnter},
		{Type: InputKeyboard, Key: rl.KeySpace},
	}
	keyboardScheme.Bindings[ActionCancel] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyEscape},
	}
	keyboardScheme.Bindings[ActionToggleHUD] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyH},
	}
	keyboardScheme.Bindings[ActionToggleStats] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyC},
	}
	keyboardScheme.Bindings[ActionToggleControls] = []InputBinding{
		{Type: InputKeyboard, Key: rl.KeyK},
	}

	cm.schemes["keyboard"] = keyboardScheme

	// Default gamepad scheme
	gamepadScheme := &ControlScheme{
		Name:     "Gamepad",
		Bindings: make(map[Action][]InputBinding),
	}

	// Movement (D-pad and left stick)
	gamepadScheme.Bindings[ActionMoveUp] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonLeftFaceUp, Gamepad: 0},
		{Type: InputGamepad, Axis: rl.GamepadAxisLeftY, Positive: false, Gamepad: 0},
	}
	gamepadScheme.Bindings[ActionMoveDown] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonLeftFaceDown, Gamepad: 0},
		{Type: InputGamepad, Axis: rl.GamepadAxisLeftY, Positive: true, Gamepad: 0},
	}
	gamepadScheme.Bindings[ActionMoveLeft] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonLeftFaceLeft, Gamepad: 0},
		{Type: InputGamepad, Axis: rl.GamepadAxisLeftX, Positive: false, Gamepad: 0},
	}
	gamepadScheme.Bindings[ActionMoveRight] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonLeftFaceRight, Gamepad: 0},
		{Type: InputGamepad, Axis: rl.GamepadAxisLeftX, Positive: true, Gamepad: 0},
	}

	// Game actions
	gamepadScheme.Bindings[ActionAttack] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonRightFaceDown, Gamepad: 0}, // A/X button
		{Type: InputGamepad, Button: rl.GamepadButtonRightTrigger1, Gamepad: 0}, // Right bumper
	}
	gamepadScheme.Bindings[ActionInteract] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonRightFaceLeft, Gamepad: 0}, // Y/Square button
	}
	gamepadScheme.Bindings[ActionEquip] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonRightFaceLeft, Gamepad: 0}, // Y/Square button
	}
	gamepadScheme.Bindings[ActionInventory] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonRightFaceUp, Gamepad: 0}, // X/Triangle button
	}
	gamepadScheme.Bindings[ActionPause] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonMiddle, Gamepad: 0}, // Start/Options button
	}

	// UI actions
	gamepadScheme.Bindings[ActionConfirm] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonRightFaceDown, Gamepad: 0}, // A/X button
	}
	gamepadScheme.Bindings[ActionCancel] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonRightFaceRight, Gamepad: 0}, // B/Circle button
	}
	gamepadScheme.Bindings[ActionToggleHUD] = []InputBinding{
		{Type: InputGamepad, Button: rl.GamepadButtonLeftTrigger1, Gamepad: 0}, // Left bumper
	}

	// Menu navigation (D-pad and left stick)
	gamepadScheme.Bindings[ActionMenuUp] = gamepadScheme.Bindings[ActionMoveUp]
	gamepadScheme.Bindings[ActionMenuDown] = gamepadScheme.Bindings[ActionMoveDown]
	gamepadScheme.Bindings[ActionMenuLeft] = gamepadScheme.Bindings[ActionMoveLeft]
	gamepadScheme.Bindings[ActionMenuRight] = gamepadScheme.Bindings[ActionMoveRight]

	cm.schemes["gamepad"] = gamepadScheme
}

// Update should be called once per frame to update input state
func (cm *ControlsManager) Update() {
	// Auto-switch to gamepad if one becomes available and we're on keyboard
	if cm.activeScheme == "keyboard" && rl.IsGamepadAvailable(cm.gamepadIndex) {
		// Check if any gamepad button was pressed
		if rl.GetGamepadButtonPressed() != rl.GamepadButtonUnknown {
			cm.SetActiveScheme("gamepad")
		}
	}

	// Auto-switch to keyboard if gamepad disconnects
	if cm.activeScheme == "gamepad" && !rl.IsGamepadAvailable(cm.gamepadIndex) {
		cm.SetActiveScheme("keyboard")
	}

	// Switch back to keyboard if any key is pressed while on gamepad
	if cm.activeScheme == "gamepad" && rl.GetKeyPressed() != 0 {
		cm.SetActiveScheme("keyboard")
	}
}

// IsActionPressed returns true if the action was just pressed this frame
func (cm *ControlsManager) IsActionPressed(action Action) bool {
	scheme := cm.schemes[cm.activeScheme]
	if scheme == nil {
		return false
	}

	bindings, exists := scheme.Bindings[action]
	if !exists {
		return false
	}

	for _, binding := range bindings {
		if cm.isBindingPressed(binding) {
			return true
		}
	}
	return false
}

// IsActionDown returns true if the action is currently being held
func (cm *ControlsManager) IsActionDown(action Action) bool {
	scheme := cm.schemes[cm.activeScheme]
	if scheme == nil {
		return false
	}

	bindings, exists := scheme.Bindings[action]
	if !exists {
		return false
	}

	for _, binding := range bindings {
		if cm.isBindingDown(binding) {
			return true
		}
	}
	return false
}

// IsActionReleased returns true if the action was just released this frame
func (cm *ControlsManager) IsActionReleased(action Action) bool {
	scheme := cm.schemes[cm.activeScheme]
	if scheme == nil {
		return false
	}

	bindings, exists := scheme.Bindings[action]
	if !exists {
		return false
	}

	for _, binding := range bindings {
		if cm.isBindingReleased(binding) {
			return true
		}
	}
	return false
}

// GetActionAxis returns analog input value for movement actions (-1.0 to 1.0)
func (cm *ControlsManager) GetActionAxis(positiveAction, negativeAction Action) float32 {
	var value float32 = 0.0

	if cm.IsActionDown(positiveAction) {
		value += 1.0
	}
	if cm.IsActionDown(negativeAction) {
		value -= 1.0
	}

	// Also check for analog gamepad input
	if cm.activeScheme == "gamepad" {
		scheme := cm.schemes[cm.activeScheme]
		if scheme != nil {
			// Check positive action for analog input
			if bindings, exists := scheme.Bindings[positiveAction]; exists {
				for _, binding := range bindings {
					if binding.Type == InputGamepad && binding.Axis >= 0 {
						axisValue := rl.GetGamepadAxisMovement(binding.Gamepad, binding.Axis)
						if binding.Positive && axisValue > cm.deadzone {
							value = axisValue
						} else if !binding.Positive && axisValue < -cm.deadzone {
							value = -axisValue
						}
					}
				}
			}
		}
	}

	return value
}

// isBindingPressed checks if a specific binding was just pressed
func (cm *ControlsManager) isBindingPressed(binding InputBinding) bool {
	switch binding.Type {
	case InputKeyboard:
		return rl.IsKeyPressed(binding.Key)
	case InputGamepad:
		if !rl.IsGamepadAvailable(binding.Gamepad) {
			return false
		}
		if binding.Axis >= 0 {
			// Axis binding - check for edge transition
			axisValue := rl.GetGamepadAxisMovement(binding.Gamepad, binding.Axis)
			key := binding.Axis*2 + map[bool]int32{false: 0, true: 1}[binding.Positive]

			var currentState bool
			if binding.Positive {
				currentState = axisValue > cm.deadzone
			} else {
				currentState = axisValue < -cm.deadzone
			}

			previousState := cm.previousButtonState[key]
			cm.previousButtonState[key] = currentState
			return currentState && !previousState
		} else {
			return rl.IsGamepadButtonPressed(binding.Gamepad, int32(binding.Button))
		}
	case InputMouse:
		return rl.IsMouseButtonPressed(binding.Button)
	}
	return false
}

// isBindingDown checks if a specific binding is currently held
func (cm *ControlsManager) isBindingDown(binding InputBinding) bool {
	switch binding.Type {
	case InputKeyboard:
		return rl.IsKeyDown(binding.Key)
	case InputGamepad:
		if !rl.IsGamepadAvailable(binding.Gamepad) {
			return false
		}
		if binding.Axis >= 0 {
			axisValue := rl.GetGamepadAxisMovement(binding.Gamepad, binding.Axis)
			if binding.Positive {
				return axisValue > cm.deadzone
			} else {
				return axisValue < -cm.deadzone
			}
		} else {
			return rl.IsGamepadButtonDown(binding.Gamepad, int32(binding.Button))
		}
	case InputMouse:
		return rl.IsMouseButtonDown(binding.Button)
	}
	return false
}

// isBindingReleased checks if a specific binding was just released
func (cm *ControlsManager) isBindingReleased(binding InputBinding) bool {
	switch binding.Type {
	case InputKeyboard:
		return rl.IsKeyReleased(binding.Key)
	case InputGamepad:
		if !rl.IsGamepadAvailable(binding.Gamepad) {
			return false
		}
		if binding.Axis >= 0 {
			// Axis binding - check for edge transition
			axisValue := rl.GetGamepadAxisMovement(binding.Gamepad, binding.Axis)
			key := binding.Axis*2 + map[bool]int32{false: 0, true: 1}[binding.Positive]

			var currentState bool
			if binding.Positive {
				currentState = axisValue > cm.deadzone
			} else {
				currentState = axisValue < -cm.deadzone
			}

			previousState := cm.previousButtonState[key]
			cm.previousButtonState[key] = currentState
			return !currentState && previousState
		} else {
			return rl.IsGamepadButtonReleased(binding.Gamepad, int32(binding.Button))
		}
	case InputMouse:
		return rl.IsMouseButtonReleased(binding.Button)
	}
	return false
}

// Configuration methods

// SetActiveScheme switches to a different control scheme
func (cm *ControlsManager) SetActiveScheme(schemeName string) {
	if _, exists := cm.schemes[schemeName]; exists {
		cm.activeScheme = schemeName
	}
}

// GetActiveScheme returns the currently active scheme name
func (cm *ControlsManager) GetActiveScheme() *ControlScheme {
	return cm.schemes[cm.activeScheme]
}

// GetAvailableSchemes returns a list of available scheme names
func (cm *ControlsManager) GetAvailableSchemes() []string {
	schemes := make([]string, 0, len(cm.schemes))
	for name := range cm.schemes {
		schemes = append(schemes, name)
	}
	return schemes
}

// SetGamepadIndex sets which gamepad to use (for multiple gamepad support)
func (cm *ControlsManager) SetGamepadIndex(index int32) {
	cm.gamepadIndex = index
}

// SetDeadzone sets the analog stick deadzone
func (cm *ControlsManager) SetDeadzone(deadzone float32) {
	cm.deadzone = deadzone
}

// IsGamepadConnected returns true if a gamepad is connected
func (cm *ControlsManager) IsGamepadConnected() bool {
	return rl.IsGamepadAvailable(cm.gamepadIndex)
}

// GetGamepadName returns the name of the connected gamepad
func (cm *ControlsManager) GetGamepadName() string {
	if cm.IsGamepadConnected() {
		return rl.GetGamepadName(cm.gamepadIndex)
	}
	return ""
}

// SaveConfig saves the current control configuration to file
func (cm *ControlsManager) SaveConfig() error {
	data, err := json.MarshalIndent(map[string]interface{}{
		"activeScheme": cm.activeScheme,
		"gamepadIndex": cm.gamepadIndex,
		"deadzone":     cm.deadzone,
		"schemes":      cm.schemes,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// LoadConfig loads control configuration from file
func (cm *ControlsManager) LoadConfig() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err // File doesn't exist yet, use defaults
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	if activeScheme, ok := config["activeScheme"].(string); ok {
		cm.activeScheme = activeScheme
	}

	if gamepadIndex, ok := config["gamepadIndex"].(float64); ok {
		cm.gamepadIndex = int32(gamepadIndex)
	}

	if deadzone, ok := config["deadzone"].(float64); ok {
		cm.deadzone = float32(deadzone)
	}

	if schemes, ok := config["schemes"].(map[string]interface{}); ok {
		for name, schemeData := range schemes {
			schemeBytes, _ := json.Marshal(schemeData)
			var scheme ControlScheme
			if json.Unmarshal(schemeBytes, &scheme) == nil {
				cm.schemes[name] = &scheme
			}
		}
	}

	return nil
}

// AddCustomBinding adds a new binding to an action
func (cm *ControlsManager) AddCustomBinding(schemeName string, action Action, binding InputBinding) error {
	scheme, exists := cm.schemes[schemeName]
	if !exists {
		return fmt.Errorf("scheme %s not found", schemeName)
	}

	if scheme.Bindings[action] == nil {
		scheme.Bindings[action] = make([]InputBinding, 0)
	}

	scheme.Bindings[action] = append(scheme.Bindings[action], binding)
	return nil
}

// RemoveBinding removes a binding from an action
func (cm *ControlsManager) RemoveBinding(schemeName string, action Action, bindingIndex int) error {
	scheme, exists := cm.schemes[schemeName]
	if !exists {
		return fmt.Errorf("scheme %s not found", schemeName)
	}

	bindings := scheme.Bindings[action]
	if bindingIndex < 0 || bindingIndex >= len(bindings) {
		return fmt.Errorf("binding index out of range")
	}

	scheme.Bindings[action] = append(bindings[:bindingIndex], bindings[bindingIndex+1:]...)
	return nil
}

// GetBindingsForAction returns all bindings for a specific action
func (cm *ControlsManager) GetBindingsForAction(schemeName string, action Action) ([]InputBinding, error) {
	scheme, exists := cm.schemes[schemeName]
	if !exists {
		return nil, fmt.Errorf("scheme %s not found", schemeName)
	}

	return scheme.Bindings[action], nil
}

func KeyCodeToString(key int32) string {
	switch key {
	case rl.KeyNull:
		return "None"
	case rl.KeyApostrophe:
		return "'"
	case rl.KeyComma:
		return ","
	case rl.KeyMinus:
		return "-"
	case rl.KeyPeriod:
		return "."
	case rl.KeySlash:
		return "/"
	case rl.KeyZero:
		return "0"
	case rl.KeyOne:
		return "1"
	case rl.KeyTwo:
		return "2"
	case rl.KeyThree:
		return "3"
	case rl.KeyFour:
		return "4"
	case rl.KeyFive:
		return "5"
	case rl.KeySix:
		return "6"
	case rl.KeySeven:
		return "7"
	case rl.KeyEight:
		return "8"
	case rl.KeyNine:
		return "9"
	case rl.KeySemicolon:
		return ";"
	case rl.KeyEqual:
		return "="
	case rl.KeyA:
		return "A"
	case rl.KeyB:
		return "B"
	case rl.KeyC:
		return "C"
	case rl.KeyD:
		return "D"
	case rl.KeyE:
		return "E"
	case rl.KeyF:
		return "F"
	case rl.KeyG:
		return "G"
	case rl.KeyH:
		return "H"
	case rl.KeyI:
		return "I"
	case rl.KeyJ:
		return "J"
	case rl.KeyK:
		return "K"
	case rl.KeyL:
		return "L"
	case rl.KeyM:
		return "M"
	case rl.KeyN:
		return "N"
	case rl.KeyO:
		return "O"
	case rl.KeyP:
		return "P"
	case rl.KeyQ:
		return "Q"
	case rl.KeyR:
		return "R"
	case rl.KeyS:
		return "S"
	case rl.KeyT:
		return "T"
	case rl.KeyU:
		return "U"
	case rl.KeyV:
		return "V"
	case rl.KeyW:
		return "W"
	case rl.KeyX:
		return "X"
	case rl.KeyY:
		return "Y"
	case rl.KeyZ:
		return "Z"
	case rl.KeySpace:
		return "Space"
	case rl.KeyEscape:
		return "Escape"
	case rl.KeyEnter:
		return "Enter"
	case rl.KeyTab:
		return "Tab"
	case rl.KeyBackspace:
		return "Backspace"
	case rl.KeyInsert:
		return "Insert"
	case rl.KeyDelete:
		return "Delete"
	case rl.KeyRight:
		return "Right"
	case rl.KeyLeft:
		return "Left"
	case rl.KeyDown:
		return "Down"
	case rl.KeyUp:
		return "Up"
	case rl.KeyPageUp:
		return "Page Up"
	case rl.KeyPageDown:
		return "Page Down"
	case rl.KeyHome:
		return "Home"
	case rl.KeyEnd:
		return "End"
	case rl.KeyCapsLock:
		return "Caps Lock"
	case rl.KeyScrollLock:
		return "Scroll Lock"
	case rl.KeyNumLock:
		return "Num Lock"
	case rl.KeyPrintScreen:
		return "Print Screen"
	case rl.KeyPause:
		return "Pause"
	case rl.KeyF1:
		return "F1"
	case rl.KeyF2:
		return "F2"
	case rl.KeyF3:
		return "F3"
	case rl.KeyF4:
		return "F4"
	case rl.KeyF5:
		return "F5"
	case rl.KeyF6:
		return "F6"
	case rl.KeyF7:
		return "F7"
	case rl.KeyF8:
		return "F8"
	case rl.KeyF9:
		return "F9"
	case rl.KeyF10:
		return "F10"
	case rl.KeyF11:
		return "F11"
	case rl.KeyF12:
		return "F12"
	case rl.KeyLeftShift:
		return "Left Shift"
	case rl.KeyLeftControl:
		return "Left Control"
	case rl.KeyLeftAlt:
		return "Left Alt"
	case rl.KeyRightShift:
		return "Right Shift"
	case rl.KeyRightControl:
		return "Right Control"
	case rl.KeyRightAlt:
		return "Right Alt"
	default:
		return fmt.Sprintf("Key(%d)", key)
	}
}
