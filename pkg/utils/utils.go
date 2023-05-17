package utils

import (
	"encoding/json"
	"path/filepath"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// GetFullPath from a relative path
func GetFullPath(relPath string) (string, error) {
	absPath, err := filepath.Abs(relPath)
	return absPath, err
}

// GetLogo prints the logo to the logs (for some fun!)
func GetLogo() string {
	return `
	██████╗ ██████╗  █████╗ ███████╗       ██████╗███████╗██╗
	██╔═══██╗██╔══██╗██╔══██╗██╔════╝      ██╔════╝██╔════╝██║
	██║   ██║██████╔╝███████║███████╗█████╗██║     ███████╗██║
	██║   ██║██╔══██╗██╔══██║╚════██║╚════╝██║     ╚════██║██║
	╚██████╔╝██║  ██║██║  ██║███████║      ╚██████╗███████║██║
	 ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝       ╚═════╝╚══════╝╚═╝`
}

// ListContains to determine if a list include a string
func ListContains(list []string, lookingFor string) bool {
	for _, b := range list {
		if b == lookingFor {
			return true
		}
	}
	return false
}

// DescToString strips a descriptor and return compacted json
func DescToString(desc ocispec.Descriptor) string {
	var stripped = struct {
		MediaType string
		Digest    string
		Size      int64
	}{
		desc.MediaType,
		desc.Digest.String(),
		desc.Size,
	}

	// marshal stripped and compact
	descJSON, _ := json.Marshal(stripped)
	return string(descJSON)
}
