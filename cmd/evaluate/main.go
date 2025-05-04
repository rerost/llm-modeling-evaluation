package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gocarina/gocsv"
	"github.com/rerost/llm-modeling-evaluation/pkg/akinator"
	"github.com/rerost/llm-modeling-evaluation/pkg/evaluate"
	"github.com/rerost/llm-modeling-evaluation/pkg/logger"
	"github.com/rerost/llm-modeling-evaluation/pkg/player"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %+v", err)
		os.Exit(1)
	}
}

func run() error {
	providor := akinator.ProvidorOpenAI
	modelsRaw := os.Getenv("MODELS")
	models := strings.Split(modelsRaw, ",")

	answers := []string{"ピザ", "北海道", "宇宙飛行士", "Twitter", "ラーメン二郎"}
	ctx := context.Background()

	var eg errgroup.Group
	var resultChan = make(chan evaluate.EvaluateResult)
	var eval []evaluate.EvaluateResult

	done := make(chan struct{})
	go func() {
		for result := range resultChan {
			eval = append(eval, result)
		}
		close(done)
	}()

	l := logger.NewLogger()
	logger.ShowImmediate(l, os.Stdout)
	defer l.Close()

	p := player.NewPlayer()
	for _, model := range models {
		for _, answer := range answers {
			eg.Go(func() error {
				akinator := akinator.New(providor, model)
				result, err := evaluate.Evaluate(ctx, model, answer, akinator, p, l)
				if err != nil {
					return errors.WithStack(err)
				}

				resultChan <- *result
				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}
	close(resultChan)

	<-done

	file, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := gocsv.MarshalFile(&eval, file); err != nil {
		panic(err)
	}

	return nil
}
