package scoring

import (
	"fmt"
	"math"
	"strings"

	"github.com/zfusx/mbti/internal/models"
)

var (
	dimensions = []string{"E", "I", "S", "N", "T", "F", "J", "P"}

	pairs = []struct {
		Label string
		Left  string
		Right string
	}{
		{Label: "EI", Left: "E", Right: "I"},
		{Label: "SN", Left: "S", Right: "N"},
		{Label: "TF", Left: "T", Right: "F"},
		{Label: "JP", Left: "J", Right: "P"},
	}
)

// Service encapsulates scoring logic and metadata.
type Service struct {
	totalQuestions int
}

// Result represents the payload returned to clients after scoring.
type Result struct {
	Type       string                   `json:"type"`
	Tally      map[string]int           `json:"tally"`
	Pairs      map[string]PairBreakDown `json:"pairs"`
	Completion float64                  `json:"completion"`
	Confidence Confidence               `json:"confidence"`
}

// PairBreakDown exposes per-pair statistics.
type PairBreakDown struct {
	Left       string `json:"left"`
	Right      string `json:"right"`
	LeftScore  int    `json:"leftScore"`
	RightScore int    `json:"rightScore"`
	LeftPct    int    `json:"leftPct"`
	RightPct   int    `json:"rightPct"`
	Gap        int    `json:"gap"`
	Winner     string `json:"winner"`
}

// Confidence communicates the minimum pair gap which acts as a proxy for certainty.
type Confidence struct {
	MinPairGap int    `json:"minPairGap"`
	Rule       string `json:"rule"`
}

// NewService creates a scoring service with the configured question count.
func NewService(totalQuestions int) (*Service, error) {
	if totalQuestions <= 0 {
		return nil, fmt.Errorf("totalQuestions must be positive")
	}

	return &Service{totalQuestions: totalQuestions}, nil
}

// Score processes answer selections and returns an aggregate result.
func (s *Service) Score(answers []models.Answer, expectedQuestions int) (Result, error) {
	if len(answers) == 0 {
		return Result{}, fmt.Errorf("answers cannot be empty")
	}
	if expectedQuestions <= 0 || expectedQuestions > s.totalQuestions {
		return Result{}, fmt.Errorf("expected question count is invalid")
	}
	if len(answers) > expectedQuestions {
		return Result{}, fmt.Errorf("answer count exceeds expected question count")
	}

	tally := make(map[string]int, len(dimensions))
	for _, dim := range dimensions {
		tally[dim] = 0
	}

	for _, ans := range answers {
		val := strings.ToUpper(ans.Value)
		if _, ok := tally[val]; !ok {
			return Result{}, fmt.Errorf("invalid answer value %q", ans.Value)
		}
		tally[val]++
	}

	pairMap := make(map[string]PairBreakDown, len(pairs))
	var builder strings.Builder
	minGap := math.MaxInt

	for _, pair := range pairs {
		leftScore := tally[pair.Left]
		rightScore := tally[pair.Right]
		winner := pair.Left
		if rightScore > leftScore {
			winner = pair.Right
		}

		total := leftScore + rightScore
		leftPct, rightPct := 0, 0
		if total > 0 {
			leftPct = int(math.Round(float64(leftScore) / float64(total) * 100))
			rightPct = int(math.Round(float64(rightScore) / float64(total) * 100))
		}

		gap := int(math.Abs(float64(leftScore - rightScore)))
		if gap < minGap {
			minGap = gap
		}

		pairMap[pair.Label] = PairBreakDown{
			Left:       pair.Left,
			Right:      pair.Right,
			LeftScore:  leftScore,
			RightScore: rightScore,
			LeftPct:    leftPct,
			RightPct:   rightPct,
			Gap:        gap,
			Winner:     winner,
		}

		builder.WriteString(winner)
	}

	completion := math.Round(float64(len(answers))/float64(expectedQuestions)*1000) / 10
	if minGap == math.MaxInt {
		minGap = 0
	}

	return Result{
		Type:       builder.String(),
		Tally:      tally,
		Pairs:      pairMap,
		Completion: completion,
		Confidence: Confidence{
			MinPairGap: minGap,
			Rule:       "tie→prefer-left",
		},
	}, nil
}
