package service

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/metaclaw/metaclaw-registry/internal/model"
	"github.com/metaclaw/metaclaw-registry/internal/store"
)

type Service struct {
	store store.Store
}

func New(s store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) Register(a model.Artifact) (model.Artifact, error) {
	if err := validateArtifact(a); err != nil {
		return model.Artifact{}, err
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	if a.Metadata == nil {
		a.Metadata = map[string]string{}
	}
	if a.Signature != nil {
		if err := verifySignature(a); err != nil {
			return model.Artifact{}, err
		}
	}
	if err := s.store.Upsert(a); err != nil {
		return model.Artifact{}, err
	}
	return a, nil
}

func (s *Service) List(filter model.ListFilter) []model.Artifact {
	return s.store.List(filter)
}

func (s *Service) Get(kind model.ArtifactKind, name, version string) (model.Artifact, bool) {
	return s.store.Get(kind, name, version)
}

func validateArtifact(a model.Artifact) error {
	switch a.Kind {
	case model.KindSkill, model.KindCapsule:
	default:
		return fmt.Errorf("invalid kind: %s", a.Kind)
	}
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	a.Version = strings.TrimSpace(a.Version)
	if a.Version == "" {
		return fmt.Errorf("version is required")
	}
	a.OCIRef = strings.TrimSpace(a.OCIRef)
	if a.OCIRef == "" {
		return fmt.Errorf("ociRef is required")
	}
	a.Digest = strings.TrimSpace(a.Digest)
	if !strings.HasPrefix(a.Digest, "sha256:") || len(strings.TrimPrefix(a.Digest, "sha256:")) != 64 {
		return fmt.Errorf("digest must be sha256:<64 hex>")
	}
	return nil
}

func verifySignature(a model.Artifact) error {
	sig := a.Signature
	if sig.Algorithm != "ed25519" {
		return fmt.Errorf("unsupported signature algorithm: %s", sig.Algorithm)
	}
	pubRaw, err := base64.StdEncoding.DecodeString(sig.PublicKey)
	if err != nil {
		return fmt.Errorf("decode signature publicKey: %w", err)
	}
	if len(pubRaw) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid ed25519 public key size: %d", len(pubRaw))
	}
	sigRaw, err := base64.StdEncoding.DecodeString(sig.Value)
	if err != nil {
		return fmt.Errorf("decode signature value: %w", err)
	}
	payload, err := signaturePayload(a)
	if err != nil {
		return err
	}
	if !ed25519.Verify(ed25519.PublicKey(pubRaw), payload, sigRaw) {
		return fmt.Errorf("invalid signature")
	}
	derived := deriveKeyID(ed25519.PublicKey(pubRaw))
	if sig.KeyID != "" && sig.KeyID != derived {
		return fmt.Errorf("signature keyId mismatch: expected %s got %s", derived, sig.KeyID)
	}
	return nil
}

func SignaturePayloadForTest(a model.Artifact) ([]byte, error) {
	return signaturePayload(a)
}

func signaturePayload(a model.Artifact) ([]byte, error) {
	meta := map[string]string{}
	for k, v := range a.Metadata {
		meta[k] = v
	}
	payload := map[string]any{
		"kind":     a.Kind,
		"name":     a.Name,
		"version":  a.Version,
		"ociRef":   a.OCIRef,
		"digest":   a.Digest,
		"metadata": meta,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal signature payload: %w", err)
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

func deriveKeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return "ed25519:" + hex.EncodeToString(sum[:8])
}
