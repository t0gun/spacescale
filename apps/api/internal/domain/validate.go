// Validation helpers for app input and image refs
// App names follow allowed patterns for safety
// Image refs must be present to deploy
// Port values are validated when provided
// Errors are returned for invalid inputs

package domain

import (
	"errors"
	"regexp"
	"strings"
)

// Validation errors returned by helper functions
var (
	ErrInvalidAppName = errors.New("invalid app name")
	ErrInvalidImage   = errors.New("invalid image ref")
	ErrInvalidPort    = errors.New("invalid port")

	// lowercase letters digits seperated by single hyphens
	appNameRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// ValidateAppName This function handles validate app name
// It supports validate app name behavior
func ValidateAppName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || !appNameRe.MatchString(name) {
		return ErrInvalidAppName
	}
	return nil
}

// ValidateImageRef This function handles validate image ref
// It supports validate image ref behavior
func ValidateImageRef(image string) error {
	if strings.TrimSpace(image) == "" {
		return ErrInvalidImage
	}
	// v we keep permissive for now parsing wil come later
	return nil
}

// ValidatePort This function handles validate port
// It supports validate port behavior
func ValidatePort(port *int) error {
	if port == nil {
		return nil
	}
	if *port < 1 || *port > 65535 {
		return ErrInvalidPort
	}
	return nil
}
