package oras

import "testing"

func TestReferenceParseSimple(t *testing.T) {

	ref := "oras://example.io/mycontainer:latest"
	var settings orasSettings
	err := settings.parseReference(ref)
	if err != nil {
		t.Errorf("Error parsing reference: %s", err.Error())
	}
}

// Write a table driven test for the parseReference function
// validate the output of the registry, reference, and tag
func TestReferenceParse(t *testing.T) {
	tests := []struct {
		name              string
		ref               string
		wantErr           bool
		expectedRegistry  string
		expectedReference string
		expectedTag       string
	}{
		{
			name:              "valid reference",
			ref:               "oras://example.io/mycontainer:latest",
			wantErr:           false,
			expectedRegistry:  "example.io",
			expectedReference: "example.io/mycontainer",
			expectedTag:       "latest",
		},
		{
			name:              "no tag",
			ref:               "oras://example.io/mycontainer",
			expectedRegistry:  "example.io",
			expectedReference: "example.io/mycontainer",
			expectedTag:       "latest",
			wantErr:           false,
		},
		{
			name:              "bad reference",
			ref:               "oras://example.io/mycontainer:latest:latest",
			wantErr:           true,
			expectedRegistry:  "",
			expectedReference: "",
			expectedTag:       "",
		},
		{
			name:              "ubuntu no tag and no registry",
			ref:               "ubuntu",
			expectedRegistry:  "docker.io",
			expectedReference: "docker.io/library/ubuntu",
			expectedTag:       "latest",
			wantErr:           false,
		},
		{
			name:              "docker.io/library/ubuntu",
			ref:               "docker.io/library/ubuntu",
			expectedRegistry:  "docker.io",
			expectedReference: "docker.io/library/ubuntu",
			expectedTag:       "latest",
			wantErr:           false,
		},
		{
			name:              "localhost and port",
			ref:               "localhost:5001/artifact:latest",
			expectedRegistry:  "localhost:5001",
			expectedReference: "localhost:5001/artifact",
			expectedTag:       "latest",
			wantErr:           false,
		},
		{
			name:              "digest",
			ref:               "localhost:5001/artifact@sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867",
			expectedRegistry:  "localhost:5001",
			expectedReference: "localhost:5001/artifact",
			expectedTag:       "sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var settings orasSettings
			if err := settings.parseReference(tt.ref); (err != nil) != tt.wantErr {
				t.Errorf("parseReference() error = %v, wantErr %v", err, tt.wantErr)
			} else if !tt.wantErr {
				if settings.registry != tt.expectedRegistry {
					t.Errorf("parseReference() registry = %v, expectedRegistry %v", settings.registry, tt.expectedRegistry)
				}
				if settings.reference != tt.expectedReference {
					t.Errorf("parseReference() reference = %v, expectedReference %v", settings.reference, tt.expectedReference)
				}
				if settings.tag != tt.expectedTag {
					t.Errorf("parseReference() tag = %v, expectedTag %v", settings.tag, tt.expectedTag)
				}
			}
		})
	}
}
