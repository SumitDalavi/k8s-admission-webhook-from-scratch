package mutate

import (
	"encoding/json"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// patchOperation represents a JSON patch operation
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// MutatePod injects default resource limits and cost attribution labels.
func MutatePod(req *admissionv1.AdmissionRequest, pod *corev1.Pod) *admissionv1.AdmissionResponse {
	var patches []patchOperation

	// 1. Add standard cost attribution label if missing
	if pod.Labels == nil {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/labels",
			Value: map[string]string{"cost-center": "default-engineering"},
		})
	} else if _, ok := pod.Labels["cost-center"]; !ok {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/labels/cost-center",
			Value: "default-engineering",
		})
	}

	// 2. Inject default resource limits if missing
	for i, c := range pod.Spec.Containers {
		if c.Resources.Limits == nil {
			patches = append(patches, patchOperation{
				Op:   "add",
				Path: "/spec/containers/" + string(rune(i+'0')) + "/resources/limits",
				Value: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			})
		}
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  nil,
		}
	}

	pt := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &pt,
	}
}
