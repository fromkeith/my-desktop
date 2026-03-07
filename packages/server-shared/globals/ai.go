package globals

import (
	"context"

	"google.golang.org/genai"
)

var (
	geminiClient *genai.Client
)

func Gemini() *genai.Client {
	if geminiClient != nil {
		return geminiClient
	}
	var err error
	// gets api key from env
	geminiClient, err = genai.NewClient(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	return geminiClient
}
