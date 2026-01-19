package domain

import (
	"errors"
	"regexp"
	"strings"
)

// Validation errors returned by helper functions.
var (
	ErrInvalidAppName = errors.New("invalid app name")
	ErrInvalidImage   = errors.New("invalid image ref")
	ErrInvalidPort    = errors.New("invalid port")

	// lowercase letters + digits, seperated by single hyphens
	appNameRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// ValidateAppName checks the app name format and returns a validation error.
func ValidateAppName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || !appNameRe.MatchString(name) {
		return ErrInvalidAppName
	}
	return nil
}

// ValidateImageRef checks that the image reference is not empty.
func ValidateImageRef(image string) error {
	if strings.TrimSpace(image) == "" {
		return ErrInvalidImage
	}
	//v0: we keep permissive for now parsing wil come later
	return nil
}

// ValidatePort ensures the port is within the valid TCP range.
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return ErrInvalidPort
	}
	return nil
}
