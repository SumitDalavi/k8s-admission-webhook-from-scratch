package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"

	"github.com/your-username/k8s-admission-webhook-from-scratch/pkg/validate"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionv1.AddToScheme(runtimeScheme)
}

func serveMock(admit func(*admissionv1.AdmissionRequest, *corev1.Pod) *admissionv1.AdmissionResponse) http.HandlerFunc {
	// Simple wrapper resembling main.go
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var admissionReviewReq admissionv1.AdmissionReview
		_, _, _ = deserializer.Decode(body, nil, &admissionReviewReq)
		
		pod := corev1.Pod{}
		_, _, _ = deserializer.Decode(admissionReviewReq.Request.Object.Raw, nil, &pod)
		
		resp := admit(admissionReviewReq.Request, &pod)
		resp.UID = admissionReviewReq.Request.UID
		
		// Omit marshaling for test brevity
		if resp.Allowed {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func TestIntegrationWebhookValidRequest(t *testing.T) {
	ts := httptest.NewTLSServer(serveMock(validate.ValidatePod))
	defer ts.Close()

	client := ts.Client()
	
	validPayload := `{"apiVersion":"admission.k8s.io/v1","kind":"AdmissionReview","request":{"uid":"123","object":{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","labels":{"team":"security"}},"spec":{"containers":[{"name":"app","securityContext":{"runAsNonRoot":true}}]}}}}`
	
	req, _ := http.NewRequest("POST", ts.URL+"/validate", strings.NewReader(validPayload))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected OK for secure pod, got %v", resp.StatusCode)
	}
}

func TestIntegrationWebhookInvalidRequest(t *testing.T) {
	ts := httptest.NewTLSServer(serveMock(validate.ValidatePod))
	defer ts.Close()

	client := ts.Client()
	
	invalidPayload := `{"apiVersion":"admission.k8s.io/v1","kind":"AdmissionReview","request":{"uid":"123","object":{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","labels":{"secure":"false"}}}}}`
	
	req, _ := http.NewRequest("POST", ts.URL+"/validate", strings.NewReader(invalidPayload))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected Forbidden for insecure pod, got %v", resp.StatusCode)
	}
}
