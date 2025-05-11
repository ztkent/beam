# Beam

2D Game Engine + Tools

## Features

### Core Game

- [x] Game state management
- [x] Tile-based map system with support for:
  - Multiple tile types (Walls, Floors, etc.)
  - Animated multi-frame textures with transitions
  - Custom tile properties (rotation, scale, offset, tinting)
- [x] NPCs
  - Customizable NPC properties (health, attack, etc.)
  - Multi-directional animation support
  - NPC behaviors (wandering, aggro, player tracking, and combat)
  - Chat and interaction system

### Map editor

- [x] Design, save, and export beam-compatible pixel maps
- [x] Grid-based tile editor with resizable canvas
- [x] Real-time tile editing with multi-layer support
- [x] Resource viewer with preview for all loaded textures.
- [x] Place NPCs and set custom properties
- For more details, [view the Map Maker tool](https://github.com/ztkent/beam/tree/main/tools/mapmaker)

### Resource management

- [x] Load and manage game resources (textures, audio, fonts, etc.)
- [x] Support for individual textures and sprite sheets
- [x] Automatic sprite sheet slicing with configurable grid size
- [x] Preview slicing and configure sprite sheet options in the [Spritesheet Viewer](https://github.com/ztkent/beam/tree/main/tools/spritesheet-viewer) utility
- [x] Scenes allow for dynamic loading/unloading of resources
- [x] Support for loading resources from local files or remote URLs
- [x] Simple rendering system for displaying textures and NPCs

### Audio

- [x] Sound effects
- [x] Game tracks
- [x] Per track volume control

### Other

- [x] Saved Game Support
- [x] Highscores

### Tools

[Pixel Map Maker](https://github.com/ztkent/beam/tree/main/tools/mapmaker) - Tool for generating beam-compatible pixel maps  
[Spritesheet Viewer](https://github.com/ztkent/beam/tree/main/tools/spritesheet-viewer) - Tool for viewing and inspecting sprite sheets
