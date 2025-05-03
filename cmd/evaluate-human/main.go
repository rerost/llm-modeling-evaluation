package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/gocarina/gocsv"
	"github.com/rerost/llm-modeling-evaluation/pkg/akinator"
	"github.com/rerost/llm-modeling-evaluation/pkg/answergenerator"
	"github.com/rerost/llm-modeling-evaluation/pkg/evaluate"
	"github.com/rerost/llm-modeling-evaluation/pkg/logger"
	"github.com/rerost/llm-modeling-evaluation/pkg/player"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %+v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	modelName := "rerost"
	answers, err := answergenerator.NewAnswerGenerator().Generate(ctx, 5)
	if err != nil {
		return errors.WithStack(err)
	}
	p := player.NewPlayer()
	akinator := akinator.New(akinator.ProvidorHuman, modelName)

	l := logger.NewLogger()
	logger.ShowAfterClose(l, os.Stdout)
	defer l.Close()

	results := make([]*evaluate.EvaluateResult, 0, len(answers))
	for _, answer := range answers {
		result, err := evaluate.Evaluate(
			context.Background(),
			modelName,
			answer,
			akinator,
			p,
			l,
		)
		fmt.Printf("答えは%s\n", answer)
		if err != nil {
			return errors.WithStack(err)
		}

		results = append(results, result)
	}

	file, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := gocsv.MarshalFile(&results, file); err != nil {
		panic(err)
	}

	return nil
}
