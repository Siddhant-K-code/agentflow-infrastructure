package pop

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Siddhant-K-code/agentflow-infrastructure/internal/db"
)

type Evaluator struct {
	db *db.PostgresDB
}

func NewEvaluator(database *db.PostgresDB) *Evaluator {
	return &Evaluator{db: database}
}

// EvaluateCase evaluates a single test case
func (e *Evaluator) EvaluateCase(ctx context.Context, testCase TestCase, actualOutput interface{}) (*EvaluationResult, error) {
	start := time.Now()
	
	result := &EvaluationResult{
		CaseID:   testCase.ID,
		Input:    testCase.Input,
		Output:   actualOutput,
		Expected: testCase.Expected,
		Latency:  time.Since(start),
	}

	// Score based on scoring configuration
	score, passed, err := e.scoreOutput(actualOutput, testCase.Expected, testCase.Scoring)
	if err != nil {
		result.Error = err.Error()
		result.Score = 0
		result.Passed = false
	} else {
		result.Score = score
		result.Passed = passed
	}

	return result, nil
}

// scoreOutput scores the actual output against expected results
func (e *Evaluator) scoreOutput(actual interface{}, expected Expected, scoring ScoringConfig) (float64, bool, error) {
	switch scoring.Type {
	case ScoringExact:
		return e.scoreExact(actual, expected.Output)
	case ScoringContains:
		return e.scoreContains(actual, expected.Contains)
	case ScoringRegex:
		return e.scoreRegex(actual, scoring.Config)
	case ScoringSchema:
		return e.scoreSchema(actual, expected.Schema)
	case ScoringLLMJudge:
		return e.scoreLLMJudge(actual, expected, scoring.Config)
	case ScoringEmbedding:
		return e.scoreEmbedding(actual, expected.Output, scoring.Config)
	default:
		return 0, false, fmt.Errorf("unknown scoring type: %s", scoring.Type)
	}
}

func (e *Evaluator) scoreExact(actual, expected interface{}) (float64, bool, error) {
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)
	
	if actualStr == expectedStr {
		return 1.0, true, nil
	}
	return 0.0, false, nil
}

func (e *Evaluator) scoreContains(actual interface{}, contains []string) (float64, bool, error) {
	actualStr := strings.ToLower(fmt.Sprintf("%v", actual))
	
	matchCount := 0
	for _, term := range contains {
		if strings.Contains(actualStr, strings.ToLower(term)) {
			matchCount++
		}
	}
	
	if len(contains) == 0 {
		return 1.0, true, nil
	}
	
	score := float64(matchCount) / float64(len(contains))
	passed := matchCount == len(contains)
	
	return score, passed, nil
}

func (e *Evaluator) scoreRegex(actual interface{}, config map[string]interface{}) (float64, bool, error) {
	pattern, ok := config["pattern"].(string)
	if !ok {
		return 0, false, fmt.Errorf("regex pattern not specified")
	}
	
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return 0, false, fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	actualStr := fmt.Sprintf("%v", actual)
	if regex.MatchString(actualStr) {
		return 1.0, true, nil
	}
	
	return 0.0, false, nil
}

func (e *Evaluator) scoreSchema(actual interface{}, expectedSchema map[string]interface{}) (float64, bool, error) {
	// Basic JSON schema validation
	// In production, would use a proper JSON schema validator
	
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return 0, false, fmt.Errorf("failed to marshal actual output: %w", err)
	}
	
	var actualObj map[string]interface{}
	if err := json.Unmarshal(actualJSON, &actualObj); err != nil {
		return 0, false, fmt.Errorf("actual output is not a valid JSON object")
	}
	
	// Check required fields
	requiredFields, ok := expectedSchema["required"].([]interface{})
	if ok {
		for _, field := range requiredFields {
			fieldName, ok := field.(string)
			if !ok {
				continue
			}
			if _, exists := actualObj[fieldName]; !exists {
				return 0, false, fmt.Errorf("required field %s is missing", fieldName)
			}
		}
	}
	
	return 1.0, true, nil
}

func (e *Evaluator) scoreLLMJudge(actual interface{}, expected Expected, config map[string]interface{}) (float64, bool, error) {
	// Mock LLM judge implementation
	// In production, this would call an LLM to evaluate the output
	
	_, ok := config["rubric"].(string)
	if !ok {
		_ = "Evaluate if the output is appropriate and helpful"
	}
	
	// Mock scoring logic
	actualStr := fmt.Sprintf("%v", actual)
	if len(actualStr) > 10 && !strings.Contains(strings.ToLower(actualStr), "error") {
		return 0.8, true, nil // Mock good score
	}
	
	return 0.3, false, nil // Mock poor score
}

func (e *Evaluator) scoreEmbedding(actual, expected interface{}, config map[string]interface{}) (float64, bool, error) {
	// Mock embedding similarity implementation
	// In production, this would compute embedding similarity
	
	threshold, ok := config["threshold"].(float64)
	if !ok {
		threshold = 0.8
	}
	
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)
	
	// Mock similarity calculation (simple string similarity)
	similarity := e.calculateStringSimilarity(actualStr, expectedStr)
	
	passed := similarity >= threshold
	return similarity, passed, nil
}

// calculateStringSimilarity provides a mock string similarity calculation
func (e *Evaluator) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	// Simple Jaccard similarity on words
	words1 := strings.Fields(strings.ToLower(s1))
	words2 := strings.Fields(strings.ToLower(s2))
	
	if len(words1) == 0 && len(words2) == 0 {
		return 1.0
	}
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}
	
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, word := range words1 {
		set1[word] = true
	}
	
	for _, word := range words2 {
		set2[word] = true
	}
	
	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}
	
	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 1.0
	}
	
	return float64(intersection) / float64(union)
}