package model

import "time"

type ArtifactKind string

const (
	KindSkill   ArtifactKind = "skill"
	KindCapsule ArtifactKind = "capsule"
)

type Signature struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"keyId"`
	PublicKey string `json:"publicKey"`
	Value     string `json:"value"`
}

type Artifact struct {
	Kind      ArtifactKind      `json:"kind"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Digest    string            `json:"digest"`
	OCIRef    string            `json:"ociRef"`
	CreatedAt time.Time         `json:"createdAt"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Signature *Signature        `json:"signature,omitempty"`
}

type ListFilter struct {
	Kind  ArtifactKind
	Name  string
	Limit int
}
