package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/SumitDalavi/k8s-admission-webhook-from-scratch/pkg/mutate"
	"github.com/SumitDalavi/k8s-admission-webhook-from-scratch/pkg/validate"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func run(certFile, keyFile string) int {
	// Validate initial cert/key pair (fail-fast on invalid paths as expected by tests)
	initialCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Printf("Failed to load initial certificates: %v", err)
		return 1
	}

	reloader := newCertReloader(certFile, keyFile, initialCert)
	reloader.watch()

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", serve(validate.ValidatePod))
	mux.HandleFunc("/mutate", serve(mutate.MutatePod))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","mode":"tls_reloader"}`))
	})

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: reloader.GetCertificate,
		},
	}

	log.Println("Starting webhook server on port 8443 with dynamic TLS cert reloader...")
	err = server.ListenAndServeTLS("", "")
	if err != nil && err != http.ErrServerClosed {
		log.Printf("Failed to start server: %v", err)
		return 1
	}
	return 0
}

func main() {
	certFile := os.Getenv("TLS_CERT_PATH")
	if certFile == "" {
		if certDir := os.Getenv("CERT_DIR"); certDir != "" {
			certFile = certDir + "/tls.crt"
		} else {
			certFile = "/etc/webhook/certs/tls.crt"
		}
	}
	keyFile := os.Getenv("TLS_KEY_PATH")
	if keyFile == "" {
		if certDir := os.Getenv("CERT_DIR"); certDir != "" {
			keyFile = certDir + "/tls.key"
		} else {
			keyFile = "/etc/webhook/certs/tls.key"
		}
	}
	os.Exit(run(certFile, keyFile))
}

type admitFunc func(*admissionv1.AdmissionRequest, *corev1.Pod) *admissionv1.AdmissionResponse

func serve(admit admitFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			if data, err := io.ReadAll(r.Body); err == nil {
				body = data
			}
		}
		if len(body) == 0 {
			http.Error(w, "empty body", http.StatusBadRequest)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			http.Error(w, "invalid Content-Type, expected `application/json`", http.StatusUnsupportedMediaType)
			return
		}

		var admissionReviewReq admissionv1.AdmissionReview
		if _, _, err := deserializer.Decode(body, nil, &admissionReviewReq); err != nil {
			http.Error(w, fmt.Sprintf("could not decode body: %v", err), http.StatusBadRequest)
			return
		}

		if admissionReviewReq.Request == nil {
			http.Error(w, "admission review request is nil", http.StatusBadRequest)
			return
		}

		var pod corev1.Pod
		if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod); err != nil {
			http.Error(w, fmt.Sprintf("could not unmarshal pod: %v", err), http.StatusBadRequest)
			return
		}

		admissionResponse := admit(admissionReviewReq.Request, &pod)
		admissionResponse.UID = admissionReviewReq.Request.UID

		admissionReviewResp := admissionv1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
			Response: admissionResponse,
		}

		respBytes, err := json.Marshal(admissionReviewResp)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			log.Printf("could not write response: %v", err)
		}
	}
}
