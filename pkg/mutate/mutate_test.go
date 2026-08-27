package mutate

import (
	"encoding/json"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMutatePodNilLabels(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{}},
		},
	}
	resp := MutatePod(nil, pod)
	if !resp.Allowed {
		t.Fatal("expected allowed")
	}

	patchStr := string(resp.Patch)
	if !strings.Contains(patchStr, "/metadata/labels") {
		t.Errorf("expected patch to add labels")
	}
	if !strings.Contains(patchStr, "/spec/containers/0/resources/limits") {
		t.Errorf("expected patch to add resource limits")
	}
}

func TestMutatePodExistingLabelsWithoutCostCenter(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "foo"},
		},
	}
	resp := MutatePod(nil, pod)
	if !resp.Allowed {
		t.Fatal("expected allowed")
	}

	patchStr := string(resp.Patch)
	if !strings.Contains(patchStr, "/metadata/labels/cost-center") {
		t.Errorf("expected patch to add cost-center label")
	}
}

func TestMutatePodExistingCostCenterAndResources(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"cost-center": "existing"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("100m"),
						},
					},
				},
			},
		},
	}
	resp := MutatePod(nil, pod)
	if !resp.Allowed {
		t.Fatal("expected allowed")
	}

	var patches []patchOperation
	json.Unmarshal(resp.Patch, &patches)
	if len(patches) != 0 {
		t.Errorf("expected no patches, got %d", len(patches))
	}
}
