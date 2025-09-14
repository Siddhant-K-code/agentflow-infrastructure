package pop

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type TemplateRenderer struct {
	funcMap template.FuncMap
}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{
		funcMap: template.FuncMap{
			"upper":    strings.ToUpper,
			"lower":    strings.ToLower,
			"title":    strings.Title,
			"trim":     strings.TrimSpace,
			"join":     strings.Join,
			"contains": strings.Contains,
			"default":  defaultValue,
		},
	}
}

// Validate checks if a template is syntactically correct
func (r *TemplateRenderer) Validate(templateText string, schema Schema) error {
	// Parse template
	tmpl, err := template.New("prompt").Funcs(r.funcMap).Parse(templateText)
	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}

	// Create dummy data based on schema
	dummyData := r.createDummyData(schema)

	// Try to execute template with dummy data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, dummyData); err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}

	// Additional validations
	if err := r.validateSafety(templateText); err != nil {
		return fmt.Errorf("safety validation failed: %w", err)
	}

	return nil
}

// Render executes a template with the given data
func (r *TemplateRenderer) Render(templateText string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("prompt").Funcs(r.funcMap).Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// createDummyData creates dummy data for template validation
func (r *TemplateRenderer) createDummyData(schema Schema) map[string]interface{} {
	data := make(map[string]interface{})

	for name, prop := range schema.Properties {
		switch prop.Type {
		case "string":
			if prop.Default != nil {
				data[name] = prop.Default
			} else if len(prop.Enum) > 0 {
				data[name] = prop.Enum[0]
			} else {
				data[name] = "test_string"
			}
		case "number", "integer":
			if prop.Default != nil {
				data[name] = prop.Default
			} else {
				data[name] = 42
			}
		case "boolean":
			if prop.Default != nil {
				data[name] = prop.Default
			} else {
				data[name] = true
			}
		case "array":
			data[name] = []interface{}{"item1", "item2"}
		case "object":
			data[name] = map[string]interface{}{"key": "value"}
		default:
			data[name] = "unknown_type"
		}
	}

	return data
}

// validateSafety performs security checks on templates
func (r *TemplateRenderer) validateSafety(templateText string) error {
	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		"{{.}}",           // Direct object access
		"call",            // Function calls
		"index",           // Index access
		"js",              // JavaScript
		"script",          // Script tags
		"eval",            // Eval functions
	}

	lowerTemplate := strings.ToLower(templateText)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerTemplate, pattern) {
			return fmt.Errorf("potentially unsafe pattern detected: %s", pattern)
		}
	}

	return nil
}

// defaultValue provides a default value if the input is nil or empty
func defaultValue(defaultVal interface{}, value interface{}) interface{} {
	if value == nil {
		return defaultVal
	}
	
	if str, ok := value.(string); ok && str == "" {
		return defaultVal
	}
	
	return value
}