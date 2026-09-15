package scoring

import (
	"testing"

	"github.com/zfusx/mbti/internal/models"
)

func TestScoreUsesExpectedQuizSizeForCompletion(t *testing.T) {
	service, err := NewService(92)
	if err != nil {
		t.Fatal(err)
	}

	answers := make([]models.Answer, 0, 20)
	for id, value := range []string{
		"E", "E", "E", "I", "I",
		"S", "S", "S", "N", "N",
		"T", "T", "T", "F", "F",
		"J", "J", "J", "P", "P",
	} {
		answers = append(answers, models.Answer{QuestionID: id + 1, Value: value})
	}

	result, err := service.Score(answers, 20)
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "ESTJ" {
		t.Fatalf("type = %q, want ESTJ", result.Type)
	}
	if result.Completion != 100 {
		t.Fatalf("completion = %v, want 100", result.Completion)
	}
}

func TestScoreRejectsInvalidValue(t *testing.T) {
	service, err := NewService(92)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Score([]models.Answer{{QuestionID: 1, Value: "X"}}, 1)
	if err == nil {
		t.Fatal("expected invalid answer value error")
	}
}

func TestScoreRejectsInvalidExpectedCount(t *testing.T) {
	service, err := NewService(92)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Score([]models.Answer{{QuestionID: 1, Value: "E"}}, 93)
	if err == nil {
		t.Fatal("expected invalid expected question count error")
	}
}
