package validate

import (
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidatePod checks if the pod creation is allowed based on security and governance rules.
func ValidatePod(req *admissionv1.AdmissionRequest, pod *corev1.Pod) *admissionv1.AdmissionResponse {
	// Rule 1: Required team label
	if _, ok := pod.Labels["team"]; !ok {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Pod is missing the required 'team' label.",
			},
		}
	}

	// Iterate over all containers (including init containers) to check security contexts
	containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
	for _, c := range containers {
		// Rule 2: No containers running as root
		if c.SecurityContext == nil || c.SecurityContext.RunAsNonRoot == nil || !*c.SecurityContext.RunAsNonRoot {
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: fmt.Sprintf("Container '%s' must set securityContext.runAsNonRoot to true.", c.Name),
				},
			}
		}

		// Rule 3: No privileged containers
		if c.SecurityContext != nil && c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged {
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: fmt.Sprintf("Container '%s' is not allowed to run as privileged.", c.Name),
				},
			}
		}
	}

	// All checks passed
	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}
