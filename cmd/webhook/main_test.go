package main

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
)

func mockAdmitFunc(req *admissionv1.AdmissionRequest, pod *corev1.Pod) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{Allowed: true}
}

func TestRunBadCerts(t *testing.T) {
	if code := run("invalid", "invalid"); code != 1 {
		t.Errorf("Expected exit code 1 for invalid certs, got %d", code)
	}
}

func TestServeEmptyBody(t *testing.T) {
	handler := serve(mockAdmitFunc)
	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", rr.Code)
	}
}

func TestServeInvalidContentType(t *testing.T) {
	handler := serve(mockAdmitFunc)
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("foo")))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415 for invalid content type, got %d", rr.Code)
	}
}

func TestServeInvalidJSON(t *testing.T) {
	handler := serve(mockAdmitFunc)
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid json, got %d", rr.Code)
	}
}

func TestServeNilRequest(t *testing.T) {
	handler := serve(mockAdmitFunc)

	review := admissionv1.AdmissionReview{} // Request is nil
	b, _ := json.Marshal(review)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for nil request, got %d", rr.Code)
	}
}

func TestServeInvalidPodJSON(t *testing.T) {
	handler := serve(mockAdmitFunc)

	review := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{Raw: []byte("invalid pod json")},
		},
	}
	b, _ := json.Marshal(review)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid pod json, got %d", rr.Code)
	}
}

func TestServeSuccess(t *testing.T) {
	handler := serve(mockAdmitFunc)

	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}}
	podBytes, _ := json.Marshal(pod)

	review := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:    "1234",
			Object: runtime.RawExtension{Raw: podBytes},
		},
	}
	b, _ := json.Marshal(review)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for valid request, got %d", rr.Code)
	}

	var resp admissionv1.AdmissionReview
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if resp.Response == nil {
		t.Fatal("response is nil")
	}

	if !resp.Response.Allowed {
		t.Errorf("expected allowed=true")
	}
	if string(resp.Response.UID) != "1234" {
		t.Errorf("expected UID 1234, got %s", resp.Response.UID)
	}
}
