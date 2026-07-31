# Terrain — JSON Schemas

Canonical JSON Schema definitions for stable public-API artifacts. Each schema is versioned in its filename; new versions adopt one-cycle deprecation per the versioning section in `docs/PRODUCT.md`.

## Schemas

| File | What it specifies | Stable from |
|---|---|---|
| `finding.v1.json` | The `Finding` object emitted by Terrain (the structured artifact behind all four diagnostic renderers + `findings.json`). See the diagnostic-format section in `docs/PRODUCT.md`. | v0.2.0 |
| `terrain.yaml.v1.json` | The adopter-facing `terrain.yaml` configuration schema. See the configuration section in `docs/PRODUCT.md`. | v0.2.0 |

## Validation

Implementations should validate emitted `Finding` objects against `finding.v1.json` in tests. Validation against `terrain.yaml.v1.json` happens at `terrain init` / `terrain test` parse time; invalid `terrain.yaml` files are reported with a structured error pointing at the offending key.

## Adding a new schema

1. RFC required if the schema becomes part of the public API surface
2. File goes in `schemas/<name>.v1.json` with `$id` set to `https://raw.githubusercontent.com/pmclSF/terrain/main/schemas/<name>.v1.json`
3. Update this README
4. Update `docs/PRODUCT.md` to reference the schema as the canonical contract

## Versioning policy

Schemas follow semver-like versioning:

- Additive changes (new optional fields) — same version; documented in `CHANGELOG.md`
- Breaking changes (renamed fields, type changes, removed fields, new required fields) — new version with one-cycle deprecation
- Old version files are kept for the deprecation cycle, then removed

Consumers should pin to the major version (`finding.v1.json`) and tolerate additive changes within it.
