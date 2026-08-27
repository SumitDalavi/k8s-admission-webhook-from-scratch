package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/your-username/k8s-admission-webhook-from-scratch/pkg/mutate"
	"github.com/your-username/k8s-admission-webhook-from-scratch/pkg/validate"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", serve(validate.ValidatePod))
	mux.HandleFunc("/mutate", serve(mutate.MutatePod))

	log.Println("Starting webhook server on port 8443...")
	err := http.ListenAndServeTLS(":8443", certFile, keyFile, mux)
	if err != nil {
		log.Printf("Failed to start server: %v", err)
		return 1
	}
	return 0
}

func main() {
	// Webhooks in k8s MUST be served over HTTPS. We assume certs are mounted at /etc/webhook/certs
	os.Exit(run("/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key"))
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
