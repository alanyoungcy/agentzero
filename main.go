package main

import (
    "context"
    "fmt"
    "log"

    "github.com/joho/godotenv"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    // 加载 .env 文件
    err := godotenv.Load()
    if err != nil {
        fmt.Println("Error loading .env file")
        return
    }

    ctx := context.Background()
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }
    prompt := "What would be a good company name for a company that makes colorful socks?"
    completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(completion)
}

