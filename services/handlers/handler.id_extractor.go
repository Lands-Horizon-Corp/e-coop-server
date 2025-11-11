package handlers

import (
	"errors"
	"fmt"
	"strings"
)

// IDExtractor extracts dynamic parameters from a URL path based on a pattern.
type IDExtractor struct {
	patternSegments []string
	pathSegments    []string
}

// NewIDExtractor creates a new IDExtractor with proper slash splitting.
func NewIDExtractor(pattern, path string) (*IDExtractor, error) {
	patternSegments := splitURLPath(pattern)
	pathSegments := splitURLPath(path)

	if len(pathSegments) < len(patternSegments) {
		return nil, errors.New("path does not match pattern: too few segments")
	}

	return &IDExtractor{
		patternSegments: patternSegments,
		pathSegments:    pathSegments,
	}, nil
}

// splitURLPath safely splits a path by '/' while preserving structure.
func splitURLPath(urlPath string) []string {
	urlPath = strings.Trim(urlPath, "/")
	if urlPath == "" {
		return []string{}
	}
	return strings.Split(urlPath, "/")
}

// ExtractParameterByName gets a single parameter by name (without the colon).
func (extractor *IDExtractor) ExtractParameterByName(parameterName string) (string, error) {
	parameterPattern := ":" + parameterName
	for index, segment := range extractor.patternSegments {
		if segment == parameterPattern {
			if index >= len(extractor.pathSegments) {
				return "", errors.New("path too short for parameter")
			}
			return extractor.pathSegments[index], nil
		}
	}
	return "", fmt.Errorf("parameter '%s' not found in pattern", parameterName)
}

// ExtractAllParameters extracts all named parameters (:param) into a map.
func (extractor *IDExtractor) ExtractAllParameters() map[string]string {
	parameters := make(map[string]string)
	for index, segment := range extractor.patternSegments {
		if strings.HasPrefix(segment, ":") && index < len(extractor.pathSegments) {
			parameterName := strings.TrimPrefix(segment, ":")
			parameters[parameterName] = extractor.pathSegments[index]
		}
	}
	return parameters
}

// IsPatternMatching checks if the path matches the pattern structure (useful for routing)
func (extractor *IDExtractor) IsPatternMatching() bool {
	if len(extractor.pathSegments) < len(extractor.patternSegments) {
		return false
	}
	for index, segment := range extractor.patternSegments {
		if index >= len(extractor.pathSegments) {
			return false
		}
		if !strings.HasPrefix(segment, ":") && segment != extractor.pathSegments[index] {
			return false
		}
	}
	return true
}

/*

// Create extractor
extractor, err := NewIDExtractor("/api/users/:userID/posts/:postID", "/api/users/123/posts/456")

// Extract single parameter
userID, err := extractor.ExtractParameterByName("userID") // Returns "123"

// Extract all parameters
allParams := extractor.ExtractAllParameters() // Returns map[string]string{"userID": "123", "postID": "456"}

// Check if pattern matches
isValid := extractor.IsPatternMatching() // Returns true

*/
