package imagepolicy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AllowedRegistries defines registries that pods may pull images from.
var AllowedRegistries = []string{
	"gcr.io/",
	"ghcr.io/",
	"registry.k8s.io/",
	"quay.io/",
	"docker.io/library/",  // official images only
}

// Handle validates that all container images come from allowed registries.
func Handle(w http.ResponseWriter, r *http.Request) {
	body := make([]byte, 0, 4096)
	for {
		tmp := make([]byte, 512)
		n, err := r.Body.Read(tmp)
		body = append(body, tmp[:n]...)
		if err != nil { break }
	}

	ar := admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, &ar); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ar.Response = validate(&ar)
	json.NewEncoder(w).Encode(ar)
}

func validate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	pod := &corev1.Pod{}
	if err := json.Unmarshal(ar.Request.Object.Raw, pod); err != nil {
		return deny(ar.Request.UID, fmt.Sprintf("decode pod: %v", err))
	}

	for _, c := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
		if !isAllowed(c.Image) {
			return deny(ar.Request.UID,
				fmt.Sprintf("image '%s' is not from an allowed registry. Allowed: %v", c.Image, AllowedRegistries))
		}
	}
	return allow(ar.Request.UID)
}

func isAllowed(image string) bool {
	for _, prefix := range AllowedRegistries {
		if strings.HasPrefix(image, prefix) {
			return true
		}
	}
	return false
}

func allow(uid types.UID) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{UID: uid, Allowed: true}
}

func deny(uid types.UID, msg string) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		UID: uid, Allowed: false,
		Result: &metav1.Status{Message: msg},
	}
}
