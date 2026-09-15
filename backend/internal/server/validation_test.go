package server

import (
	"testing"

	"github.com/zfusx/mbti/internal/models"
)

func balancedFixture() ([]models.Answer, []models.Question) {
	dimensions := []string{"EI", "SN", "TF", "JP"}
	answers := make([]models.Answer, 0, 20)
	questions := make([]models.Question, 0, 20)
	id := 1
	for _, dimension := range dimensions {
		for range 5 {
			questions = append(questions, models.Question{
				ID:        id,
				Dimension: dimension,
				Options: []models.Option{
					{Text: "left", Value: dimension[:1]},
					{Text: "right", Value: dimension[1:]},
				},
			})
			answers = append(answers, models.Answer{QuestionID: id, Value: dimension[:1]})
			id++
		}
	}
	return answers, questions
}

func TestValidateAnswersAcceptsBalancedKnownQuestions(t *testing.T) {
	answers, questions := balancedFixture()
	if err := validateAnswers(answers, questions); err != nil {
		t.Fatalf("validateAnswers() error = %v", err)
	}
}

func TestValidateAnswersRejectsInvalidOption(t *testing.T) {
	answers, questions := balancedFixture()
	answers[0].Value = "I"
	questions[0].Options = questions[0].Options[:1]
	if err := validateAnswers(answers, questions); err == nil {
		t.Fatal("expected invalid option error")
	}
}

func TestValidateAnswersRejectsUnbalancedDimensions(t *testing.T) {
	answers, questions := balancedFixture()
	questions[0].Dimension = "SN"
	if err := validateAnswers(answers, questions); err == nil {
		t.Fatal("expected unbalanced dimensions error")
	}
}

func TestValidMBTIType(t *testing.T) {
	for _, value := range []string{"ENFP", "ISTJ"} {
		if !validMBTIType(value) {
			t.Fatalf("validMBTIType(%q) = false", value)
		}
	}
	for _, value := range []string{"ESTX", "AAAA", "ENF", "ENFPX"} {
		if validMBTIType(value) {
			t.Fatalf("validMBTIType(%q) = true", value)
		}
	}
}
