package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/gocarina/gocsv"
	openai "github.com/sashabaranov/go-openai"
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
	answers := []string{"ピザ", "北海道", "宇宙飛行士", "Twitter", "ラーメン二郎"}

	ctx := context.Background()

	result, err := evaluate(ctx, answers)
	if err != nil {
		return errors.WithStack(err)
	}

	file, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := gocsv.MarshalFile(&result, file); err != nil {
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

func evaluate(ctx context.Context, answers []string) ([]Evaluate, error) {
	result := make([]Evaluate, 0, len(answers))
	for _, answer := range answers {
		akinator := NewAkinator()
		loopCount := 0
		for i := 0; i < 20; i++ {
			loopCount++
			question, err := akinator.Question(ctx)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			fmt.Println("question: ", question)

			player := NewPlayer()
			res, err := player.Answer(ctx, answer, question)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			fmt.Printf("answer(yes: true, no: false): %v\n  finished?: %v\n  comment: %s\n", res.IsYes, res.Finished, res.Comment)
			if res.Finished {
				break
			}

			if res.IsYes {
				akinator.SetAnswer(question, "Yes")
			} else {
				akinator.SetAnswer(question, "No")
			}
		}

		result = append(result, Evaluate{
			ModelName:     akinator.ModelName(),
			Answer:        answer,
			QuestionCount: loopCount,
			Finished:      true,
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
}

func NewAkinator() *Akinator {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	return &Akinator{
		client:   client,
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
	return openai.GPT4Dot1
}
