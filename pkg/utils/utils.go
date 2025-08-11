package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GenerateID generates a unique ID
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// HashString creates an MD5 hash of a string
func HashString(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// CleanString removes extra whitespace and normalizes line endings
func CleanString(s string) string {
	// Replace Windows line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// Replace Mac line endings
	s = strings.ReplaceAll(s, "\r", "\n")
	// Remove trailing whitespace from each line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// ExtractCodeBlocks extracts code blocks from markdown text
func ExtractCodeBlocks(markdown string) []string {
	var blocks []string
	lines := strings.Split(markdown, "\n")
	var currentBlock strings.Builder
	inBlock := false

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inBlock {
				// End of block
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				inBlock = false
			} else {
				// Start of block
				inBlock = true
			}
		} else if inBlock {
			currentBlock.WriteString(line + "\n")
		}
	}

	return blocks
}

// SanitizeFilename removes invalid characters from a filename
func SanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	return filename
}

// FormatBytes formats bytes in a human-readable way
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// ExtractDomain extracts domain from URL
func ExtractDomain(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}

	if idx := strings.Index(url, "/"); idx > 0 {
		url = url[:idx]
	}

	return url
}

// SliceContains checks if a slice contains a string
func SliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ChunkString splits a string into chunks of specified size
func ChunkString(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{s}
	}

	var chunks []string
	runes := []rune(s)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

// WordCount counts words in a string
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// EstimateReadingTime estimates reading time in minutes
func EstimateReadingTime(text string) int {
	words := WordCount(text)
	// Average reading speed: 200 words per minute
	minutes := words / 200
	if minutes < 1 {
		return 1
	}
	return minutes
}

// PascalCase converts a string to PascalCase
func PascalCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// CamelCase converts a string to camelCase
func CamelCase(s string) string {
	pascal := PascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(pascal[:1]) + pascal[1:]
	}
	return pascal
}

// KebabCase converts a string to kebab-case
func KebabCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "-")
}

// SnakeCase converts a string to snake_case
func SnakeCase(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "_")
}

// Pluralize adds 's' to a word if count is not 1
func Pluralize(word string, count int) string {
	if count == 1 {
		return word
	}

	// Simple pluralization rules
	if strings.HasSuffix(word, "y") {
		return word[:len(word)-1] + "ies"
	}
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "sh") || strings.HasSuffix(word, "ch") {
		return word + "es"
	}
	return word + "s"
}

// FormatNumber formats a number with thousand separators
func FormatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}

	return result.String()
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps a value between min and max
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// DefaultString returns the default value if the string is empty
func DefaultString(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}

// DefaultInt returns the default value if the int is zero
func DefaultInt(i, defaultValue int) int {
	if i == 0 {
		return defaultValue
	}
	return i
}

// URLToPackagePath converts a repository URL to a package path
// Example: https://github.com/gorilla/websocket -> github.com/gorilla/websocket
func URLToPackagePath(repoURL string) string {
	// Remove protocol
	repoURL = strings.TrimPrefix(repoURL, "https://")
	repoURL = strings.TrimPrefix(repoURL, "http://")

	// Remove .git suffix
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Remove trailing slash
	repoURL = strings.TrimSuffix(repoURL, "/")

	return repoURL
}

// PackagePathToURL converts a package path to a repository URL
// Example: github.com/gorilla/websocket -> https://github.com/gorilla/websocket
func PackagePathToURL(packagePath string) string {
	if !strings.HasPrefix(packagePath, "http") {
		return "https://" + packagePath
	}
	return packagePath
}

// SanitizePackagePath sanitizes a package path for use in URLs
func SanitizePackagePath(packagePath string) string {
	// Replace invalid URL characters
	packagePath = strings.ReplaceAll(packagePath, " ", "-")
	packagePath = strings.ReplaceAll(packagePath, "_", "-")

	// Use regex to keep only valid characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-\./]`)
	packagePath = reg.ReplaceAllString(packagePath, "")

	return strings.ToLower(packagePath)
}

// ExtractRepositoryInfo extracts owner and repo name from a repository URL
func ExtractRepositoryInfo(repoURL string) (owner, repo string) {
	packagePath := URLToPackagePath(repoURL)

	// Split by / and get the last two parts
	parts := strings.Split(packagePath, "/")
	if len(parts) >= 3 {
		// For github.com/owner/repo format
		owner = parts[len(parts)-2]
		repo = parts[len(parts)-1]
	} else if len(parts) == 2 {
		// For owner/repo format
		owner = parts[0]
		repo = parts[1]
	}

	return owner, repo
}

// IsValidPackagePath checks if a string is a valid package path
func IsValidPackagePath(packagePath string) bool {
	// Basic validation for package path format
	if packagePath == "" {
		return false
	}

	// Should contain at least one slash
	if !strings.Contains(packagePath, "/") {
		return false
	}

	// Should not start or end with slash
	if strings.HasPrefix(packagePath, "/") || strings.HasSuffix(packagePath, "/") {
		return false
	}

	return true
}
