package evaluate

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/rerost/llm-modeling-evaluation/pkg/akinator"
	"github.com/rerost/llm-modeling-evaluation/pkg/logger"
	"github.com/rerost/llm-modeling-evaluation/pkg/player"
)

// ある質問に対しての評価
type EvaluateResult struct {
	ModelName     string `csv:"model_name"`
	Answer        string `csv:"answer"`
	QuestionCount int    `csv:"question_count"`
	// 回答することができたか？
	Finished bool `csv:"finished"`
}

func Evaluate(
	ctx context.Context,
	model string,
	answer string,
	akinator akinator.Akinator,
	p player.Player,
	logger *logger.Logger,
) (*EvaluateResult, error) {
	loopCount := 0
	var res *player.Result
	for i := 0; i < 20; i++ {
		loopCount++
		question, err := akinator.Question(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		logger.Write(fmt.Sprintf("question(model: %s, ans: %s, num:%d): %s", akinator.ModelName(), answer, loopCount, question))

		res, err = p.Answer(ctx, answer, question)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		logger.Write(fmt.Sprintf("answer(ans: %s, num:%d): %v\n  finished?: %v\n  comment: %s", answer, loopCount, res.IsYes, res.Finished, res.Comment))

		if res.Finished {
			break
		}

		if res.IsYes {
			if err := akinator.SetAnswer(question, "Yes"); err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err := akinator.SetAnswer(question, "No"); err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}
	logger.Write("")

	return &EvaluateResult{
		ModelName:     akinator.ModelName(),
		Answer:        answer,
		QuestionCount: loopCount,
		Finished:      res.Finished,
	}, nil
}
