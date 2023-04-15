package utils

import (
	"path/filepath"
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
