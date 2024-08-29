package utils

import "strings"

func IsValidFileType(contentType string, allowTypes []string) bool {
	for _, allowType := range allowTypes {
		if strings.HasPrefix(contentType, allowType) {
			return true
		}
	}
	return false
}
