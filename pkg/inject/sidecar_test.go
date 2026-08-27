package inject

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func createReview(uid types.UID, pod corev1.Pod) []byte {
	podBytes, _ := json.Marshal(pod)
	req := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
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
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("{invalid json")))
	rr := httptest.NewRecorder()

	Handle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 bad request, got %d", rr.Code)
	}
}

func TestMutateInvalidPodJSON(t *testing.T) {
	ar := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "123",
			Object: runtime.RawExtension{
				Raw: []byte("invalid pod json"),
			},
		},
	}

	resp := mutate(ar)
	if resp.Allowed {
		t.Errorf("expected allowed=false on invalid pod json")
	}
}

func TestMutateNoAnnotation(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
		},
	}
	ar := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "123",
			Object: runtime.RawExtension{
				Raw: func() []byte { b, _ := json.Marshal(pod); return b }(),
			},
		},
	}

	resp := mutate(ar)
	if !resp.Allowed {
		t.Errorf("expected allowed=true")
	}
	if resp.Patch != nil {
		t.Errorf("expected nil patch")
	}
}

func TestMutateAlreadyInjected(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
			Annotations: map[string]string{
				injectAnnotation:   "true",
				injectedAnnotation: "injected",
			},
		},
	}
	ar := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "123",
			Object: runtime.RawExtension{
				Raw: func() []byte { b, _ := json.Marshal(pod); return b }(),
			},
		},
	}

	resp := mutate(ar)
	if !resp.Allowed {
		t.Errorf("expected allowed=true")
	}
	if resp.Patch != nil {
		t.Errorf("expected nil patch")
	}
}

func TestMutateInject(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
			Annotations: map[string]string{
				injectAnnotation: "true",
			},
		},
	}
	ar := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "123",
			Object: runtime.RawExtension{
				Raw: func() []byte { b, _ := json.Marshal(pod); return b }(),
			},
		},
	}

	resp := mutate(ar)
	if !resp.Allowed {
		t.Errorf("expected allowed=true")
	}
	if resp.Patch == nil {
		t.Errorf("expected non-nil patch")
	}
}

func TestHandleFullFlow(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod",
			Annotations: map[string]string{
				injectAnnotation: "true",
			},
		},
	}
	reqBody := createReview("456", pod)
	
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}

	var resp admissionv1.AdmissionReview
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Response == nil {
		t.Fatalf("response is nil")
	}
	if !resp.Response.Allowed {
		t.Errorf("expected allowed=true")
	}
	if resp.Response.Patch == nil {
		t.Errorf("expected patch to be generated")
	}
}
