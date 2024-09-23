package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type webhookPayload struct {
	Text   string  `json:"text"`
	Blocks []Block `json:"blocks"`
}

type Element struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

type Block struct {
	Type     string    `json:"type"`
	Text     *Element  `json:"text,omitempty"`
	Elements []Element `json:"elements,omitempty"`
}

func sendToSlack(url, text string, blocks []Block) error {
	p := webhookPayload{Text: text, Blocks: blocks}
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack webhook body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create Slack webhook request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do Slack webhook request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Slack: %s", resp.Status)
	}
	return resp.Body.Close()
}
