package imagepolicy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func createReview(uid types.UID, pod corev1.Pod) []byte {
	podBytes, _ := json.Marshal(pod)
	req := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: uid,
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}
	b, _ := json.Marshal(req)
	return b
}

func TestHandleInvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer([]byte("{invalid json")))
	rr := httptest.NewRecorder()

	Handle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 bad request, got %d", rr.Code)
	}
}

func TestHandleValidAllowed(t *testing.T) {
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Image: "gcr.io/my-project/my-image"}},
		},
	}
	reqBody := createReview("123", pod)
	
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}

	var resp admissionv1.AdmissionReview
	json.NewDecoder(rr.Body).Decode(&resp)

	if !resp.Response.Allowed {
		t.Errorf("expected allowed=true")
	}
}

func TestHandleValidDenied(t *testing.T) {
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Image: "docker.io/malicious/image"}}, // malicious, not library
			Containers: []corev1.Container{{Image: "gcr.io/ok/image"}},
		},
	}
	reqBody := createReview("124", pod)
	
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	Handle(rr, req)

	var resp admissionv1.AdmissionReview
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Response.Allowed {
		t.Errorf("expected allowed=false")
	}
	if resp.Response.UID != "124" {
		t.Errorf("expected uid 124")
	}
}

func TestValidateInvalidPodJSON(t *testing.T) {
	ar := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "125",
			Object: runtime.RawExtension{
				Raw: []byte("invalid pod json"),
			},
		},
	}

	resp := validate(ar)
	if resp.Allowed {
		t.Errorf("expected allowed=false on invalid pod json")
	}
}
