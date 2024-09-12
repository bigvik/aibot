package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/teilomillet/gollm"
)

func main() {
	botApi := "https://api.telegram.org/bot"
	botToken := os.Getenv("CONFIRMAT_BOT_TOKEN")
	botUrl := botApi + botToken
	offset := 0

	llm, err := gollm.NewLLM(
		gollm.SetProvider("ollama"),
		gollm.SetModel("llama3.1"),
		gollm.SetDebugLevel(gollm.LogLevelWarn),
	)
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	for {
		updates, err := getUpdates(botUrl, offset)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, update := range updates {
			prompt := gollm.NewPrompt(update.Message.Text)
			ctx := context.Background()
			response, err := llm.Generate(ctx, prompt)
			if err != nil {
				log.Fatalf("Failed to generate response: %v", err)
			}
			if err := respond(botUrl, update, response); err != nil {
				fmt.Println(err)
				continue
			}
			offset = update.UpdateId + 1
		}
	}
}

func getUpdates(botUrl string, offset int) ([]Update, error) {
	r, err := http.Get(botUrl + "/getUpdates?offset=" + fmt.Sprint(offset))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var restResponse RestResponse
	if err := json.Unmarshal(body, &restResponse); err != nil {
		return nil, err
	}
	return restResponse.Results, nil
}

func respond(botUrl string, update Update, response string) error {
	var botMessage BotMessage
	botMessage.ChatId = update.Message.Chat.ChatId
	botMessage.Text = response //update.Message.Text
	botMessageJson, err := json.Marshal(botMessage)
	if err != nil {
		return err
	}
	r, err := http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(botMessageJson))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}
