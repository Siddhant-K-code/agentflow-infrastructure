package scl

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

type Sanitizer struct {
	injectionPatterns []*regexp.Regexp
	htmlTags          *regexp.Regexp
	scriptTags        *regexp.Regexp
	sqlPatterns       []*regexp.Regexp
}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		injectionPatterns: compileInjectionPatterns(),
		htmlTags:          regexp.MustCompile(`<[^>]*>`),
		scriptTags:        regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		sqlPatterns:       compileSQLPatterns(),
	}
}

// Sanitize cleans and validates content for security issues
func (s *Sanitizer) Sanitize(content interface{}) (interface{}, []string, error) {
	warnings := make([]string, 0)

	switch v := content.(type) {
	case string:
		sanitized, warns := s.sanitizeString(v)
		warnings = append(warnings, warns...)
		return sanitized, warnings, nil
	case map[string]interface{}:
		sanitized, warns, err := s.sanitizeMap(v)
		warnings = append(warnings, warns...)
		return sanitized, warnings, err
	case []interface{}:
		sanitized, warns, err := s.sanitizeSlice(v)
		warnings = append(warnings, warns...)
		return sanitized, warnings, err
	default:
		return content, warnings, nil
	}
}

func (s *Sanitizer) sanitizeString(input string) (string, []string) {
	warnings := make([]string, 0)
	result := input

	// Check for prompt injection patterns
	for _, pattern := range s.injectionPatterns {
		if pattern.MatchString(result) {
			warnings = append(warnings, fmt.Sprintf("Potential prompt injection detected: %s", pattern.String()))
		}
	}

	// Check for SQL injection patterns
	for _, pattern := range s.sqlPatterns {
		if pattern.MatchString(strings.ToLower(result)) {
			warnings = append(warnings, fmt.Sprintf("Potential SQL injection detected: %s", pattern.String()))
		}
	}

	// Remove script tags
	if s.scriptTags.MatchString(result) {
		warnings = append(warnings, "Script tags removed")
		result = s.scriptTags.ReplaceAllString(result, "")
	}

	// Escape HTML if present
	if s.htmlTags.MatchString(result) {
		warnings = append(warnings, "HTML content escaped")
		result = html.EscapeString(result)
	}

	// Normalize whitespace
	result = strings.TrimSpace(result)
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	// Check for excessive length
	if len(result) > 100000 { // 100KB limit
		warnings = append(warnings, "Content truncated due to excessive length")
		result = result[:100000] + "... [truncated]"
	}

	return result, warnings
}

func (s *Sanitizer) sanitizeMap(input map[string]interface{}) (map[string]interface{}, []string, error) {
	warnings := make([]string, 0)
	result := make(map[string]interface{})

	for key, value := range input {
		// Sanitize key
		sanitizedKey, keyWarnings := s.sanitizeString(key)
		warnings = append(warnings, keyWarnings...)

		// Sanitize value
		sanitizedValue, valueWarnings, err := s.Sanitize(value)
		if err != nil {
			return nil, warnings, fmt.Errorf("failed to sanitize value for key %s: %w", key, err)
		}
		warnings = append(warnings, valueWarnings...)

		result[sanitizedKey] = sanitizedValue
	}

	return result, warnings, nil
}

func (s *Sanitizer) sanitizeSlice(input []interface{}) ([]interface{}, []string, error) {
	warnings := make([]string, 0)
	result := make([]interface{}, len(input))

	for i, item := range input {
		sanitizedItem, itemWarnings, err := s.Sanitize(item)
		if err != nil {
			return nil, warnings, fmt.Errorf("failed to sanitize item at index %d: %w", i, err)
		}
		warnings = append(warnings, itemWarnings...)
		result[i] = sanitizedItem
	}

	return result, warnings, nil
}

func compileInjectionPatterns() []*regexp.Regexp {
	patterns := []string{
		// Common prompt injection patterns
		`(?i)ignore\s+(previous|above|all)\s+(instructions?|prompts?)`,
		`(?i)forget\s+(everything|all|previous)`,
		`(?i)you\s+are\s+now\s+a`,
		`(?i)act\s+as\s+(if\s+you\s+are\s+)?a`,
		`(?i)pretend\s+(to\s+be\s+)?you\s+are`,
		`(?i)roleplay\s+as`,
		`(?i)simulate\s+(being\s+)?a`,
		`(?i)new\s+instructions?:`,
		`(?i)system\s*:\s*`,
		`(?i)assistant\s*:\s*`,
		`(?i)human\s*:\s*`,
		`(?i)user\s*:\s*`,
		`(?i)###\s*(instruction|prompt|system)`,
		`(?i)---\s*(instruction|prompt|system)`,
		`(?i)\[INST\]`,
		`(?i)\[/INST\]`,
		`(?i)<\|.*?\|>`,
		`(?i)jailbreak`,
		`(?i)break\s+out\s+of`,
		`(?i)escape\s+your`,
		`(?i)override\s+your`,
		`(?i)bypass\s+your`,
		`(?i)disable\s+your`,
		`(?i)turn\s+off\s+your`,
		`(?i)stop\s+being\s+helpful`,
		`(?i)be\s+harmful`,
		`(?i)be\s+evil`,
		`(?i)be\s+malicious`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, re)
		}
	}

	return compiled
}

func compileSQLPatterns() []*regexp.Regexp {
	patterns := []string{
		// SQL injection patterns
		`(?i)union\s+select`,
		`(?i)drop\s+table`,
		`(?i)delete\s+from`,
		`(?i)insert\s+into`,
		`(?i)update\s+.*\s+set`,
		`(?i)exec\s*\(`,
		`(?i)execute\s*\(`,
		`(?i)sp_executesql`,
		`(?i)xp_cmdshell`,
		`(?i)--\s*$`,
		`(?i)/\*.*\*/`,
		`(?i);\s*drop`,
		`(?i);\s*delete`,
		`(?i);\s*insert`,
		`(?i);\s*update`,
		`(?i)'\s*or\s*'1'\s*=\s*'1`,
		`(?i)'\s*or\s*1\s*=\s*1`,
		`(?i)"\s*or\s*"1"\s*=\s*"1`,
		`(?i)"\s*or\s*1\s*=\s*1`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, re)
		}
	}

	return compiled
}