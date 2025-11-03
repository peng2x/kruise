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
	"encoding/json"
	"reflect"
	"testing"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeNamespaceSelectors(t *testing.T) {
	tests := []struct {
		name      string
		selector1 *metav1.LabelSelector
		selector2 *metav1.LabelSelector
		expected  *metav1.LabelSelector
	}{
		{
			name:      "both nil",
			selector1: nil,
			selector2: nil,
			expected:  nil,
		},
		{
			name:      "first nil",
			selector1: nil,
			selector2: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			},
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			},
		},
		{
			name: "second nil",
			selector1: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			},
			selector2: nil,
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			},
		},
		{
			name: "merge match labels",
			selector1: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			},
			selector2: &metav1.LabelSelector{
				MatchLabels: map[string]string{"region": "us-west"},
			},
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env":    "prod",
					"region": "us-west",
				},
			},
		},
		{
			name: "merge match expressions",
			selector1: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "control-plane",
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
				},
			},
			selector2: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "env",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"prod", "staging"},
					},
				},
			},
			expected: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "control-plane",
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
					{
						Key:      "env",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"prod", "staging"},
					},
				},
			},
		},
		{
			name: "merge both labels and expressions",
			selector1: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "control-plane",
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
				},
			},
			selector2: &metav1.LabelSelector{
				MatchLabels: map[string]string{"region": "us-west"},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "tier",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"frontend", "backend"},
					},
				},
			},
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env":    "prod",
					"region": "us-west",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "control-plane",
						Operator: metav1.LabelSelectorOpDoesNotExist,
					},
					{
						Key:      "tier",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"frontend", "backend"},
					},
				},
			},
		},
		{
			name: "overlapping match labels - selector2 takes precedence",
			selector1: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			},
			selector2: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "staging"},
			},
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "staging"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeNamespaceSelectors(tt.selector1, tt.selector2)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mergeNamespaceSelectors() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParseMutatingTemplateWithNamespaceSelector(t *testing.T) {
	// Create a webhook template with a namespace selector
	templateWebhook := admissionregistrationv1.MutatingWebhook{
		Name: "test-webhook",
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"env": "prod"},
		},
	}
	templateBytes, _ := json.Marshal([]admissionregistrationv1.MutatingWebhook{templateWebhook})

	// Create a current webhook configuration with a different namespace selector
	currentWebhook := admissionregistrationv1.MutatingWebhook{
		Name: "test-webhook",
		NamespaceSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "control-plane",
					Operator: metav1.LabelSelectorOpDoesNotExist,
				},
			},
		},
	}

	mutatingConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"template": string(templateBytes),
			},
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{currentWebhook},
	}

	result, err := parseMutatingTemplate(mutatingConfig)
	if err != nil {
		t.Fatalf("parseMutatingTemplate failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 webhook, got %d", len(result))
	}

	webhook := result[0]
	expectedSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{"env": "prod"},
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "control-plane",
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
		},
	}

	if !reflect.DeepEqual(webhook.NamespaceSelector, expectedSelector) {
		t.Errorf("NamespaceSelector not merged correctly. Expected %v, got %v", expectedSelector, webhook.NamespaceSelector)
	}
}

func TestParseValidatingTemplateWithNamespaceSelector(t *testing.T) {
	// Create a webhook template with a namespace selector
	templateWebhook := admissionregistrationv1.ValidatingWebhook{
		Name: "test-webhook",
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"region": "us-west"},
		},
	}
	templateBytes, _ := json.Marshal([]admissionregistrationv1.ValidatingWebhook{templateWebhook})

	// Create a current webhook configuration with a different namespace selector
	currentWebhook := admissionregistrationv1.ValidatingWebhook{
		Name: "test-webhook",
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"tier": "frontend"},
		},
	}

	validatingConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"template": string(templateBytes),
			},
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{currentWebhook},
	}

	result, err := parseValidatingTemplate(validatingConfig)
	if err != nil {
		t.Fatalf("parseValidatingTemplate failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 webhook, got %d", len(result))
	}

	webhook := result[0]
	expectedSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"region": "us-west",  // From template
			"tier":   "frontend", // From current config (takes precedence for overlapping keys)
		},
	}

	if !reflect.DeepEqual(webhook.NamespaceSelector, expectedSelector) {
		t.Errorf("NamespaceSelector not merged correctly. Expected %v, got %v", expectedSelector, webhook.NamespaceSelector)
	}
}
