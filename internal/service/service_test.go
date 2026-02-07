package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/metaclaw/metaclaw-registry/internal/model"
	"github.com/metaclaw/metaclaw-registry/internal/store"
)

type memStore struct {
	items map[string]model.Artifact
}

func newMemStore() *memStore {
	return &memStore{items: map[string]model.Artifact{}}
}

func mk(kind model.ArtifactKind, name, version string) string {
	return string(kind) + ":" + name + ":" + version
}

func (m *memStore) Upsert(a model.Artifact) error {
	m.items[mk(a.Kind, a.Name, a.Version)] = a
	return nil
}

func (m *memStore) Get(kind model.ArtifactKind, name, version string) (model.Artifact, bool) {
	a, ok := m.items[mk(kind, name, version)]
	return a, ok
}

func (m *memStore) List(filter model.ListFilter) []model.Artifact {
	out := make([]model.Artifact, 0, len(m.items))
	for _, v := range m.items {
		out = append(out, v)
	}
	return out
}

func TestRegisterValidArtifact(t *testing.T) {
	svc := New(newMemStore())
	_, err := svc.Register(model.Artifact{
		Kind:    model.KindSkill,
		Name:    "obsidian.search",
		Version: "v1.0.0",
		Digest:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		OCIRef:  "ghcr.io/metaclaw/skills/obsidian.search:v1.0.0",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
}

func TestRegisterRejectsBadDigest(t *testing.T) {
	svc := New(newMemStore())
	_, err := svc.Register(model.Artifact{
		Kind:    model.KindSkill,
		Name:    "obsidian.search",
		Version: "v1.0.0",
		Digest:  "abc",
		OCIRef:  "ghcr.io/x/y:v1",
	})
	if err == nil {
		t.Fatal("expected digest validation error")
	}
}

func TestRegisterVerifiesEd25519Signature(t *testing.T) {
	svc := New(newMemStore())
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	a := model.Artifact{
		Kind:    model.KindCapsule,
		Name:    "obsidian-bot",
		Version: "v0.1.0",
		Digest:  "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		OCIRef:  "ghcr.io/metaclaw/capsules/obsidian-bot:v0.1.0",
		Metadata: map[string]string{
			"runtime": "apple_container",
		},
	}
	payload, err := SignaturePayloadForTest(a)
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	sig := ed25519.Sign(priv, payload)
	a.Signature = &model.Signature{
		Algorithm: "ed25519",
		PublicKey: base64.StdEncoding.EncodeToString(pub),
		Value:     base64.StdEncoding.EncodeToString(sig),
	}
	if _, err := svc.Register(a); err != nil {
		t.Fatalf("register signed artifact: %v", err)
	}
	a.Signature.Value = base64.StdEncoding.EncodeToString([]byte("tamper"))
	if _, err := svc.Register(a); err == nil {
		t.Fatal("expected signature verification failure")
	}
}

var _ store.Store = (*memStore)(nil)
