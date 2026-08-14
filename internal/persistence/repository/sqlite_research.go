package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ilyaskhan/term-agent/internal/persistence"
	"github.com/ilyaskhan/term-agent/internal/workflows/research/domain"
)

var (
	ErrResearchProjectNotFound  = errors.New("research project not found")
	ErrResearchQuestionNotFound = errors.New("research question not found")
	ErrResearchPaperNotFound    = errors.New("research paper not found")
)

// SQLiteResearchRepository implements persistence operations for Research Workflows.
type SQLiteResearchRepository struct {
	db *persistence.DB
}

// NewSQLiteResearchRepository constructs a new SQLiteResearchRepository.
func NewSQLiteResearchRepository(db *persistence.DB) *SQLiteResearchRepository {
	return &SQLiteResearchRepository{db: db}
}

// SaveProject persists a research project entity.
func (r *SQLiteResearchRepository) SaveProject(ctx context.Context, p *domain.ResearchProject) error {
	if p == nil || p.ID == "" {
		return fmt.Errorf("invalid research project")
	}

	query := `
		INSERT INTO research_projects (id, session_id, title, objective, template_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			objective = excluded.objective,
			template_id = excluded.template_id,
			status = excluded.status,
			updated_at = excluded.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.SessionID, p.Title, p.Objective, p.TemplateID, string(p.Status), p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save research project: %w", err)
	}
	return nil
}

// GetProject retrieves a research project by ID.
func (r *SQLiteResearchRepository) GetProject(ctx context.Context, id string) (*domain.ResearchProject, error) {
	query := `
		SELECT id, session_id, title, objective, template_id, status, created_at, updated_at
		FROM research_projects
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var p domain.ResearchProject
	var statusStr string
	err := row.Scan(&p.ID, &p.SessionID, &p.Title, &p.Objective, &p.TemplateID, &statusStr, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrResearchProjectNotFound
		}
		return nil, fmt.Errorf("failed to get research project: %w", err)
	}
	p.Status = domain.ProjectStatus(statusStr)
	return &p, nil
}

// SaveQuestion persists a research question.
func (r *SQLiteResearchRepository) SaveQuestion(ctx context.Context, q *domain.ResearchQuestion) error {
	if q == nil || q.ID == "" {
		return fmt.Errorf("invalid research question")
	}

	query := `
		INSERT INTO research_questions (id, project_id, question, scope, priority, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			question = excluded.question,
			scope = excluded.scope,
			priority = excluded.priority,
			status = excluded.status
	`
	_, err := r.db.ExecContext(ctx, query,
		q.ID, q.ProjectID, q.Question, q.Scope, q.Priority, string(q.Status), q.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save research question: %w", err)
	}
	return nil
}

// ListQuestions retrieves all research questions for a given project.
func (r *SQLiteResearchRepository) ListQuestions(ctx context.Context, projectID string) ([]*domain.ResearchQuestion, error) {
	query := `
		SELECT id, project_id, question, scope, priority, status, created_at
		FROM research_questions
		WHERE project_id = ?
		ORDER BY priority ASC, created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list research questions: %w", err)
	}
	defer rows.Close()

	var questions []*domain.ResearchQuestion
	for rows.Next() {
		var q domain.ResearchQuestion
		var statusStr string
		if err := rows.Scan(&q.ID, &q.ProjectID, &q.Question, &q.Scope, &q.Priority, &statusStr, &q.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}
		q.Status = domain.QuestionStatus(statusStr)
		questions = append(questions, &q)
	}
	return questions, rows.Err()
}

// SaveFinding persists a structured finding entity.
func (r *SQLiteResearchRepository) SaveFinding(ctx context.Context, f *domain.ResearchFinding) error {
	if f == nil || f.ID == "" {
		return fmt.Errorf("invalid research finding")
	}
	if err := f.Validate(); err != nil {
		return fmt.Errorf("invalid research finding payload: %w", err)
	}

	payloadJSON, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("failed to marshal research finding: %w", err)
	}

	query := `
		INSERT INTO research_findings (id, project_id, question_id, task_id, agent_type, payload_json, confidence, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			payload_json = excluded.payload_json,
			confidence = excluded.confidence
	`
	_, err = r.db.ExecContext(ctx, query,
		f.ID, f.ProjectID, f.QuestionID, f.TaskID, f.AgentType, string(payloadJSON), f.Confidence, f.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save research finding: %w", err)
	}
	return nil
}

// ListFindings retrieves all findings for a research project.
func (r *SQLiteResearchRepository) ListFindings(ctx context.Context, projectID string) ([]*domain.ResearchFinding, error) {
	query := `
		SELECT payload_json
		FROM research_findings
		WHERE project_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list research findings: %w", err)
	}
	defer rows.Close()

	var findings []*domain.ResearchFinding
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			return nil, fmt.Errorf("failed to scan finding row: %w", err)
		}
		var f domain.ResearchFinding
		if err := json.Unmarshal([]byte(rawJSON), &f); err != nil {
			return nil, fmt.Errorf("failed to unmarshal finding JSON: %w", err)
		}
		findings = append(findings, &f)
	}
	return findings, rows.Err()
}

// SavePaper persists the final research paper document.
func (r *SQLiteResearchRepository) SavePaper(ctx context.Context, p *domain.ResearchPaper) error {
	if p == nil || p.ID == "" {
		return fmt.Errorf("invalid research paper")
	}

	paperJSON, err := json.Marshal(p.Sections)
	if err != nil {
		return fmt.Errorf("failed to marshal paper sections: %w", err)
	}

	query := `
		INSERT INTO research_papers (id, project_id, template_id, title, paper_json, markdown_output, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			paper_json = excluded.paper_json,
			markdown_output = excluded.markdown_output,
			status = excluded.status,
			updated_at = excluded.updated_at
	`
	_, err = r.db.ExecContext(ctx, query,
		p.ID, p.ProjectID, p.TemplateID, p.Title, string(paperJSON), p.MarkdownOutput, string(p.Status), p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save research paper: %w", err)
	}
	return nil
}

// GetPaperByProject retrieves the research paper for a project.
func (r *SQLiteResearchRepository) GetPaperByProject(ctx context.Context, projectID string) (*domain.ResearchPaper, error) {
	query := `
		SELECT id, project_id, template_id, title, paper_json, markdown_output, status, created_at, updated_at
		FROM research_papers
		WHERE project_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, projectID)

	var p domain.ResearchPaper
	var paperJSON, statusStr string
	err := row.Scan(&p.ID, &p.ProjectID, &p.TemplateID, &p.Title, &paperJSON, &p.MarkdownOutput, &statusStr, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrResearchPaperNotFound
		}
		return nil, fmt.Errorf("failed to get research paper: %w", err)
	}
	p.Status = domain.PaperStatus(statusStr)

	if paperJSON != "" {
		var sections []domain.PaperSection
		if err := json.Unmarshal([]byte(paperJSON), &sections); err == nil {
			p.Sections = sections
		}
	}

	return &p, nil
}
