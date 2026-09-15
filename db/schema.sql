-- PostgreSQL schema for MBTI92 platform
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS mbti_questions (
    id INT PRIMARY KEY,
    dimension TEXT NOT NULL,
    payload JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_mbti_questions_dimension ON mbti_questions (dimension);
CREATE INDEX IF NOT EXISTS idx_mbti_questions_payload_gin ON mbti_questions USING GIN (payload jsonb_path_ops);

CREATE TABLE IF NOT EXISTS mbti_results (
    type CHAR(4) PRIMARY KEY,
    payload JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_mbti_results_payload_gin ON mbti_results USING GIN (payload jsonb_path_ops);

CREATE TABLE IF NOT EXISTS mbti_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    answers JSONB NOT NULL,
    result JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mbti_sessions_created_at ON mbti_sessions (created_at);
CREATE INDEX IF NOT EXISTS idx_mbti_sessions_result_gin ON mbti_sessions USING GIN (result jsonb_path_ops);
