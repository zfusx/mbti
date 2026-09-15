package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected DATABASE_URL validation error")
	}
}

func TestLoadAcceptsValidConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://example.invalid/mbti")
	t.Setenv("QUESTION_COUNT", "92")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.QuestionCount != 92 {
		t.Fatalf("QuestionCount = %d, want 92", cfg.QuestionCount)
	}
}

func TestLoadRejectsUnbalancedQuestionCount(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://example.invalid/mbti")
	t.Setenv("QUESTION_COUNT", "91")
	if _, err := Load(); err == nil {
		t.Fatal("expected QUESTION_COUNT divisibility error")
	}
}
