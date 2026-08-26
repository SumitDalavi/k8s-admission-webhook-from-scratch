package inject

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecFactory  = serializer.NewCodecFactory(runtimeScheme)
)

// SidecarConfig defines what gets injected
type SidecarConfig struct {
	Containers []corev1.Container `json:"containers"`
	Volumes    []corev1.Volume    `json:"volumes"`
}

// DefaultSidecar is the observability sidecar injected by default
var DefaultSidecar = SidecarConfig{
	Containers: []corev1.Container{
		{
			Name:  "otel-sidecar",
			Image: "otel/opentelemetry-collector-contrib:latest",
			Args:  []string{"--config=/etc/otel/config.yaml"},
			Resources: corev1.ResourceRequirements{},
			VolumeMounts: []corev1.VolumeMount{
				{Name: "otel-config", MountPath: "/etc/otel"},
			},
		},
	},
	Volumes: []corev1.Volume{
		{
			Name: "otel-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "otel-config"},
				},
			},
		},
	},
}

const injectAnnotation = "sidecar-injector/inject"
const injectedAnnotation = "sidecar-injector/status"

// Handle processes mutating admission webhook requests for sidecar injection
func Handle(w http.ResponseWriter, r *http.Request) {
	body := readBody(r)
	ar := admissionv1.AdmissionReview{}
	if _, _, err := codecFactory.UniversalDeserializer().Decode(body, nil, &ar); err != nil {
		http.Error(w, fmt.Sprintf("decode error: %v", err), http.StatusBadRequest)
		return
	}

	response := mutate(&ar)
	ar.Response = response
	json.NewEncoder(w).Encode(ar)
}

func mutate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	pod := &corev1.Pod{}
	if err := json.Unmarshal(ar.Request.Object.Raw, pod); err != nil {
		return errResponse(ar.Request.UID, fmt.Sprintf("decode pod: %v", err))
	}

	// Only inject if annotation is present and not already injected
	if pod.Annotations[injectAnnotation] != "true" || pod.Annotations[injectedAnnotation] == "injected" {
		return &admissionv1.AdmissionResponse{UID: ar.Request.UID, Allowed: true}
	}

	patches := buildPatches(pod)
	patchBytes, _ := json.Marshal(patches)
	patchType := admissionv1.PatchTypeJSONPatch

	return &admissionv1.AdmissionResponse{
		UID: ar.Request.UID, Allowed: true,
		Patch: patchBytes, PatchType: &patchType,
	}
}

func buildPatches(pod *corev1.Pod) []map[string]interface{} {
	patches := []map[string]interface{}{}
	for _, c := range DefaultSidecar.Containers {
		patches = append(patches, map[string]interface{}{
			"op": "add", "path": "/spec/containers/-", "value": c,
		})
	}
	for _, v := range DefaultSidecar.Volumes {
		patches = append(patches, map[string]interface{}{
			"op": "add", "path": "/spec/volumes/-", "value": v,
		})
	}
	// Mark as injected
	patches = append(patches, map[string]interface{}{
		"op": "add",
		"path": "/metadata/annotations/sidecar-injector~1status",
		"value": "injected",
	})
	return patches
}

func errResponse(uid types.UID, msg string) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		UID: uid, Allowed: false,
		Result: &metav1.Status{Message: msg},
	}
}

func readBody(r *http.Request) []byte {
	buf := make([]byte, 0, 4096)
	for {
		tmp := make([]byte, 512)
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return buf
}
