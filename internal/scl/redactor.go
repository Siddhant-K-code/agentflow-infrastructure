package scl

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

type Redactor struct {
	piiPatterns map[string]*regexp.Regexp
	tokenMap    map[string]string // For reversible redaction
}

func NewRedactor() *Redactor {
	return &Redactor{
		piiPatterns: compilePIIPatterns(),
		tokenMap:    make(map[string]string),
	}
}

// Redact removes or replaces PII and sensitive information
func (r *Redactor) Redact(content interface{}) (interface{}, map[string]string, error) {
	redactionMap := make(map[string]string)

	switch v := content.(type) {
	case string:
		redacted, mapping := r.redactString(v)
		for k, v := range mapping {
			redactionMap[k] = v
		}
		return redacted, redactionMap, nil
	case map[string]interface{}:
		redacted, mapping, err := r.redactMap(v)
		for k, v := range mapping {
			redactionMap[k] = v
		}
		return redacted, redactionMap, err
	case []interface{}:
		redacted, mapping, err := r.redactSlice(v)
		for k, v := range mapping {
			redactionMap[k] = v
		}
		return redacted, redactionMap, err
	default:
		return content, redactionMap, nil
	}
}

func (r *Redactor) redactString(input string) (string, map[string]string) {
	redactionMap := make(map[string]string)
	result := input

	// Apply PII patterns
	for piiType, pattern := range r.piiPatterns {
		matches := pattern.FindAllString(result, -1)
		for _, match := range matches {
			if match == "" {
				continue
			}

			// Generate deterministic token
			token := r.generateToken(match, piiType)
			redactionMap[token] = match

			// Replace in result
			result = strings.ReplaceAll(result, match, token)
		}
	}

	return result, redactionMap
}

func (r *Redactor) redactMap(input map[string]interface{}) (map[string]interface{}, map[string]string, error) {
	redactionMap := make(map[string]string)
	result := make(map[string]interface{})

	for key, value := range input {
		// Check if key itself needs redaction
		redactedKey, keyMapping := r.redactString(key)
		for k, v := range keyMapping {
			redactionMap[k] = v
		}

		// Redact value
		redactedValue, valueMapping, err := r.Redact(value)
		if err != nil {
			return nil, redactionMap, fmt.Errorf("failed to redact value for key %s: %w", key, err)
		}
		for k, v := range valueMapping {
			redactionMap[k] = v
		}

		result[redactedKey] = redactedValue
	}

	return result, redactionMap, nil
}

func (r *Redactor) redactSlice(input []interface{}) ([]interface{}, map[string]string, error) {
	redactionMap := make(map[string]string)
	result := make([]interface{}, len(input))

	for i, item := range input {
		redactedItem, itemMapping, err := r.Redact(item)
		if err != nil {
			return nil, redactionMap, fmt.Errorf("failed to redact item at index %d: %w", i, err)
		}
		for k, v := range itemMapping {
			redactionMap[k] = v
		}
		result[i] = redactedItem
	}

	return result, redactionMap, nil
}

func (r *Redactor) generateToken(original, piiType string) string {
	// Generate deterministic token based on content and type
	// In production, would use proper encryption/hashing

	// Check if we already have a token for this value
	key := fmt.Sprintf("%s:%s", piiType, original)
	if token, exists := r.tokenMap[key]; exists {
		return token
	}

	// Generate new token
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	token := fmt.Sprintf("[REDACTED_%s_%s]", strings.ToUpper(piiType), hex.EncodeToString(randomBytes)[:8])

	r.tokenMap[key] = token
	return token
}

func compilePIIPatterns() map[string]*regexp.Regexp {
	patterns := map[string]string{
		// Email addresses
		"email": `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,

		// Phone numbers (US format)
		"phone": `\b(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})\b`,

		// Social Security Numbers
		"ssn": `\b(?!000|666|9\d{2})\d{3}[-\s]?(?!00)\d{2}[-\s]?(?!0000)\d{4}\b`,

		// Credit Card Numbers (basic pattern)
		"credit_card": `\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3[0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`,

		// IP Addresses
		"ip_address": `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`,

		// URLs
		"url": `https?://(?:[-\w.])+(?:[:\d]+)?(?:/(?:[\w/_.])*(?:\?(?:[\w&=%.])*)?(?:#(?:[\w.])*)?)?`,

		// API Keys (common patterns)
		"api_key": `(?i)\b(?:api[_-]?key|access[_-]?token|secret[_-]?key|private[_-]?key)["\s]*[:=]["\s]*([a-zA-Z0-9_-]{20,})\b`,

		// AWS Access Keys
		"aws_access_key": `\bAKIA[0-9A-Z]{16}\b`,

		// AWS Secret Keys
		"aws_secret_key": `\b[A-Za-z0-9/+=]{40}\b`,

		// JWT Tokens
		"jwt": `\beyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*\b`,

		// Generic secrets (password-like patterns)
		"password": `(?i)\b(?:password|passwd|pwd)["\s]*[:=]["\s]*([^\s"']{8,})\b`,

		// Database connection strings
		"db_connection": `(?i)\b(?:mongodb|mysql|postgresql|postgres)://[^\s"']+\b`,

		// Private keys
		"private_key": `-----BEGIN (?:RSA |EC |DSA )?PRIVATE KEY-----[\s\S]*?-----END (?:RSA |EC |DSA )?PRIVATE KEY-----`,

		// MAC Addresses
		"mac_address": `\b(?:[0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b`,

		// IPv6 Addresses
		"ipv6": `\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`,

		// Bank Account Numbers (basic pattern)
		"bank_account": `\b\d{8,17}\b`,

		// Driver's License (basic US pattern)
		"drivers_license": `\b[A-Z]{1,2}\d{6,8}\b`,

		// Passport Numbers (basic pattern)
		"passport": `\b[A-Z]{1,2}\d{6,9}\b`,
	}

	compiled := make(map[string]*regexp.Regexp)
	for name, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiled[name] = re
		}
	}

	return compiled
}

// UnredactString reverses redaction using the redaction map
func (r *Redactor) UnredactString(redacted string, redactionMap map[string]string) string {
	result := redacted
	for token, original := range redactionMap {
		result = strings.ReplaceAll(result, token, original)
	}
	return result
}

// GetRedactionStats returns statistics about redacted content
func (r *Redactor) GetRedactionStats(redactionMap map[string]string) map[string]int {
	stats := make(map[string]int)

	for token := range redactionMap {
		// Extract PII type from token
		if strings.HasPrefix(token, "[REDACTED_") {
			parts := strings.Split(token, "_")
			if len(parts) >= 2 {
				piiType := strings.ToLower(parts[1])
				stats[piiType]++
			}
		}
	}

	return stats
}
