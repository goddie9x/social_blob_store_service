package utils

import "strings"

func GetValidFileType(contentType string, allowTypes []string) string {
	for _, allowType := range allowTypes {
		if strings.HasPrefix(contentType, allowType) {
			return allowType
		}
	}
	return ""
}
