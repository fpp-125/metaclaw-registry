# metaclaw-registry

Publish/distribution backend for MetaClaw artifacts.

## Repository Boundary

This repo is the registry service backend only.

- includes: registry API, metadata index, signature/digest verification, auth, search/list/download policy
- excludes: engine compiler/runtime logic, concrete business bots, example projects

## Ecosystem Repo Map

| Repo | Primary Responsibility | URL |
| --- | --- | --- |
| `metaclaw` | Engine core: compiler, runtime adapters, lifecycle/state | https://github.com/fpp-125/metaclaw |
| `metaclaw-examples` | Runnable end-to-end examples and starter templates | https://github.com/fpp-125/metaclaw-examples |
| `metaclaw-skills` | Reusable capabilities (`SKILL.md` + `capability.contract`) | https://github.com/fpp-125/metaclaw-skills |
| `metaclaw-registry` | Publish/distribution backend for skill/capsule metadata | https://github.com/fpp-125/metaclaw-registry |

## Artifact Model

Registry records metadata for `skill` and `capsule` artifacts:

- `kind`, `name`, `version`
- `digest` (immutable content identity)
- `ociRef` (where artifact is stored, for example `ghcr.io/...`)
- optional signature metadata

Note: artifact binaries/images are stored in OCI registry; this repo stores service/API + metadata index.

## API

- `GET /healthz`
- `POST /v1/artifacts`
- `GET /v1/artifacts?kind=skill&name=obsidian.search`
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

## Related Repos

- engine: https://github.com/fpp-125/metaclaw
- skills: https://github.com/fpp-125/metaclaw-skills
- examples: https://github.com/fpp-125/metaclaw-examples
