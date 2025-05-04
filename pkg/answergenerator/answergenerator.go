package answergenerator

// 20の質問の答を事前に生成する

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
)

type AnswerGenerator interface {
	Generate(ctx context.Context, size int) ([]string, error)
}

func NewAnswerGenerator() AnswerGenerator {
	return &gen{}
}

type gen struct{}

//go:embed question_schema.json
var questionSchema string

//go:embed system_prompt.txt
var systemPrompt string

func (p *gen) Generate(ctx context.Context, size int) ([]string, error) {
	prompt := fmt.Sprintf(`ランダムにそこそこ具体的な単語%d個あげて`, size)
	jsonEscapedPrompt, err := json.Marshal(prompt)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reqBody := fmt.Sprintf(`
  {
    "model": "gpt-4o",
    "input": [
      {
        "role": "user",
        "content": [
          {
            "type": "input_text",
            "text": %s
          }
        ]
      }
    ],
    "text": {
      "format": {
        "type": "json_schema",
        %s
      }
    },
    "reasoning": {},
    "tools": [],
    "temperature": 1,
    "max_output_tokens": 2048,
    "top_p": 1,
    "store": false
  }`, string(jsonEscapedPrompt), questionSchema)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, errors.Newf("status code: %d, body: %s", res.StatusCode, string(b))
	}

	response := &Response{}
	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return nil, errors.WithStack(err)
	}

	result := &Result{}
	if err := json.Unmarshal([]byte(response.Output[0].Content[0].Text), result); err != nil {
		return nil, errors.WithStack(err)
	}

	return result.Answers, nil
}

type Content struct {
	Text string `json:"text"`
}

type Output struct {
	Content []Content `json:"content"`
}
type Response struct {
	Output []Output `json:"output"`
}

type Result struct {
	Answers []string `json:"answers"`
}
