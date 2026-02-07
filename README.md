# metaclaw-registry

Registry backend for publishing and distributing MetaClaw artifacts.

## Scope

- Stores metadata for `skill` and `capsule` artifacts.
- Tracks OCI reference + digest for each version.
- Verifies optional Ed25519 signatures on publish.
- Provides list/get APIs.
- Supports bearer-token auth for write APIs.

## API

- `GET /healthz`
- `POST /v1/artifacts`
- `GET /v1/artifacts?kind=skill&name=obsidian`
- `GET /v1/artifacts/{kind}/{name}/{version}`

## Run

```bash
go run ./cmd/metaclaw-registry --addr :8088 --data ./data/registry.json --admin-token dev-token
```

## Publish Example

```bash
curl -X POST http://localhost:8088/v1/artifacts \
  -H 'Authorization: Bearer dev-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "kind":"skill",
    "name":"obsidian.search",
    "version":"v1.0.0",
    "digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "ociRef":"ghcr.io/metaclaw/skills/obsidian.search:v1.0.0"
  }'
```

## Test

```bash
go test ./...
```
