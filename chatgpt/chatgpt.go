package chatgpt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	apiURL = "https://api.openai.com/v1/engines/davinci-codex/completions"
)

type CompletionRequest struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
}

type CompletionResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

type ChatGpt struct {
	client *resty.Client
	apiKey string
}

func NewChatGpt(apiKey string) *ChatGpt {
	return &ChatGpt{
		client: resty.New(),
		apiKey: apiKey,
	}
}

func (c *ChatGpt) Prompt(imageUrl string) (bool, string, error) {
	requestBody := CompletionRequest{
		Prompt: imageUrl + " I have an satellite image on kamchatka territory. " +
			"If on this image is any kind of felling, tell me FELLING. If on this image is some kind of burning, tell me BURNING." +
			"If none of this, tell me NONE.",
		MaxTokens:   50,
		Temperature: 0.7,
		TopP:        1.0,
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.apiKey)).
		SetBody(requestBody).
		Post(apiURL)

	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return false, "", errors.New("unexpected status code: " + strconv.Itoa(resp.StatusCode()))
	}

	var response CompletionResponse
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return false, "", errors.New("error parsing " + err.Error())
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Generated text: %s\n", response.Choices[0].Text)
		if strings.Contains(strings.ToLower(response.Choices[0].Text), "felling") {
			return true, "felling", nil
		} else if strings.Contains(strings.ToLower(response.Choices[0].Text), "burning") {
			return true, "burning", nil
		}
		return false, "", errors.New("not found any troubles on the image")
	} else {
		fmt.Println("No response choices found.")
		return false, "", errors.New("no response choices found")
	}
}
