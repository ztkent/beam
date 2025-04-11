# Pixel Map Maker

2D Map Editor  
Design, save, and export game maps compatible with the Beam engine.

## Features

### Map Creation

- Grid-based tile editor with resizable canvas
- Multiple resolution support with dynamic grid sizing
- Real-time tile editing with multi-layer support
- Advanced texture management, with a variety of editing tools
- Viewport navigation for large maps

### Tools

- **Paintbrush**: Freehand tile placement
- **Paint Bucket**: Fill connected areas with same texture
- **Eraser**: Two modes - Full tile or layer-by-layer (long right-click to switch)
- **Selection**: Select and inspect tile properties
- **Layers**: Toggle between ground and wall tiles (long right-click to switch)
- **Location**: Place special locations (long right-click to cycle modes):
  - Player Start
  - Dungeon Entrance (multiple allowed)
  - Respawn Point
  - Exit Point

### Resource Management

- Support for individual textures and sprite sheets
- Automatic sprite sheet slicing with configurable grid size
- Preview slicing and configure sprite sheet options in the 'Spritesheet Viewer' utility
- Recent textures toolbar for quick access
- Resource viewer with preview for all loaded textures
- Resource management mode for removing textures

### File Operations

- Save/Load maps in JSON format
- Auto-save support with session recovery
- Export maps compatible with Beam engine
- Project state persistence including resources

## Quick Start

1. Launch the application
2. Use the add resource button (+) to import textures or sprite sheets
3. Configure sprite sheet options if needed in the viewer
4. Select a tool from the toolbar
5. Click the view button to browse and select textures
6. Click and drag on the grid to place textures
7. Save your map using the save button

## Controls

- **Left Click**: Select tiles, textures, or tools
- **Right Click**: Apply current tool action
- **Long Right Click**: Switch tool modes (e.g., eraser mode)
- **Mouse Wheel**: Scroll resource viewer
- **Ctrl/Cmd + S**: Quick save

### Viewport Navigation

For maps larger than the screen size:

- Navigation buttons appear at the left side of the grid
- Click arrows to move the viewport in any direction
- Visual indicators show available scroll directions
- Viewport automatically adjusts to maintain optimal view size

## Recent Textures

Click the active texture preview to show recently used textures. Recently used textures are available for quick access.
