package validate

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestValidatePod(t *testing.T) {
	tests := []struct {
		name    string
		pod     *corev1.Pod
		allowed bool
	}{
		{
			name: "Valid pod",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"team": "security"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot: ptr.To(true),
								Privileged:   ptr.To(false),
							},
						},
					},
				},
			},
			allowed: true,
		},
		{
			name: "Missing team label",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"env": "prod"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot: ptr.To(true),
							},
						},
					},
				},
			},
			allowed: false,
		},
		{
			name: "Root container allowed",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"team": "security"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot: ptr.To(false),
							},
						},
					},
				},
			},
			allowed: false,
		},
		{
			name: "Privileged container",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"team": "security"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot: ptr.To(true),
								Privileged:   ptr.To(true),
							},
						},
					},
				},
			},
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ValidatePod(nil, tt.pod)
			if resp.Allowed != tt.allowed {
				t.Errorf("expected allowed %v, got %v: %s", tt.allowed, resp.Allowed, resp.Result.Message)
			}
		})
	}
}
