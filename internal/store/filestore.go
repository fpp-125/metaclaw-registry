package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/metaclaw/metaclaw-registry/internal/model"
)

type FileStore struct {
	mu   sync.RWMutex
	path string
	data map[string]model.Artifact
}

func NewFileStore(path string) (*FileStore, error) {
	fs := &FileStore{path: path, data: map[string]model.Artifact{}}
	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

func key(kind model.ArtifactKind, name, version string) string {
	return string(kind) + ":" + strings.ToLower(strings.TrimSpace(name)) + ":" + strings.TrimSpace(version)
}

func (s *FileStore) Upsert(a model.Artifact) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key(a.Kind, a.Name, a.Version)] = a
	return s.persistLocked()
}

func (s *FileStore) Get(kind model.ArtifactKind, name, version string) (model.Artifact, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.data[key(kind, name, version)]
	return a, ok
}

func (s *FileStore) List(filter model.ListFilter) []model.Artifact {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Artifact, 0, len(s.data))
	nameQuery := strings.ToLower(strings.TrimSpace(filter.Name))
	for _, a := range s.data {
		if filter.Kind != "" && a.Kind != filter.Kind {
			continue
		}
		if nameQuery != "" && !strings.Contains(strings.ToLower(a.Name), nameQuery) {
			continue
		}
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			if out[i].Name == out[j].Name {
				return out[i].Version > out[j].Version
			}
			return out[i].Name < out[j].Name
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *FileStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path == "" {
		s.data = map[string]model.Artifact{}
		return nil
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.data = map[string]model.Artifact{}
			return nil
		}
		return err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		s.data = map[string]model.Artifact{}
		return nil
	}
	items := make([]model.Artifact, 0)
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	s.data = make(map[string]model.Artifact, len(items))
	for _, a := range items {
		s.data[key(a.Kind, a.Name, a.Version)] = a
	}
	return nil
}

func (s *FileStore) persistLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	items := make([]model.Artifact, 0, len(s.data))
	for _, a := range s.data {
		items = append(items, a)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Kind == items[j].Kind {
			if items[i].Name == items[j].Name {
				return items[i].Version < items[j].Version
			}
			return items[i].Name < items[j].Name
		}
		return items[i].Kind < items[j].Kind
	})
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(b, '\n'), 0o644)
}
