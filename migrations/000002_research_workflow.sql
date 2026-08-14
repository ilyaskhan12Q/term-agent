-- Migration 000002: Research Workflow Persistence Schema
PRAGMA foreign_keys = ON;

-- 1. Research Projects Table
CREATE TABLE IF NOT EXISTS research_projects (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    title TEXT NOT NULL,
    objective TEXT NOT NULL,
    template_id TEXT NOT NULL DEFAULT 'academic_research',
    status TEXT NOT NULL DEFAULT 'CREATED',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_research_projects_session ON research_projects(session_id);

-- 2. Research Questions Table
CREATE TABLE IF NOT EXISTS research_questions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    question TEXT NOT NULL,
    scope TEXT NOT NULL DEFAULT '',
    priority INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'PENDING',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_research_questions_project ON research_questions(project_id);

-- 3. Research Sources Table
CREATE TABLE IF NOT EXISTS research_sources (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    uri TEXT NOT NULL DEFAULT '',
    authors TEXT NOT NULL DEFAULT '[]', -- JSON array of author strings
    year INT NOT NULL DEFAULT 0,
    publisher TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT 'ACADEMIC_PAPER',
    trust_score REAL NOT NULL DEFAULT 1.0,
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_research_sources_project ON research_sources(project_id);

-- 4. Research Evidence Table
CREATE TABLE IF NOT EXISTS research_evidence (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    source_id TEXT NOT NULL,
    snippet TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    verification_status TEXT NOT NULL DEFAULT 'UNVERIFIED',
    extractor_agent_id TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE,
    FOREIGN KEY (source_id) REFERENCES research_sources(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_research_evidence_project ON research_evidence(project_id);
CREATE INDEX IF NOT EXISTS idx_research_evidence_source ON research_evidence(source_id);

-- 5. Research Claims Table
CREATE TABLE IF NOT EXISTS research_claims (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    statement TEXT NOT NULL,
    strength TEXT NOT NULL DEFAULT 'DIRECT',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_research_claims_project ON research_claims(project_id);

-- 6. Claim Evidence Junction Table
CREATE TABLE IF NOT EXISTS claim_evidence_map (
    claim_id TEXT NOT NULL,
    evidence_id TEXT NOT NULL,
    PRIMARY KEY (claim_id, evidence_id),
    FOREIGN KEY (claim_id) REFERENCES research_claims(id) ON DELETE CASCADE,
    FOREIGN KEY (evidence_id) REFERENCES research_evidence(id) ON DELETE CASCADE
);

-- 7. Structured Findings Table
CREATE TABLE IF NOT EXISTS research_findings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    question_id TEXT NOT NULL,
    task_id TEXT NOT NULL DEFAULT '',
    agent_type TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    confidence REAL NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES research_questions(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_research_findings_project ON research_findings(project_id);
CREATE INDEX IF NOT EXISTS idx_research_findings_question ON research_findings(question_id);

-- 8. Final Research Papers Table
CREATE TABLE IF NOT EXISTS research_papers (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL UNIQUE,
    template_id TEXT NOT NULL DEFAULT 'academic_research',
    title TEXT NOT NULL,
    paper_json TEXT NOT NULL DEFAULT '{}',
    markdown_output TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'DRAFT',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES research_projects(id) ON DELETE CASCADE
);
