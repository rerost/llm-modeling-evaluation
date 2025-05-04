package akinator

import (
	"context"
	"fmt"
)

func NewHuman() Akinator {
	return &humanImpl{}
}

type humanImpl struct{}

func (human *humanImpl) Question(ctx context.Context) (string, error) {
	fmt.Print("質問: ")

	var question string
	fmt.Scanln(&question)

	return question, nil
}

func (human *humanImpl) SetAnswer(_, answer string) error {
	fmt.Printf("回答: %s\n", answer)
	return nil
}

func (human *humanImpl) ModelName() string {
	return "human"
}

func (human *humanImpl) ProvidorName() string {
	return ProvidorHuman
}
