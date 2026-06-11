package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const projectFileDir = ".tfplan-pretty-output"
const projectFileName = "project.json"

type projectFile struct {
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

func GetOrCreateProjectID(workdir string) (string, error) {
	path := filepath.Join(workdir, projectFileDir, projectFileName)
	data, err := os.ReadFile(path)
	if err == nil {
		var pf projectFile
		if jerr := json.Unmarshal(data, &pf); jerr == nil && pf.ProjectID != "" {
			return pf.ProjectID, nil
		}
		
	}

	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("could not read %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("could not create %s: %w", filepath.Dir(path), err)
	}

	id, err := newRandomID(12) 
	if err != nil {
		return "", err
	}
	pf := projectFile{ProjectID: id, CreatedAt: time.Now().UTC()}
	out, _ := json.MarshalIndent(pf, "", "  ")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return "", fmt.Errorf("could not write %s: %w", path, err)
	}
	return id, nil
}

func newRandomID(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("could not generate random id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
