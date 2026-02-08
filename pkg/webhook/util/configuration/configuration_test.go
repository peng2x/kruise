/*
Copyright 2025 The Kruise Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configuration

import (
	"reflect"
	"testing"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestAreSemanticEqualMutatingWebhooks_WebhookOrderChange verifies that AreSemanticEqualMutatingWebhooks
// returns true when only the order of webhooks changes.
func TestAreSemanticEqualMutatingWebhooks_WebhookOrderChange(t *testing.T) {
	sideEffectNone := admissionregistrationv1.SideEffectClassNone
	path1 := "/mutate-pod"
	path2 := "/mutate-deployment"

	webhook1 := admissionregistrationv1.MutatingWebhook{
		Name: "webhook-1",
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			Service: &admissionregistrationv1.ServiceReference{
				Name:      "webhook-service",
				Namespace: "kruise-system",
				Path:      &path1,
			},
			CABundle: []byte("ca-bundle-data"),
		},
		SideEffects: &sideEffectNone,
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"env": "prod"},
		},
	}

	webhook2 := admissionregistrationv1.MutatingWebhook{
		Name: "webhook-2",
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			Service: &admissionregistrationv1.ServiceReference{
				Name:      "webhook-service",
				Namespace: "kruise-system",
				Path:      &path2,
			},
			CABundle: []byte("ca-bundle-data"),
		},
		SideEffects: &sideEffectNone,
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"env": "staging"},
		},
	}

	config1 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{webhook1, webhook2},
	}

	config2 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{webhook2, webhook1}, // Reversed order
	}

	if !AreSemanticEqualMutatingWebhooks(config1, config2) {
		t.Error("AreSemanticEqualMutatingWebhooks should return true when only webhook order changes, but returned false. This would cause unnecessary webhook updates.")
	}

	// Verify that reflect.DeepEqual returns false (to confirm the difference in behavior)
	if reflect.DeepEqual(config1, config2) {
		t.Error("reflect.DeepEqual unexpectedly returned true for reordered webhooks")
	}
}

// TestAreSemanticEqualMutatingWebhooks_MatchExpressionsOrderChange verifies that AreSemanticEqualMutatingWebhooks
// returns true when only the order of MatchExpressions in NamespaceSelector changes.
func TestAreSemanticEqualMutatingWebhooks_MatchExpressionsOrderChange(t *testing.T) {
	sideEffectNone := admissionregistrationv1.SideEffectClassNone
	path := "/mutate-pod"

	expr1 := metav1.LabelSelectorRequirement{
		Key:      "control-plane",
		Operator: metav1.LabelSelectorOpDoesNotExist,
	}
	expr2 := metav1.LabelSelectorRequirement{
		Key:      "env",
		Operator: metav1.LabelSelectorOpIn,
		Values:   []string{"prod", "staging"},
	}
	expr3 := metav1.LabelSelectorRequirement{
		Key:      "region",
		Operator: metav1.LabelSelectorOpNotIn,
		Values:   []string{"eu-west"},
	}

	config1 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("ca-bundle-data"),
				},
				SideEffects: &sideEffectNone,
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{expr1, expr2, expr3},
				},
			},
		},
	}

	config2 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("ca-bundle-data"),
				},
				SideEffects: &sideEffectNone,
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{expr3, expr1, expr2}, // Reordered
				},
			},
		},
	}

	if !AreSemanticEqualMutatingWebhooks(config1, config2) {
		t.Error("AreSemanticEqualMutatingWebhooks should return true when only MatchExpressions order changes, but returned false. This would cause unnecessary webhook updates.")
	}

	// Verify that reflect.DeepEqual returns false (to confirm the difference in behavior)
	if reflect.DeepEqual(config1, config2) {
		t.Error("reflect.DeepEqual unexpectedly returned true for reordered MatchExpressions")
	}
}

// TestAreSemanticEqualValidatingWebhooks_WebhookOrderChange verifies that AreSemanticEqualValidatingWebhooks
// returns true when only the order of webhooks changes.
func TestAreSemanticEqualValidatingWebhooks_WebhookOrderChange(t *testing.T) {
	sideEffectNone := admissionregistrationv1.SideEffectClassNone
	path1 := "/validate-pod"
	path2 := "/validate-deployment"

	webhook1 := admissionregistrationv1.ValidatingWebhook{
		Name: "webhook-1",
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			Service: &admissionregistrationv1.ServiceReference{
				Name:      "webhook-service",
				Namespace: "kruise-system",
				Path:      &path1,
			},
			CABundle: []byte("ca-bundle-data"),
		},
		SideEffects: &sideEffectNone,
	}

	webhook2 := admissionregistrationv1.ValidatingWebhook{
		Name: "webhook-2",
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			Service: &admissionregistrationv1.ServiceReference{
				Name:      "webhook-service",
				Namespace: "kruise-system",
				Path:      &path2,
			},
			CABundle: []byte("ca-bundle-data"),
		},
		SideEffects: &sideEffectNone,
	}

	config1 := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{webhook1, webhook2},
	}

	config2 := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{webhook2, webhook1}, // Reversed order
	}

	if !AreSemanticEqualValidatingWebhooks(config1, config2) {
		t.Error("AreSemanticEqualValidatingWebhooks should return true when only webhook order changes, but returned false. This would cause unnecessary webhook updates.")
	}
}

// TestAreSemanticEqualMutatingWebhooks_ActualChangesDetected verifies that AreSemanticEqualMutatingWebhooks
// correctly returns false when there are actual semantic differences.
func TestAreSemanticEqualMutatingWebhooks_ActualChangesDetected(t *testing.T) {
	sideEffectNone := admissionregistrationv1.SideEffectClassNone
	path := "/mutate-pod"

	config1 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("old-ca-bundle"),
				},
				SideEffects: &sideEffectNone,
			},
		},
	}

	config2 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("new-ca-bundle"), // Different CABundle
				},
				SideEffects: &sideEffectNone,
			},
		},
	}

	if AreSemanticEqualMutatingWebhooks(config1, config2) {
		t.Error("AreSemanticEqualMutatingWebhooks should return false when CABundle changes, but returned true. This would skip necessary webhook updates.")
	}
}

// TestAreSemanticEqualMutatingWebhooks_MatchLabelsNilVsEmpty verifies that AreSemanticEqualMutatingWebhooks
// returns true when one config has nil MatchLabels and another has empty MatchLabels.
func TestAreSemanticEqualMutatingWebhooks_MatchLabelsNilVsEmpty(t *testing.T) {
	sideEffectNone := admissionregistrationv1.SideEffectClassNone
	path := "/mutate-pod"

	config1 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("ca-bundle-data"),
				},
				SideEffects: &sideEffectNone,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: nil, // nil MatchLabels
				},
			},
		},
	}

	config2 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("ca-bundle-data"),
				},
				SideEffects: &sideEffectNone,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{}, // empty MatchLabels
				},
			},
		},
	}

	if !AreSemanticEqualMutatingWebhooks(config1, config2) {
		t.Error("AreSemanticEqualMutatingWebhooks should return true when comparing nil vs empty MatchLabels, but returned false. This would cause unnecessary webhook updates.")
	}
}

// TestAreSemanticEqualMutatingWebhooks_NamespaceSelectorLabelChanges verifies that AreSemanticEqualMutatingWebhooks
// correctly returns false when NamespaceSelector labels actually change.
func TestAreSemanticEqualMutatingWebhooks_NamespaceSelectorLabelChanges(t *testing.T) {
	sideEffectNone := admissionregistrationv1.SideEffectClassNone
	path := "/mutate-pod"

	config1 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("ca-bundle-data"),
				},
				SideEffects: &sideEffectNone,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"env": "prod"},
				},
			},
		},
	}

	config2 := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("ca-bundle-data"),
				},
				SideEffects: &sideEffectNone,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"env": "staging"}, // Different value
				},
			},
		},
	}

	if AreSemanticEqualMutatingWebhooks(config1, config2) {
		t.Error("AreSemanticEqualMutatingWebhooks should return false when NamespaceSelector labels change, but returned true. This would skip necessary webhook updates.")
	}
}

// TestAreSemanticEqualMutatingWebhooks_NormalizationDoesNotMutate verifies that normalization
// doesn't mutate the configurations in a way that causes update loops.
func TestAreSemanticEqualMutatingWebhooks_NormalizationDoesNotMutate(t *testing.T) {
	sideEffectNone := admissionregistrationv1.SideEffectClassNone
	path := "/mutate-pod"

	// Create config with nil MatchLabels
	config := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-webhook-config",
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "test-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      "webhook-service",
						Namespace: "kruise-system",
						Path:      &path,
					},
					CABundle: []byte("ca-bundle-data"),
				},
				SideEffects: &sideEffectNone,
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: nil, // Explicitly nil
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "env",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"prod"},
						},
					},
				},
			},
		},
	}

	// Make a deep copy for comparison
	configCopy := config.DeepCopy()

	// Call AreSemanticEqualMutatingWebhooks - this should normalize both configs
	if !AreSemanticEqualMutatingWebhooks(config, configCopy) {
		t.Error("AreSemanticEqualMutatingWebhooks should return true for identical configs")
	}

	// Verify that the original configs are still deeply equal after normalization
	if !reflect.DeepEqual(config, configCopy) {
		t.Error("Normalization should not mutate the original configs. This causes infinite reconciliation loops!")
	}
}
