package pop

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service implements the PromptOps Platform
type Service struct {
	db *gorm.DB
}

// NewService creates a new POP service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreatePromptVersion creates a new prompt template version
func (s *Service) CreatePromptVersion(c *gin.Context) {
	name := c.Param("name")
	
	var req CreatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	// Get next version number
	var maxVersion int
	s.db.Model(&PromptTemplate{}).
		Where("org_id = ? AND name = ?", orgID, name).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion)

	// TODO: Validate template syntax and schema

	template := PromptTemplate{
		OrgID:    orgID,
		Name:     name,
		Version:  maxVersion + 1,
		Template: req.Template,
		Schema:   req.Schema,
	}

	if err := s.db.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetPrompt retrieves the latest version of a prompt
func (s *Service) GetPrompt(c *gin.Context) {
	name := c.Param("name")
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	var template PromptTemplate
	if err := s.db.Where("org_id = ? AND name = ?", orgID, name).
		Order("version DESC").First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// GetPromptVersion retrieves a specific version of a prompt
func (s *Service) GetPromptVersion(c *gin.Context) {
	name := c.Param("name")
	versionStr := c.Param("version")
	
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	var template PromptTemplate
	if err := s.db.Where("org_id = ? AND name = ? AND version = ?", orgID, name, version).
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// EvaluatePrompt runs evaluation on a prompt template
func (s *Service) EvaluatePrompt(c *gin.Context) {
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify prompt exists
	var template PromptTemplate
	if err := s.db.Where("id = ?", req.PromptID).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "prompt not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify suite exists
	var suite PromptSuite
	if err := s.db.Where("id = ?", req.SuiteID).First(&suite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "evaluation suite not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: Run actual evaluation
	result := EvaluationResult{
		ID:          uuid.New(),
		PromptID:    template.ID,
		SuiteID:     suite.ID,
		Score:       0.85, // placeholder
		PassedCases: 8,
		TotalCases:  10,
		Details:     []CaseResult{}, // placeholder
	}

	c.JSON(http.StatusOK, result)
}

// DeployPrompt creates or updates a prompt deployment
func (s *Service) DeployPrompt(c *gin.Context) {
	var req DeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	// Verify stable version exists
	var template PromptTemplate
	if err := s.db.Where("org_id = ? AND name = ? AND version = ?", 
		orgID, req.PromptName, req.StableVersion).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "stable version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Verify canary version exists if specified
	if req.CanaryVersion != nil {
		var canaryTemplate PromptTemplate
		if err := s.db.Where("org_id = ? AND name = ? AND version = ?", 
			orgID, req.PromptName, *req.CanaryVersion).First(&canaryTemplate).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "canary version not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Create or update deployment
	deployment := PromptDeployment{
		OrgID:         orgID,
		PromptName:    req.PromptName,
		StableVersion: req.StableVersion,
		CanaryVersion: req.CanaryVersion,
		CanaryRatio:   req.CanaryRatio,
	}

	// Use UPSERT pattern
	if err := s.db.Where("org_id = ? AND prompt_name = ?", orgID, req.PromptName).
		Assign(&deployment).FirstOrCreate(&deployment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

// GetDeployment retrieves current deployment for a prompt
func (s *Service) GetDeployment(c *gin.Context) {
	name := c.Param("name")
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // placeholder

	var deployment PromptDeployment
	if err := s.db.Where("org_id = ? AND prompt_name = ?", orgID, name).
		First(&deployment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployment)
}