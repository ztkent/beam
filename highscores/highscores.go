package highscores

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
)

const (
	defaultHighScoresFile  = "highscores.csv"
	defaultMaxHighScores   = 15
	defaultMaxStoredScores = 50
)

// HighScore represents a single score entry with its metadata
type HighScore struct {
	Score    int
	Duration float32
	Date     string
}

// HighScoreManager handles the loading, saving and manipulation of high scores
type HighScoreManager struct {
	Scores         []HighScore
	filePath       string
	maxHighScores  int
	maxStoredScore int
}

// NewHighScoreManager creates and initializes a new high score manager
func NewHighScoreManager(filePath string, maxHighScores, maxStoredScores int) *HighScoreManager {
	if filePath == "" {
		filePath = defaultHighScoresFile
	}
	if maxHighScores <= 0 {
		maxHighScores = defaultMaxHighScores
	}
	if maxStoredScores <= 0 {
		maxStoredScores = defaultMaxStoredScores
	}

	manager := &HighScoreManager{
		Scores:         make([]HighScore, 0),
		filePath:       filePath,
		maxHighScores:  maxHighScores,
		maxStoredScore: maxStoredScores,
	}
	if err := manager.Load(); err != nil {
		fmt.Println("Failed to load high scores:", err)
	}
	return manager
}

// Load reads the high scores from disk
func (m *HighScoreManager) Load() error {
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to open high scores file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read high scores: %w", err)
	}

	m.Scores = make([]HighScore, 0)
	for _, record := range records {
		if len(record) != 3 {
			continue
		}
		score, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		duration, err := strconv.ParseFloat(record[1], 32)
		if err != nil {
			continue
		}
		m.Scores = append(m.Scores, HighScore{
			Score:    score,
			Duration: float32(duration),
			Date:     record[2],
		})
	}

	m.sort()
	return nil
}

// Save writes the current high scores to disk
func (m *HighScoreManager) Save() error {
	file, err := os.Create(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to create high scores file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	m.sort()
	limit := min(len(m.Scores), m.maxStoredScore)
	for i := 0; i < limit; i++ {
		record := []string{
			strconv.Itoa(m.Scores[i].Score),
			fmt.Sprintf("%.1f", m.Scores[i].Duration),
			m.Scores[i].Date,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write score record: %w", err)
		}
	}

	return nil
}

// IsHighScore checks if the given score qualifies as a high score
func (m *HighScoreManager) IsHighScore(score int) bool {
	if len(m.Scores) < m.maxHighScores {
		return true
	}
	return score > m.Scores[len(m.Scores)-1].Score
}

// AddScore adds a new high score to the manager
func (m *HighScoreManager) AddScore(score HighScore) {
	m.Scores = append(m.Scores, score)
	m.sort()
}

// GetScores returns a copy of the current high scores
func (m *HighScoreManager) GetScores() []HighScore {
	scores := make([]HighScore, len(m.Scores))
	copy(scores, m.Scores)
	return scores
}

// sort orders the scores by score (descending) and duration (ascending)
func (m *HighScoreManager) sort() {
	sort.Slice(m.Scores, func(i, j int) bool {
		if m.Scores[i].Score == m.Scores[j].Score {
			return m.Scores[i].Duration < m.Scores[j].Duration
		}
		return m.Scores[i].Score > m.Scores[j].Score
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
