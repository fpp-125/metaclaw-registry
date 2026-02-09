package store

import "github.com/fpp-125/metaclaw-registry/internal/model"

type Store interface {
	Upsert(a model.Artifact) error
	Get(kind model.ArtifactKind, name, version string) (model.Artifact, bool)
	List(filter model.ListFilter) []model.Artifact
}
