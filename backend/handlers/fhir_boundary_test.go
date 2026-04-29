package handlers

import "testing"

func TestValidateFHIRBundleSchema(t *testing.T) {
	ok := fhirBundle{
		ResourceType: "Bundle",
		Type:         "collection",
		Entry: []fhirEntry{
			{Resource: map[string]any{"resourceType": "Patient", "id": "p1", "identifier": []any{map[string]any{"value": "a@b.com"}}}},
		},
	}
	if err := validateFHIRBundleSchema(ok); err != nil {
		t.Fatalf("expected valid bundle: %v", err)
	}
	bad := fhirBundle{ResourceType: "Patient"}
	if err := validateFHIRBundleSchema(bad); err == nil {
		t.Fatal("expected invalid schema")
	}
}

func TestFHIRTransformationRoundTripSampleBundle(t *testing.T) {
	bundle := fhirBundle{
		ResourceType: "Bundle",
		Type:         "collection",
		Entry: []fhirEntry{
			{Resource: map[string]any{
				"resourceType": "Patient", "id": "11111111-1111-1111-1111-111111111111",
				"identifier": []any{map[string]any{"value": "patient@healthonyx.demo"}},
			}},
			{Resource: map[string]any{
				"resourceType": "Encounter", "id": "enc-1", "status": "planned",
				"subject":     map[string]any{"reference": "Patient/11111111-1111-1111-1111-111111111111"},
				"participant": []any{map[string]any{"individual": map[string]any{"reference": "Practitioner/22222222-2222-2222-2222-222222222222"}}},
				"reasonCode":  []any{map[string]any{"text": "Follow-up"}},
			}},
			{Resource: map[string]any{
				"resourceType": "Observation", "id": "obs-1",
				"subject":     map[string]any{"reference": "Patient/11111111-1111-1111-1111-111111111111"},
				"valueString": "CBC within normal range",
			}},
			{Resource: map[string]any{
				"resourceType": "MedicationRequest", "id": "med-1", "status": "active",
				"subject":                   map[string]any{"reference": "Patient/11111111-1111-1111-1111-111111111111"},
				"requester":                 map[string]any{"reference": "Practitioner/22222222-2222-2222-2222-222222222222"},
				"medicationCodeableConcept": map[string]any{"text": "Aspirin 75mg"},
			}},
		},
	}
	patients, encounters, observations, meds, err := parseBundleToCanonical(bundle)
	if err != nil {
		t.Fatalf("expected transform success, got %v", err)
	}
	if len(patients) != 1 || len(encounters) != 1 || len(observations) != 1 || len(meds) != 1 {
		t.Fatalf("unexpected transform counts: %d %d %d %d", len(patients), len(encounters), len(observations), len(meds))
	}
}
