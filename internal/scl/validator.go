package scl

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type Validator struct {
	schemaCache map[string]interface{}
}

func NewValidator() *Validator {
	return &Validator{
		schemaCache: make(map[string]interface{}),
	}
}

// ValidateSchema validates content against a JSON schema
func (v *Validator) ValidateSchema(content interface{}, schemaURI string) error {
	// Parse schema URI
	parsedURI, err := url.Parse(schemaURI)
	if err != nil {
		return fmt.Errorf("invalid schema URI: %w", err)
	}

	// Get schema (mock implementation)
	schema, err := v.getSchema(schemaURI)
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	// Validate content against schema
	return v.validateAgainstSchema(content, schema, parsedURI.Fragment)
}

func (v *Validator) getSchema(schemaURI string) (map[string]interface{}, error) {
	// Check cache first
	if cached, exists := v.schemaCache[schemaURI]; exists {
		if schema, ok := cached.(map[string]interface{}); ok {
			return schema, nil
		}
	}

	// Mock schema loading - in production would fetch from URI
	schema := v.getMockSchema(schemaURI)
	v.schemaCache[schemaURI] = schema
	
	return schema, nil
}

func (v *Validator) getMockSchema(schemaURI string) map[string]interface{} {
	// Return different mock schemas based on URI
	if strings.Contains(schemaURI, "document") {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "string",
				},
				"content": map[string]interface{}{
					"type": "string",
				},
				"metadata": map[string]interface{}{
					"type": "object",
				},
			},
			"required": []string{"title", "content"},
		}
	}

	// Default schema
	return map[string]interface{}{
		"type": "object",
		"additionalProperties": true,
	}
}

func (v *Validator) validateAgainstSchema(content interface{}, schema map[string]interface{}, fragment string) error {
	// Basic JSON schema validation implementation
	// In production, would use a proper JSON schema validator library

	schemaType, ok := schema["type"].(string)
	if !ok {
		return nil // No type constraint
	}

	switch schemaType {
	case "object":
		return v.validateObject(content, schema)
	case "array":
		return v.validateArray(content, schema)
	case "string":
		return v.validateString(content, schema)
	case "number":
		return v.validateNumber(content, schema)
	case "boolean":
		return v.validateBoolean(content, schema)
	default:
		return fmt.Errorf("unsupported schema type: %s", schemaType)
	}
}

func (v *Validator) validateObject(content interface{}, schema map[string]interface{}) error {
	contentMap, ok := content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object, got %T", content)
	}

	// Check required properties
	if required, exists := schema["required"]; exists {
		if requiredSlice, ok := required.([]interface{}); ok {
			for _, req := range requiredSlice {
				if reqStr, ok := req.(string); ok {
					if _, exists := contentMap[reqStr]; !exists {
						return fmt.Errorf("required property %s is missing", reqStr)
					}
				}
			}
		}
	}

	// Validate properties
	if properties, exists := schema["properties"]; exists {
		if propsMap, ok := properties.(map[string]interface{}); ok {
			for propName, propSchema := range propsMap {
				if propValue, exists := contentMap[propName]; exists {
					if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
						if err := v.validateAgainstSchema(propValue, propSchemaMap, ""); err != nil {
							return fmt.Errorf("property %s: %w", propName, err)
						}
					}
				}
			}
		}
	}

	// Check additional properties
	if additionalProps, exists := schema["additionalProperties"]; exists {
		if allowed, ok := additionalProps.(bool); ok && !allowed {
			if properties, exists := schema["properties"]; exists {
				if propsMap, ok := properties.(map[string]interface{}); ok {
					for propName := range contentMap {
						if _, allowed := propsMap[propName]; !allowed {
							return fmt.Errorf("additional property %s is not allowed", propName)
						}
					}
				}
			}
		}
	}

	return nil
}

func (v *Validator) validateArray(content interface{}, schema map[string]interface{}) error {
	contentSlice, ok := content.([]interface{})
	if !ok {
		return fmt.Errorf("expected array, got %T", content)
	}

	// Check min/max items
	if minItems, exists := schema["minItems"]; exists {
		if min, ok := minItems.(float64); ok {
			if len(contentSlice) < int(min) {
				return fmt.Errorf("array has %d items, minimum is %d", len(contentSlice), int(min))
			}
		}
	}

	if maxItems, exists := schema["maxItems"]; exists {
		if max, ok := maxItems.(float64); ok {
			if len(contentSlice) > int(max) {
				return fmt.Errorf("array has %d items, maximum is %d", len(contentSlice), int(max))
			}
		}
	}

	// Validate items
	if items, exists := schema["items"]; exists {
		if itemSchema, ok := items.(map[string]interface{}); ok {
			for i, item := range contentSlice {
				if err := v.validateAgainstSchema(item, itemSchema, ""); err != nil {
					return fmt.Errorf("item %d: %w", i, err)
				}
			}
		}
	}

	return nil
}

func (v *Validator) validateString(content interface{}, schema map[string]interface{}) error {
	contentStr, ok := content.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", content)
	}

	// Check min/max length
	if minLength, exists := schema["minLength"]; exists {
		if min, ok := minLength.(float64); ok {
			if len(contentStr) < int(min) {
				return fmt.Errorf("string length %d is less than minimum %d", len(contentStr), int(min))
			}
		}
	}

	if maxLength, exists := schema["maxLength"]; exists {
		if max, ok := maxLength.(float64); ok {
			if len(contentStr) > int(max) {
				return fmt.Errorf("string length %d exceeds maximum %d", len(contentStr), int(max))
			}
		}
	}

	// Check pattern
	if pattern, exists := schema["pattern"]; exists {
		if patternStr, ok := pattern.(string); ok {
			// Would use regexp.MatchString in production
			if !strings.Contains(contentStr, patternStr) {
				return fmt.Errorf("string does not match pattern %s", patternStr)
			}
		}
	}

	// Check enum
	if enum, exists := schema["enum"]; exists {
		if enumSlice, ok := enum.([]interface{}); ok {
			found := false
			for _, enumValue := range enumSlice {
				if enumStr, ok := enumValue.(string); ok && enumStr == contentStr {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("string value %s is not in allowed enum values", contentStr)
			}
		}
	}

	return nil
}

func (v *Validator) validateNumber(content interface{}, schema map[string]interface{}) error {
	var contentNum float64
	var ok bool

	switch v := content.(type) {
	case float64:
		contentNum = v
		ok = true
	case int:
		contentNum = float64(v)
		ok = true
	case json.Number:
		if f, err := v.Float64(); err == nil {
			contentNum = f
			ok = true
		}
	}

	if !ok {
		return fmt.Errorf("expected number, got %T", content)
	}

	// Check minimum
	if minimum, exists := schema["minimum"]; exists {
		if min, ok := minimum.(float64); ok {
			if contentNum < min {
				return fmt.Errorf("number %f is less than minimum %f", contentNum, min)
			}
		}
	}

	// Check maximum
	if maximum, exists := schema["maximum"]; exists {
		if max, ok := maximum.(float64); ok {
			if contentNum > max {
				return fmt.Errorf("number %f exceeds maximum %f", contentNum, max)
			}
		}
	}

	return nil
}

func (v *Validator) validateBoolean(content interface{}, schema map[string]interface{}) error {
	_, ok := content.(bool)
	if !ok {
		return fmt.Errorf("expected boolean, got %T", content)
	}

	return nil
}