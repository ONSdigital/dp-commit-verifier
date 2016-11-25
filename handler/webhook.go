// Package handler provides functionality for processing GitHub webhooks and
// generating alerts for unverified commits or unknown commit authors.
package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/LloydGriffiths/slackshot"
	"github.com/ONSdigital/dp-ci/commit-verification/identity"
	"github.com/ONSdigital/go-ns/log"
)

type author struct {
	Username string `json:"username"`
}

type commit struct {
	Author *author `json:"author"`
	ID     string  `json:"id"`
	URL    string  `json:"url"`
}

type owner struct {
	Name string `json:"name"`
}

type repository struct {
	Name  string `json:"name"`
	Owner *owner `json:"owner"`
	URL   string `json:"html_url"`
}

type payload struct {
	Commit     commit      `json:"head_commit"`
	Repository *repository `json:"repository"`
}

// Webhook represents a GitHub webhook handler.
type Webhook struct {
	SlackURL string
}

// Handle is an HTTP handler which processes a GitHub webhook to verify the
// identity of a commit author.
func (w *Webhook) Handle(wr http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.TraceR(req, fmt.Sprintf("error reading request body: %s", err), nil)
		wr.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var p payload
	if err = json.Unmarshal(b, &p); err != nil {
		log.TraceR(req, fmt.Sprintf("error unmarshaling payload: %s", err), nil)
		wr.WriteHeader(http.StatusBadRequest)
		return
	}

	v, err := identity.IsValid(p.Repository.Owner.Name, p.Repository.Name, p.Commit.ID)
	if err != nil {
		log.Debug(fmt.Sprintf("error validating signature: %s", err), nil)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}
	if v {
		log.Debug(fmt.Sprintf("valid signature from %s for %s:%s", p.Commit.Author.Username, p.Repository.Name, p.Commit.ID), nil)
		wr.WriteHeader(http.StatusOK)
		return
	}

	log.Debug(fmt.Sprintf("invalid signature from %s for %s:%s", p.Commit.Author.Username, p.Repository.Name, p.Commit.ID), nil)

	message := &slackshot.Payload{
		Colour: "danger",
		Message: &slackshot.Message{
			Headline: fmt.Sprintf("*Unverified commit from %s*", p.Commit.Author.Username),
			Summary:  fmt.Sprintf("_<%s>_", p.Commit.URL),
		},
	}

	if err := (&slackshot.Webhook{Endpoint: w.SlackURL}).Send(message); err != nil {
		log.Error(fmt.Errorf("error sending notification to Slack: %s", err), nil)
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	wr.WriteHeader(http.StatusBadRequest)
}
