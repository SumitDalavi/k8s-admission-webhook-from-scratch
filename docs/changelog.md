# Changelog

## [v0.1.0] — 2026-10-18
### Added
- Automated GitHub Actions release pipeline triggered on `v*.*.*` git tags (`.github/workflows/release.yml`).
- Container image packaging (`ghcr.io/sumitdalavi/k8s-admission-webhook-from-scratch`).
- Vulnerability scanning step via Aqua Security Trivy (`aquasecurity/trivy-action@v0.36.0`).
- SPDX Software Bill of Materials (SBOM) generation step (`anchore/sbom-action@v0`).
- Automated GitHub release publishing attaching `sbom.spdx.json` via `softprops/action-gh-release@v2`.

## [2026-08-29] — Phase 2 Evidence
### Added
- Created `demo/run_webhook_demo.sh` script to verify cert rotation and failure-policy testing on `kind`.
- Updated `ARCHITECTURE.md` with complete Mermaid diagram and component/dependency tables.
- Standardized documentation (`runbook.md`, `decisions.md`).
