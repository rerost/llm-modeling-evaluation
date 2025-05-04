package akinator

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/sashabaranov/go-openai"
)

func NewOpenAI(model string, apiKey string) Akinator {
	client := openai.NewClient(apiKey)
	return &openAIAkinator{
		client: client,
		model:  model,
	}
}

type openAIAkinator struct {
	answered []Answered

	client   *openai.Client
	model    string
	providor string
}

func (akinator *openAIAkinator) Question(ctx context.Context) (string, error) {
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

func (akinator *openAIAkinator) SetAnswer(question, answer string) error {
	akinator.answered = append(akinator.answered, Answered{Question: question, Answer: answer})
	return nil
}

func (akinator *openAIAkinator) ProvidorName() string {
	return akinator.providor
}

func (akinator *openAIAkinator) ModelName() string {
	return akinator.model
}
