package mutate

import (
	"encoding/json"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestMutatePod(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app1"}, // No resources
			},
		},
	}

	resp := MutatePod(nil, pod)
	if !resp.Allowed {
		t.Fatalf("expected allowed to be true")
	}

	if len(resp.Patch) == 0 {
		t.Fatalf("expected patch to be generated")
	}

	var patches []patchOperation
	err := json.Unmarshal(resp.Patch, &patches)
	if err != nil {
		t.Fatalf("failed to unmarshal patch: %v", err)
	}

	hasLabelPatch := false
	hasResourcePatch := false

	for _, p := range patches {
		if strings.Contains(p.Path, "/metadata/labels") && p.Value == "default-engineering" || p.Value.(map[string]interface{})["cost-center"] == "default-engineering" {
			hasLabelPatch = true
		}
		if strings.Contains(p.Path, "/resources/limits") {
			hasResourcePatch = true
		}
	}

	if !hasLabelPatch {
		t.Errorf("expected patch to include cost-center label")
	}
	if !hasResourcePatch {
		t.Errorf("expected patch to include resource limits")
	}
}
