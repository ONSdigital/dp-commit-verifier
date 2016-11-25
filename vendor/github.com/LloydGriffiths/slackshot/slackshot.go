// Package slackshot provides opinionated functionality for Slack webhooks.
package slackshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Webhook represents a Slack webhook.
type Webhook struct {
	Endpoint string
}

// Payload represents a webhook payload.
type Payload struct {
	Colour  string
	Message *Message
}

// Message represents a payload message.
type Message struct {
	Headline string
	Summary  string
}

type attachment struct {
	Colour    string   `json:"color"`
	Markdowns []string `json:"mrkdwn_in"`
	Text      string   `json:"text"`
}

type attachments struct {
	Attachments []*attachment `json:"attachments"`
}

func (m *Message) String() string {
	if len(m.Summary) > 1 {
		return fmt.Sprintf("%s\n%s", m.Headline, m.Summary)
	}

	return m.Headline
}

// Send sends a payload to a webhook.
func (h *Webhook) Send(p *Payload) error {
	b, err := json.Marshal(&attachments{[]*attachment{{p.Colour, []string{"text"}, p.Message.String()}}})
	if err != nil {
		return err
	}

	q, err := http.NewRequest(http.MethodPost, h.Endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	r, err := http.DefaultClient.Do(q)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unknown slack error: %s", r.Status)
	}

	return nil
}
