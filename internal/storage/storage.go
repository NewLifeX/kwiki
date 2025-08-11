package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/stcn52/kwiki/pkg/models"
)

// Storage interface defines the storage operations
type Storage interface {
	SaveWiki(wiki *models.Wiki) error
	LoadWiki(id string) (*models.Wiki, error)
	LoadAllWikis() (map[string]*models.Wiki, error)
	DeleteWiki(id string) error
	SaveLogs(wikiID string, logs []string) error
	LoadLogs(wikiID string) ([]string, error)
}

// FileStorage implements Storage interface using JSON files
type FileStorage struct {
	dataDir string
	mutex   sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(dataDir string) (*FileStorage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create subdirectories
	wikisDir := filepath.Join(dataDir, "wikis")
	logsDir := filepath.Join(dataDir, "logs")

	if err := os.MkdirAll(wikisDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create wikis directory: %w", err)
	}

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
	}, nil
}

// SaveWiki saves a wiki to disk
func (fs *FileStorage) SaveWiki(wiki *models.Wiki) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filePath := filepath.Join(fs.dataDir, "wikis", fmt.Sprintf("%s.json", wiki.ID))

	data, err := json.MarshalIndent(wiki, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wiki: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write wiki file: %w", err)
	}

	return nil
}

// LoadWiki loads a wiki from disk
func (fs *FileStorage) LoadWiki(id string) (*models.Wiki, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filePath := filepath.Join(fs.dataDir, "wikis", fmt.Sprintf("%s.json", id))

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Wiki not found
		}
		return nil, fmt.Errorf("failed to read wiki file: %w", err)
	}

	var wiki models.Wiki
	if err := json.Unmarshal(data, &wiki); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wiki: %w", err)
	}

	return &wiki, nil
}

// LoadAllWikis loads all wikis from disk
func (fs *FileStorage) LoadAllWikis() (map[string]*models.Wiki, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	wikisDir := filepath.Join(fs.dataDir, "wikis")
	wikis := make(map[string]*models.Wiki)

	files, err := ioutil.ReadDir(wikisDir)
	if err != nil {
		return wikis, nil // Return empty map if directory doesn't exist
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			wikiID := strings.TrimSuffix(file.Name(), ".json")
			wiki, err := fs.LoadWiki(wikiID)
			if err != nil {
				log.Printf("Failed to load wiki %s: %v", wikiID, err)
				continue
			}
			if wiki != nil {
				wikis[wikiID] = wiki
			}
		}
	}

	return wikis, nil
}

// DeleteWiki deletes a wiki from disk
func (fs *FileStorage) DeleteWiki(id string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filePath := filepath.Join(fs.dataDir, "wikis", fmt.Sprintf("%s.json", id))

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete wiki file: %w", err)
	}

	// Also delete logs
	logsPath := filepath.Join(fs.dataDir, "logs", fmt.Sprintf("%s.json", id))
	if err := os.Remove(logsPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to delete logs for wiki %s: %v", id, err)
	}

	return nil
}

// SaveLogs saves logs for a wiki
func (fs *FileStorage) SaveLogs(wikiID string, logs []string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filePath := filepath.Join(fs.dataDir, "logs", fmt.Sprintf("%s.json", wikiID))

	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write logs file: %w", err)
	}

	return nil
}

// LoadLogs loads logs for a wiki
func (fs *FileStorage) LoadLogs(wikiID string) ([]string, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filePath := filepath.Join(fs.dataDir, "logs", fmt.Sprintf("%s.json", wikiID))

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // No logs found
		}
		return nil, fmt.Errorf("failed to read logs file: %w", err)
	}

	var logs []string
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal logs: %w", err)
	}

	return logs, nil
}
