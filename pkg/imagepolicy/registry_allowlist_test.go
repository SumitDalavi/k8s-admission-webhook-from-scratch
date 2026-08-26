package imagepolicy

import "testing"

func TestIsAllowed_OfficialDocker(t *testing.T) {
	if !isAllowed("docker.io/library/nginx:1.25") {
		t.Error("expected docker.io/library/ to be allowed")
	}
}

func TestIsAllowed_GHCR(t *testing.T) {
	if !isAllowed("ghcr.io/org/myapp:v1.2.3") {
		t.Error("expected ghcr.io to be allowed")
	}
}

func TestIsAllowed_UnknownRegistry(t *testing.T) {
	if isAllowed("untrusted.registry.io/evil:latest") {
		t.Error("expected untrusted registry to be rejected")
	}
}

func TestIsAllowed_DockerHub_NonOfficial(t *testing.T) {
	// Only docker.io/library/ is in allowlist, not arbitrary docker.io
	if isAllowed("docker.io/someuser/image:tag") {
		t.Error("expected non-official dockerhub image to be rejected")
	}
}
