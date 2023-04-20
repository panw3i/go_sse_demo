package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

func main() {
	r := gin.Default()

	r.StaticFile("/", "./static/index.html")
	r.Static("/static", "./static")
	r.POST("/sse", sseHandler)

	r.Run(":8080")
}

type JsonRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Content   string `json:"content"`
}

func sseHandler(c *gin.Context) {

	var jsonRequest JsonRequest
	if err := c.ShouldBindJSON(&jsonRequest); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 从环境变量中获取API密钥
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		os.Exit(1)
	}
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     jsonRequest.Model,
		MaxTokens: jsonRequest.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: jsonRequest.Content,
			},
		},
		Stream: true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	c.Header("Access-Control-Allow-Origin", origin)
	for {
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}
		fmt.Printf(res.Choices[0].Delta.Content)
		c.SSEvent("message", res.Choices[0].Delta.Content)

	}
}
