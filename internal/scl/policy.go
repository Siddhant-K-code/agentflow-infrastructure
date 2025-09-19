package scl

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type PolicyEngine struct {
	rules map[string]PolicyRule
}

type PolicyRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Condition   string                 `json:"condition"` // CEL expression
	Action      PolicyAction           `json:"action"`
	Severity    string                 `json:"severity"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type PolicyAction string

const (
	ActionAllow  PolicyAction = "allow"
	ActionDeny   PolicyAction = "deny"
	ActionWarn   PolicyAction = "warn"
	ActionRedact PolicyAction = "redact"
)

type PolicyResult struct {
	Allowed    bool                   `json:"allowed"`
	Reason     string                 `json:"reason,omitempty"`
	Violations []PolicyViolationInfo  `json:"violations,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type PolicyViolationInfo struct {
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Location   string `json:"location,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		rules: getDefaultPolicyRules(),
	}
}

// Evaluate evaluates content against policies
func (pe *PolicyEngine) Evaluate(ctx context.Context, orgID uuid.UUID, content interface{}, hints map[string]interface{}) (*PolicyResult, error) {
	result := &PolicyResult{
		Allowed:    true,
		Violations: make([]PolicyViolationInfo, 0),
		Metadata:   make(map[string]interface{}),
	}

	// Convert content to string for evaluation
	contentStr := fmt.Sprintf("%v", content)

	// Apply each rule
	for _, rule := range pe.rules {
		violation, err := pe.evaluateRule(rule, contentStr, hints)
		if err != nil {
			continue // Skip rules that fail to evaluate
		}

		if violation != nil {
			result.Violations = append(result.Violations, *violation)

			// Determine if this should block the request
			if rule.Action == ActionDeny {
				result.Allowed = false
				if result.Reason == "" {
					result.Reason = violation.Message
				}
			}
		}
	}

	return result, nil
}

// CheckAccess checks if access to a context bundle is allowed
func (pe *PolicyEngine) CheckAccess(ctx context.Context, orgID uuid.UUID, bundleID uuid.UUID, policy map[string]interface{}) (bool, error) {
	// Mock access control implementation
	// In production, this would integrate with OpenFGA or similar

	// Check if user has read access to the bundle
	if policy != nil {
		if requiredRole, exists := policy["required_role"]; exists {
			if role, ok := requiredRole.(string); ok {
				// Mock role check
				if role == "admin" || role == "reader" {
					return true, nil
				}
				return false, fmt.Errorf("insufficient permissions: requires %s role", role)
			}
		}
	}

	return true, nil // Default allow
}

func (pe *PolicyEngine) evaluateRule(rule PolicyRule, content string, hints map[string]interface{}) (*PolicyViolationInfo, error) {
	// Simple rule evaluation - in production would use CEL or similar
	contentLower := strings.ToLower(content)

	switch rule.ID {
	case "no_secrets":
		if pe.containsSecrets(contentLower) {
			return &PolicyViolationInfo{
				Rule:       rule.ID,
				Severity:   rule.Severity,
				Message:    "Content contains potential secrets",
				Suggestion: "Remove or redact sensitive information",
			}, nil
		}

	case "no_pii":
		if pe.containsPII(contentLower) {
			return &PolicyViolationInfo{
				Rule:       rule.ID,
				Severity:   rule.Severity,
				Message:    "Content contains potential PII",
				Suggestion: "Remove or redact personal information",
			}, nil
		}

	case "content_length":
		maxLength := 50000 // Default
		if hints != nil {
			if ml, exists := hints["max_length"]; exists {
				if length, ok := ml.(int); ok {
					maxLength = length
				}
			}
		}
		if len(content) > maxLength {
			return &PolicyViolationInfo{
				Rule:       rule.ID,
				Severity:   rule.Severity,
				Message:    fmt.Sprintf("Content exceeds maximum length of %d characters", maxLength),
				Suggestion: "Reduce content size or split into smaller chunks",
			}, nil
		}

	case "no_malicious_content":
		if pe.containsMaliciousContent(contentLower) {
			return &PolicyViolationInfo{
				Rule:       rule.ID,
				Severity:   rule.Severity,
				Message:    "Content contains potentially malicious patterns",
				Suggestion: "Review and sanitize content",
			}, nil
		}

	case "language_filter":
		if pe.containsInappropriateLanguage(contentLower) {
			return &PolicyViolationInfo{
				Rule:       rule.ID,
				Severity:   rule.Severity,
				Message:    "Content contains inappropriate language",
				Suggestion: "Remove offensive language",
			}, nil
		}

	case "external_links":
		if pe.containsExternalLinks(content) {
			return &PolicyViolationInfo{
				Rule:       rule.ID,
				Severity:   "warning",
				Message:    "Content contains external links",
				Suggestion: "Verify external links are safe and necessary",
			}, nil
		}
	}

	return nil, nil
}

func (pe *PolicyEngine) containsSecrets(content string) bool {
	secretPatterns := []string{
		"password", "passwd", "pwd",
		"secret", "key", "token",
		"api_key", "apikey",
		"private_key", "privatekey",
		"access_token", "accesstoken",
		"bearer",
	}

	for _, pattern := range secretPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (pe *PolicyEngine) containsPII(content string) bool {
	// Simple PII detection - in production would use more sophisticated methods
	piiPatterns := []string{
		"@", // Email indicator
		"ssn", "social security",
		"credit card", "creditcard",
		"phone", "telephone",
		"address", "street",
		"birthday", "birth date",
		"driver", "license",
		"passport",
	}

	for _, pattern := range piiPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (pe *PolicyEngine) containsMaliciousContent(content string) bool {
	maliciousPatterns := []string{
		"<script", "</script>",
		"javascript:",
		"eval(",
		"exec(",
		"system(",
		"shell_exec",
		"passthru",
		"file_get_contents",
		"curl_exec",
		"wget",
		"nc -", "netcat",
		"/etc/passwd",
		"/etc/shadow",
		"rm -rf",
		"format c:",
		"del /f",
	}

	for _, pattern := range maliciousPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

func (pe *PolicyEngine) containsInappropriateLanguage(content string) bool {
	// Basic profanity filter - in production would use comprehensive lists
	inappropriateWords := []string{
		// Add inappropriate words here
		"badword1", "badword2", // Placeholder
	}

	words := strings.Fields(content)
	for _, word := range words {
		for _, inappropriate := range inappropriateWords {
			if strings.EqualFold(word, inappropriate) {
				return true
			}
		}
	}

	return false
}

func (pe *PolicyEngine) containsExternalLinks(content string) bool {
	return strings.Contains(content, "http://") || strings.Contains(content, "https://")
}

func getDefaultPolicyRules() map[string]PolicyRule {
	return map[string]PolicyRule{
		"no_secrets": {
			ID:          "no_secrets",
			Name:        "No Secrets",
			Description: "Prevent secrets and credentials from being processed",
			Action:      ActionDeny,
			Severity:    "high",
		},
		"no_pii": {
			ID:          "no_pii",
			Name:        "No PII",
			Description: "Prevent personally identifiable information from being processed",
			Action:      ActionWarn,
			Severity:    "medium",
		},
		"content_length": {
			ID:          "content_length",
			Name:        "Content Length Limit",
			Description: "Enforce maximum content length",
			Action:      ActionDeny,
			Severity:    "low",
		},
		"no_malicious_content": {
			ID:          "no_malicious_content",
			Name:        "No Malicious Content",
			Description: "Prevent potentially malicious content",
			Action:      ActionDeny,
			Severity:    "high",
		},
		"language_filter": {
			ID:          "language_filter",
			Name:        "Language Filter",
			Description: "Filter inappropriate language",
			Action:      ActionWarn,
			Severity:    "medium",
		},
		"external_links": {
			ID:          "external_links",
			Name:        "External Links",
			Description: "Warn about external links in content",
			Action:      ActionWarn,
			Severity:    "low",
		},
	}
}
