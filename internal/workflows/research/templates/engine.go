package templates

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

//go:embed academic_research.json
var academicResearchJSON []byte

//go:embed technical_survey.json
var technicalSurveyJSON []byte

// SectionSpec describes a required paper section defined in a template.
type SectionSpec struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Template represents a global research paper outline specification.
type Template struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Sections    []SectionSpec `json:"sections"`
}

// TemplateEngine manages research templates and validates generated papers against them.
type TemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]Template
}

// NewTemplateEngine initializes a TemplateEngine loaded with default embedded templates.
func NewTemplateEngine() (*TemplateEngine, error) {
	eng := &TemplateEngine{
		templates: make(map[string]Template),
	}

	if err := eng.LoadTemplateBytes(academicResearchJSON); err != nil {
		return nil, fmt.Errorf("failed to load embedded academic_research template: %w", err)
	}
	if err := eng.LoadTemplateBytes(technicalSurveyJSON); err != nil {
		return nil, fmt.Errorf("failed to load embedded technical_survey template: %w", err)
	}

	return eng, nil
}

// LoadTemplateBytes parses and registers a JSON template spec.
func (e *TemplateEngine) LoadTemplateBytes(data []byte) error {
	var t Template
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("failed to unmarshal template JSON: %w", err)
	}
	if t.ID == "" {
		return errors.New("template ID cannot be empty")
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.templates[t.ID] = t
	return nil
}

// GetTemplate retrieves a template by ID.
func (e *TemplateEngine) GetTemplate(id string) (Template, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	t, ok := e.templates[id]
	if !ok {
		return Template{}, fmt.Errorf("template not found: %s", id)
	}
	return t, nil
}

// CreatePaperSkeleton generates empty PaperSection structs conforming to the template specs.
func (e *TemplateEngine) CreatePaperSkeleton(templateID string) ([]domain.PaperSection, error) {
	t, err := e.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	var sections []domain.PaperSection
	for i, spec := range t.Sections {
		sections = append(sections, domain.PaperSection{
			ID:        spec.ID,
			Title:     spec.Title,
			Content:   "",
			Order:     i + 1,
			Citations: []domain.Citation{},
		})
	}
	return sections, nil
}

// ValidatePaperCompleteness checks if all required template sections contain content.
func (e *TemplateEngine) ValidatePaperCompleteness(paper *domain.ResearchPaper) error {
	if paper == nil {
		return errors.New("paper is nil")
	}
	t, err := e.GetTemplate(paper.TemplateID)
	if err != nil {
		return err
	}

	sectionMap := make(map[string]domain.PaperSection)
	for _, sec := range paper.Sections {
		sectionMap[sec.ID] = sec
	}

	for _, reqSec := range t.Sections {
		if reqSec.Required {
			sec, exists := sectionMap[reqSec.ID]
			if !exists || len(sec.Content) == 0 {
				return fmt.Errorf("required section '%s' (%s) is missing or empty", reqSec.Title, reqSec.ID)
			}
		}
	}
	return nil
}
