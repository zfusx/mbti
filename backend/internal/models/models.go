package models

// Option is a single answer choice for a question.
type Option struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// Question represents an MBTI question record.
type Question struct {
	ID        int      `json:"id"`
	Dimension string   `json:"dimension"`
	Text      string   `json:"text"`
	Options   []Option `json:"options"`
}

// Answer captures a user's response to a question.
type Answer struct {
	QuestionID int    `json:"questionId"`
	Value      string `json:"value"`
}

// ResultDocument mirrors the stored JSONB payload for MBTI type descriptions.
type ResultDocument struct {
	Type        string   `json:"type"`
	Nickname    string   `json:"nickname"`
	Image       string   `json:"image"`
	Ratio       string   `json:"ratio"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Matches     []string `json:"matches"`
	Careers     []string `json:"careers"`
}
