// ABOUTME: Encoding utilities for UTF-8 validation and content sanitization
// ABOUTME: Provides robust handling of non-UTF-8 content and encoding detection

package utils

import (
	"strings"
	"unicode/utf8"
)

// IsValidUTF8 checks if a string contains valid UTF-8 encoding
func IsValidUTF8(s string) bool {
	return utf8.ValidString(s)
}

// SanitizeUTF8 removes or replaces invalid UTF-8 sequences
func SanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	
	// Convert to valid UTF-8, replacing invalid sequences with replacement character
	return strings.ToValidUTF8(s, "�")
}

// CleanContent removes problematic characters that might cause encoding issues
func CleanContent(content string) string {
	// Remove null bytes and other control characters that can cause issues
	content = strings.ReplaceAll(content, "\x00", "")
	
	// Remove BOM (Byte Order Mark) if present
	content = strings.TrimPrefix(content, "\xEF\xBB\xBF") // UTF-8 BOM
	content = strings.TrimPrefix(content, "\xFF\xFE")     // UTF-16 LE BOM
	content = strings.TrimPrefix(content, "\xFE\xFF")     // UTF-16 BE BOM
	
	// Ensure valid UTF-8
	return SanitizeUTF8(content)
}

// IsProbablyBinary detects if content is likely binary data
func IsProbablyBinary(content string) bool {
	if len(content) == 0 {
		return false
	}
	
	// Check for null bytes (common in binary files)
	if strings.Contains(content, "\x00") {
		return true
	}
	
	// Check for high ratio of non-printable characters
	nonPrintable := 0
	for _, r := range content[:min(len(content), 8192)] { // Check first 8KB
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			nonPrintable++
		}
	}
	
	// If more than 30% are non-printable, consider it binary
	if float64(nonPrintable)/float64(len(content)) > 0.3 {
		return true
	}
	
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidateContentForStorage checks if content is safe for database storage
func ValidateContentForStorage(content string) (bool, string) {
	// Check if content is probably binary
	if IsProbablyBinary(content) {
		return false, "content appears to be binary data"
	}
	
	// Check for valid UTF-8
	if !utf8.ValidString(content) {
		return false, "content contains invalid UTF-8 sequences"
	}
	
	// Check for extremely long lines that might cause issues
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if len(line) > 10000 { // 10KB per line limit
			return false, "content contains extremely long lines (line " + string(rune(i+1)) + ")"
		}
	}
	
	return true, ""
}