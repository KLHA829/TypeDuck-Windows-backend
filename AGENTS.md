# AGENTS.md

Backend-only agent contract. Product-wide rules, cross-repo protocol, and installer context live in the shared `TypeDuck-HK/TypeDuck-Windows` repo file `AGENTS.md` — read that before changes that touch protobuf, profile GUID, or runtime layout.

## Tests

Default Linux/CI (no native Rime):

```bash
go test ./agenttrust ./imecore ./input_methods/moqi ./input_methods/fcitx5 ./input_methods/rime/cloudclipboard
```

Windows / full backend (from repo root):

```powershell
go test ./...
go test ./imecore ./input_methods/rime ./mobilebridge
pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/Test-TypeDuckCandidateParity.ps1 -RepoRoot .
pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/Test-TypeDuckSettingsCustomization.ps1 -RepoRoot .
```

Native Rime tests: set `MOQI_RIME_PACKAGE_DIR`, `MOQI_RIME_INIT_MAX_MS`, or `MOQI_REAL_APPDATA` only when a test file documents that requirement.

## Non-goals

- No TSF, launcher, installer, or settings UI work in this repository.
- No broad Moqi→TypeDuck identifier sweeps (module path, protobuf package names, COM names) without a coordinated two-repo migration.
- No re-enabling TypeDuck v1 removed backend commands (legacy appearance, cloud clipboard, scheme download, etc.); keep `input_methods/rime/removed_surfaces_test.go` authoritative.
- No hand-editing `proto/moqi.pb.go`; change `proto/moqi.proto` and regenerate bindings with the frontend schema kept aligned.
- No product-feature churn framed as “cleanup” (renames, dead-code resurrection, logging of user content).

## Landmines

- Shipped runtime layout is **`TypeDuckRuntime`** (`scripts/build/TypeDuckRuntime/` from `scripts/build.ps1`). Some workflows and legacy scripts still say `moqi-ime` — treat that as stale packaging, not the product path.
- Go module remains **`github.com/gaboolic/moqi-ime`**; import paths are not the public product name.
- **`server.exe`** speaks length-prefixed protobuf on stdin/stdout; keep framing compatible with `TypeDuck-Windows:proto/ProtoFraming.h`.
- Rime is split across **`native_cgo.go`** (Windows) and **`native_stub.go`** (other GOOS). Root `go test ./...` on Linux may fail to build until stub and production hooks match.
- TypeDuck user settings flow: launcher → `METHOD_TYPEDUCK_SETTINGS_UPDATE` → `input_methods/rime/appearance_config.go` and redeploy paths — not backend menu UI.
- Do not log typed text, candidate bodies, or raw settings values in routine responses.

## Copy-trap (do not propagate)

Do **not** copy the invalid command below into README, workflows, or scripts. CI fails if this exact string appears outside this file:

`go test -tags=typeduck_enable_all_legacy_moqi ./...`
