package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gocarina/gocsv"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %+v", err)
		os.Exit(1)
	}
}

func playerTest() {
	ctx := context.Background()
	player := NewPlayer()
	fmt.Println(player.Answer(ctx, "ピザ", "家にありますか？"))
}

func run() error {
	modelsRaw := os.Getenv("MODELS")
	models := strings.Split(modelsRaw, ",")

	answers := []string{"ピザ", "北海道", "宇宙飛行士", "Twitter", "ラーメン二郎"}
	ctx := context.Background()

	var eg errgroup.Group
	var resultChan = make(chan Evaluate)
	var eval []Evaluate

	done := make(chan struct{})
	go func() {
		for result := range resultChan {
			eval = append(eval, result)
		}
		close(done)
	}()

	for _, model := range models {
		eg.Go(func() error {
			result, err := evaluate(ctx, model, answers)
			if err != nil {
				return errors.WithStack(err)
			}

			for _, result := range result {
				resultChan <- result
			}
			return nil
		})
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

type Evaluate struct {
	ModelName     string `csv:"model_name"`
	Answer        string `csv:"answer"`
	QuestionCount int    `csv:"question_count"`
	// 回答することができたか？
	Finished bool `csv:"finished"`
}

func evaluate(ctx context.Context, model string, answers []string) ([]Evaluate, error) {
	result := make([]Evaluate, 0, len(answers))
	for _, answer := range answers {
		fmt.Println("ANSWER = ", answer)
		akinator := NewAkinator(model)
		loopCount := 0
		var res *Result
		for i := 0; i < 20; i++ {
			var stream string
			loopCount++
			question, err := akinator.Question(ctx)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			stream = stream + fmt.Sprintf("question(model: %s, ans: %s, num:%d): %s\n", akinator.ModelName(), answer, loopCount, question)

			player := NewPlayer()
			res, err = player.Answer(ctx, answer, question)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			stream = stream + fmt.Sprintf("answer(ans: %s, num:%d): %v\n  finished?: %v\n  comment: %s\n", answer, loopCount, res.IsYes, res.Finished, res.Comment)

			fmt.Print(stream)
			if res.Finished {
				break
			}

			if res.IsYes {
				akinator.SetAnswer(question, "Yes")
			} else {
				akinator.SetAnswer(question, "No")
			}
		}
		fmt.Println("")

		result = append(result, Evaluate{
			ModelName:     akinator.ModelName(),
			Answer:        answer,
			QuestionCount: loopCount,
			Finished:      res.Finished,
		})
	}

	return result, nil
}

// 回答済みの質問と回答
type Answered struct {
	Question string
	Answer   string
}

// 質問をする側
type Akinator struct {
	answered []Answered

	client *openai.Client
	model  string
}

func NewAkinator(model string) *Akinator {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	return &Akinator{
		client:   client,
		model:    model,
		answered: []Answered{},
	}
}

func (akinator *Akinator) Question(ctx context.Context) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "20の質問をします。相手はYes / Noを返してくるので1行で質問もしくは回答をしてください。正解するまで質問もしくは回答を続けてください",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "私が何を考えているか当ててください",
		},
	}

	for _, answered := range akinator.answered {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: answered.Question,
		})
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: answered.Answer,
		})
	}

	response, err := akinator.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    akinator.ModelName(),
		Messages: messages,
	})

	if err != nil {
		return "", errors.WithStack(err)
	}

	return response.Choices[0].Message.Content, nil
}

func (akinator *Akinator) SetAnswer(question, answer string) {
	akinator.answered = append(akinator.answered, Answered{Question: question, Answer: answer})
}

func (akinator *Akinator) ModelName() string {
	return akinator.model
}
