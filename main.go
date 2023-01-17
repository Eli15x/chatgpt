package main

import (
	"context"
	"fmt"
	"encoding/json"
	"github.com/slack-go/slack"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack/slackevents"

	gpt3 "github.com/PullRequestInc/go-gpt3"
)

func GetResponse(client gpt3.Client, ctx context.Context, quesiton string) (string, error) {
	
	answer := ""
	err := client.CompletionStreamWithEngine(ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
		Prompt: []string{
			quesiton,
		},
		MaxTokens:   gpt3.IntPtr(3000),
		Temperature: gpt3.Float32Ptr(0),
	}, func(resp *gpt3.CompletionResponse) {
		answer = resp.Choices[0].Text
	})
	if err != nil {
		return "",err
	}
	return answer,nil
}

func main() {
	api := slack.New("xoxb-4598852757735-4643514435328-mjgTygAdLhBWU1lKwv5wMFg4")

 	///chatgpt
	apiKey := "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"
	if apiKey == "" {
		panic("Missing API KEY")
	}

	ctx := context.Background()
	client := gpt3.NewClient(apiKey)
	signingSecret := "a177a841e99538e563d2b0987010bffb"
	
	http.HandleFunc("/events-endpoint", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == "app_mention" {
			data := eventsAPIEvent.Data //ver oque vem na interface
			fmt.Println(data)
			response, err := GetResponse(client, ctx, "")
			text := ""

			if err != nil {
				text = "Desculpe, ocorreu um erro na sua resposta, tente novamente"
			} else {
				text = response
			}

			attachment := slack.Attachment{
				Text: text,
			}
			_, _, err = api.PostMessage(
				"C04J3R7USLC",
				slack.MsgOptionAttachments(attachment),
			)
		
		}
	})
	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}