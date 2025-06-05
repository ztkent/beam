package audio

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AudioManager struct {
	Views        []AudioView
	Volume       float32
	CurrentMusic *Music
	IsPlaying    bool
	embeddedFS   fs.FS
}

type AudioView struct {
	Name   string
	Tracks []Music
	SFX    []Sound
}

type Music struct {
	Name         string
	Path         string
	Stream       rl.Music
	Loaded       bool
	EmbeddedData []byte
}
type Sound struct {
	Name         string
	Path         string
	Sound        rl.Sound
	Loaded       bool
	EmbeddedData []byte
}
type Audio struct {
	Name string
	Path string
}

// Define a audio manager that supports global audio.
// This manager will create an audio view called "default" that is always loaded.
// This means you can define the audio once, and not mess with views.
// You can also add new audio views to the manager, and load/unload them as needed.
func NewAudioManagerWithGlobal(defaultMusic []Audio, defaultSounds []Audio) *AudioManager {
	am := &AudioManager{
		Volume: .5,
		Views:  make([]AudioView, 0),
	}
	am.AddAudioView("default", defaultMusic, defaultSounds)
	am.init()
	return am
}

func NewAudioManagerWithGlobalEmbed(defaultMusic []Audio, defaultSounds []Audio, embeddedFS fs.FS) *AudioManager {
	am := &AudioManager{
		Volume:     .5,
		Views:      make([]AudioView, 0),
		embeddedFS: embeddedFS,
	}
	am.AddAudioView("default", defaultMusic, defaultSounds)
	am.init()
	return am
}

// Create a new audio manager.
// This manager will not have any global audio resources loaded.
// You can add new audio views to the manager, and load/unload them as needed.
func NewAudioManager() *AudioManager {
	am := &AudioManager{
		Volume: 1.0,
		Views:  make([]AudioView, 0),
	}
	am.init()
	return am
}

func (am *AudioManager) init() {
	rl.InitAudioDevice()
	rl.SetMasterVolume(am.Volume)
	// Load default audio resources
	am.LoadAudioView("default")
}

// GetEmbeddedFS returns a pointer to the embedded filesystem if available
func (am *AudioManager) GetEmbeddedFS() *fs.FS {
	if am.embeddedFS == nil {
		return nil
	}
	return &am.embeddedFS
}

// LoadMusic wrapper that handles both embedded and file system loading
func (am *AudioManager) LoadMusic(path string) (rl.Music, []byte) {
	if am.embeddedFS != nil {
		return am.loadMusicFromEmbedded(path)
	}
	return rl.LoadMusicStream(path), nil
}

// LoadSound wrapper that handles both embedded and file system loading
func (am *AudioManager) LoadSound(path string) (rl.Sound, []byte) {
	if am.embeddedFS != nil {
		return am.loadSoundFromEmbedded(path)
	}
	return rl.LoadSound(path), nil
}
func (am *AudioManager) loadMusicFromEmbedded(path string) (rl.Music, []byte) {
	data, err := fs.ReadFile(am.embeddedFS, path)
	if err != nil {
		fmt.Printf("Failed to load embedded music %s: %v\n", path, err)
		return rl.Music{}, nil
	}

	// Create a copy of the data to ensure it's not affected by GC
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Determine file extension for proper loading
	ext := strings.ToLower(filepath.Ext(path))
	var music rl.Music

	switch ext {
	case ".mp3":
		music = rl.LoadMusicStreamFromMemory(".mp3", dataCopy, int32(len(dataCopy)))
	case ".wav":
		music = rl.LoadMusicStreamFromMemory(".wav", dataCopy, int32(len(dataCopy)))
	case ".flac":
		music = rl.LoadMusicStreamFromMemory(".flac", dataCopy, int32(len(dataCopy)))
	default:
		fmt.Printf("Unsupported music format for %s\n", path)
		return rl.Music{}, nil
	}

	// Validate the loaded music stream
	if !rl.IsMusicValid(music) {
		fmt.Printf("Failed to load embedded music %s - invalid stream\n", path)
		return rl.Music{}, nil
	}

	return music, dataCopy
}

func (am *AudioManager) loadSoundFromEmbedded(path string) (rl.Sound, []byte) {
	data, err := fs.ReadFile(am.embeddedFS, path)
	if err != nil {
		fmt.Printf("Failed to load embedded sound %s: %v\n", path, err)
		return rl.Sound{}, nil
	}

	// Create a copy of the data to ensure it's not affected by GC
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Determine file extension for proper loading
	ext := strings.ToLower(filepath.Ext(path))
	var wave rl.Wave

	switch ext {
	case ".wav":
		wave = rl.LoadWaveFromMemory(".wav", dataCopy, int32(len(dataCopy)))
	case ".mp3":
		wave = rl.LoadWaveFromMemory(".mp3", dataCopy, int32(len(dataCopy)))
	case ".flac":
		wave = rl.LoadWaveFromMemory(".flac", dataCopy, int32(len(dataCopy)))
	default:
		fmt.Printf("Unsupported sound format for %s\n", path)
		return rl.Sound{}, nil
	}

	if wave.Data == nil {
		fmt.Printf("Failed to decode embedded sound %s\n", path)
		return rl.Sound{}, nil
	}

	sound := rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)
	return sound, dataCopy
}

// Close the audio device and unload all audio resources.
func (am *AudioManager) Close() {
	for _, view := range am.Views {
		for _, track := range view.Tracks {
			if track.Loaded {
				rl.UnloadMusicStream(track.Stream)
			}
		}
		for _, sfx := range view.SFX {
			if sfx.Loaded {
				rl.UnloadSound(sfx.Sound)
			}
		}
	}
	rl.CloseAudioDevice()
}

// AddAudioView adds a new audio view to the audio manager. A view is a collection
// of music tracks and sound effects that can be loaded and unloaded together,
// making it useful for managing audio resources across different game scenes.
//
//   - viewName: unique identifier for the view
//   - musicDefs: list of Audio structs that define music tracks
//   - soundDefs: list of Audio structs that define sound effects
func (am *AudioManager) AddAudioView(viewName string, musicDefs []Audio, soundDefs []Audio) error {
	// Make sure the view doesn't already exist
	for _, view := range am.Views {
		if view.Name == viewName {
			return fmt.Errorf("view already exists: %s", viewName)
		}
	}

	musicTracks := make([]Music, 0)
	for _, def := range musicDefs {
		musicTracks = append(musicTracks, Music{
			Name:   def.Name,
			Path:   def.Path,
			Loaded: false,
		})
	}
	sfxSounds := make([]Sound, 0)
	for _, def := range soundDefs {
		sfxSounds = append(sfxSounds, Sound{
			Name:   def.Name,
			Path:   def.Path,
			Loaded: false,
		})
	}
	am.Views = append(am.Views, AudioView{
		Name:   viewName,
		Tracks: musicTracks,
		SFX:    sfxSounds,
	})
	return nil
}

// LoadAudioView loads all music tracks and sound effects for a given audio view.
// This should be called before playing any audio from the view.
func (am *AudioManager) LoadAudioView(viewName string) error {
	for i := range am.Views {
		if am.Views[i].Name == viewName {
			view := &am.Views[i]
			for j := range view.Tracks {
				if !view.Tracks[j].Loaded {
					stream, embeddedData := am.LoadMusic(view.Tracks[j].Path)
					view.Tracks[j].Stream = stream
					view.Tracks[j].EmbeddedData = embeddedData
					rl.SetMusicVolume(view.Tracks[j].Stream, am.Volume)
					rl.SetMusicPitch(view.Tracks[j].Stream, 1.0)
					view.Tracks[j].Loaded = true
				}
			}
			for j := range view.SFX {
				if !view.SFX[j].Loaded {
					sound, embeddedData := am.LoadSound(view.SFX[j].Path)
					view.SFX[j].Sound = sound
					view.SFX[j].EmbeddedData = embeddedData
					rl.SetSoundVolume(view.SFX[j].Sound, am.Volume)
					rl.SetSoundPitch(view.SFX[j].Sound, 1.0)
					view.SFX[j].Loaded = true
				}
			}
			return nil
		}
	}
	return fmt.Errorf("view not found: %s", viewName)
}

// UnloadAudioView unloads all music tracks and sound effects for a given audio view.
// This should be called when the audio resources are no longer needed.
func (am *AudioManager) UnloadAudioView(viewName string) error {
	for _, view := range am.Views {
		if view.Name == viewName {
			for _, track := range view.Tracks {
				if track.Loaded {
					rl.UnloadMusicStream(track.Stream)
				}
			}
			for _, sfx := range view.SFX {
				if sfx.Loaded {
					rl.UnloadSound(sfx.Sound)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("view not found: %s", viewName)
}

// RemoveAudioView removes an audio view from the audio manager.
// All music tracks and sound effects will be unloaded.
func (am *AudioManager) RemoveAudioView(viewName string) error {
	for i, view := range am.Views {
		if view.Name == viewName {
			for _, track := range view.Tracks {
				if track.Loaded {
					rl.UnloadMusicStream(track.Stream)
				}
			}
			for _, sfx := range view.SFX {
				if sfx.Loaded {
					rl.UnloadSound(sfx.Sound)
				}
			}
			am.Views = append(am.Views[:i], am.Views[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("view not found: %s", viewName)
}

// PlayMusic starts a music track from the given view.
// If a music track is already playing, it will be stopped.
// To continue playing the current music, use UpdateMusic in your game loop.
func (am *AudioManager) PlayMusic(viewName, musicName string) error {
	for _, view := range am.Views {
		if view.Name == viewName {
			for i := range view.Tracks {
				if view.Tracks[i].Name == musicName {
					music := &view.Tracks[i]
					if !music.Loaded {
						return fmt.Errorf("music not loaded: %s", musicName)
					}

					// Stop current music if playing
					if am.CurrentMusic != nil && am.CurrentMusic.Loaded && rl.IsMusicValid(am.CurrentMusic.Stream) {
						fmt.Println("Stopping current music")
						rl.StopMusicStream(am.CurrentMusic.Stream)
						am.IsPlaying = false
					}

					// Validate the music stream before playing
					if !rl.IsMusicValid(music.Stream) {
						return fmt.Errorf("invalid music stream for %s", musicName)
					}

					am.CurrentMusic = music
					fmt.Printf("Playing new music: %s\n", musicName)
					rl.SeekMusicStream(music.Stream, 0.0)
					rl.PlayMusicStream(music.Stream)
					rl.SetMusicVolume(music.Stream, am.Volume)
					am.IsPlaying = true
					fmt.Println("Music started successfully")
					return nil
				}
			}
		}
	}
	return fmt.Errorf("music not found: %s in view %s", musicName, viewName)
}

// PlaySound immediately plays a sound effect from the given view.
func (am *AudioManager) PlaySound(viewName, soundName string) error {
	for _, view := range am.Views {
		if view.Name == viewName {
			for i := range view.SFX {
				if view.SFX[i].Name == soundName {
					sound := view.SFX[i]
					if sound.Loaded {
						rl.SetSoundVolume(sound.Sound, am.Volume)
						rl.PlaySound(sound.Sound)
					}
					return nil
				}
			}
		}
	}
	return fmt.Errorf("sound not found: %s in view %s", soundName, viewName)
}

// UpdateMusic should be called in your game loop to keep your current music playing.
// Example usage:
//
//	currentTime := float32(rl.GetTime())
//	deltaTime := currentTime - lastUpdateTime
//	if deltaTime >= 1.0/60.0 {
//		audioManager.UpdateMusic()
//	}
func (am *AudioManager) UpdateMusic() {
	if am.CurrentMusic == nil || !am.CurrentMusic.Loaded {
		return
	}

	if !rl.IsMusicValid(am.CurrentMusic.Stream) {
		fmt.Printf("Invalid music stream detected, stopping playback\n")
		am.CurrentMusic = nil
		am.IsPlaying = false
		return
	}

	if !rl.IsMusicStreamPlaying(am.CurrentMusic.Stream) && am.IsPlaying {
		fmt.Println("Music ended, restarting...")
		rl.SeekMusicStream(am.CurrentMusic.Stream, 0.0)
		rl.PlayMusicStream(am.CurrentMusic.Stream)
	}

	rl.UpdateMusicStream(am.CurrentMusic.Stream)
}

// Sets the master volume for all audio.
// Volume should be a float between 0 and 100.
func (am *AudioManager) SetMasterVolume(volume float32) {
	am.Volume = volume / 100.0
	rl.SetMasterVolume(am.Volume)
	// Also update current music volume if playing
	if am.CurrentMusic != nil && am.CurrentMusic.Loaded {
		rl.SetMusicVolume(am.CurrentMusic.Stream, am.Volume)
	}
}

// LoadMusic loads a music file from the given path.
// The music will be loaded into memory and ready to play.
func LoadMusic(name string, path string) *Music {
	stream := rl.LoadMusicStream(path)
	return &Music{
		Name:   name,
		Path:   path,
		Stream: stream,
		Loaded: true,
	}
}

// LoadSound loads a sound file from the given path.
// The sound will be loaded into memory and ready to play.
func LoadSound(name string, path string) *Sound {
	sound := rl.LoadSound(path)
	return &Sound{
		Name:   name,
		Path:   path,
		Sound:  sound,
		Loaded: true,
	}
}

// NormalizeSettings holds the parameters for ffmpeg's loudnorm filter.
type NormalizeSettings struct {
	IntegratedLoudness float64 // Target integrated loudness in LUFS (e.g., -23.0)
	TruePeak           float64 // Target true peak in dBTP (e.g., -2.0)
	LoudnessRange      float64 // Target loudness range in LU (e.g., 7.0)
}

// DefaultNormalizeSettings provides common normalization parameters (EBU R128).
var DefaultNormalizeSettings = NormalizeSettings{
	IntegratedLoudness: -23.0,
	TruePeak:           -2.0,
	LoudnessRange:      7.0,
}

// NormalizeAudioFiles processes a list of audio files using ffmpeg's loudnorm filter.
// It creates new files with the suffix "_normalized" before the extension
func NormalizeAudioFiles(inputFilePaths []string, settings *NormalizeSettings) ([]string, error) {
	if len(inputFilePaths) == 0 {
		return nil, fmt.Errorf("no input files provided for normalization")
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg command not found in system PATH: %w", err)
	}

	currentSettings := DefaultNormalizeSettings
	if settings != nil {
		currentSettings = *settings
	}

	var cmdArgs []string
	var filterComplexParts []string
	var outputNormalizedFiles []string

	// Automatically overwrite output files if they exist
	cmdArgs = append(cmdArgs, "-y")

	// Prepare -i arguments and output file names
	for i, inputPath := range inputFilePaths {
		cmdArgs = append(cmdArgs, "-i", inputPath)

		dir := filepath.Dir(inputPath)
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		nameWithoutExt := strings.TrimSuffix(base, ext)
		outputPath := filepath.Join(dir, fmt.Sprintf("%s_normalized%s", nameWithoutExt, ext))
		outputNormalizedFiles = append(outputNormalizedFiles, outputPath)

		filterComplexParts = append(filterComplexParts,
			fmt.Sprintf("[%d:a]loudnorm=I=%.1f:TP=%.1f:LRA=%.1f[norm%d]",
				i, currentSettings.IntegratedLoudness, currentSettings.TruePeak, currentSettings.LoudnessRange, i))
	}

	// Add filter_complex argument
	if len(filterComplexParts) > 0 {
		cmdArgs = append(cmdArgs, "-filter_complex", strings.Join(filterComplexParts, ";"))
	}

	// Prepare -map arguments for each output file
	for i, outputPath := range outputNormalizedFiles {
		cmdArgs = append(cmdArgs, "-map", fmt.Sprintf("[norm%d]", i), outputPath)
	}

	cmd := exec.Command(ffmpegPath, cmdArgs...)

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg execution failed: %w\nffmpeg output:\n%s", err, string(cmdOutput))
	}

	return outputNormalizedFiles, nil
}
