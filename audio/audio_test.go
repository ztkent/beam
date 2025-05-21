package audio

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNormalizeAudioFiles_Integration tests the NormalizeAudioFiles function by
// attempting to normalize a test audio file using ffmpeg.
func TestNormalizeAudioFiles_Integration(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH, skipping integration test")
	}
	testCases := []struct {
		inputFilename        string
		settings             NormalizeSettings
		expectedOutputSuffix string
	}{
		{
			inputFilename:        "test.mp3",
			settings:             NormalizeSettings{IntegratedLoudness: -30.0, TruePeak: -2.0, LoudnessRange: 7.0},
			expectedOutputSuffix: "_normalized.mp3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.inputFilename, func(t *testing.T) {
			normalizedFiles, err := NormalizeAudioFiles([]string{tc.inputFilename}, &tc.settings)
			if err != nil {
				t.Fatalf("NormalizeAudioFiles failed for %s: %v", tc.inputFilename, err)
			}

			if len(normalizedFiles) != 1 {
				t.Fatalf("Expected 1 normalized file, got %d", len(normalizedFiles))
			}

			expectedOutputPath := tc.inputFilename[:len(tc.inputFilename)-len(filepath.Ext(tc.inputFilename))] + tc.expectedOutputSuffix
			if normalizedFiles[0] != expectedOutputPath {
				t.Errorf("Expected output path %s, got %s", expectedOutputPath, normalizedFiles[0])
			}

			if _, err := os.Stat(expectedOutputPath); os.IsNotExist(err) {
				t.Errorf("Expected normalized file %s to be created, but it was not", expectedOutputPath)
			} else if err != nil {
				t.Errorf("Error checking for normalized file %s: %v", expectedOutputPath, err)
			}
		})
	}
}
