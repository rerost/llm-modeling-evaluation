package akinator

import (
	"context"
	"os"
)

type Akinator interface {
	Question(ctx context.Context) (string, error)
	SetAnswer(question string, answer string) error
	ModelName() string
	ProvidorName() string
}

// 回答済みの質問と回答
type Answered struct {
	Question string
	Answer   string
}

func New(providor, model string) Akinator {
	switch providor {
	case ProvidorOpenAI:
		return NewOpenAI(model, os.Getenv("OPENAI_API_KEY"))
	case ProvidorHuman:
		return NewHuman()
	default:
		panic("unknown providor")
	}
}

const ProvidorOpenAI = "openai"
const ProvidorHuman = "human"
