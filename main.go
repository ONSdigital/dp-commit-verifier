package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-ci/commit-verification/handler"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/pat"
)

var (
	bindAddr = os.Getenv("BIND_ADDR")
	slackURL = os.Getenv("SLACK_URL")
)

func main() {
	if len(slackURL) < 1 {
		log.Error(errors.New("no Slack URL provided"), nil)
		os.Exit(1)
	}
	if len(bindAddr) < 1 {
		bindAddr = ":3000"
	}

	h := &handler.Webhook{SlackURL: slackURL}
	r := pat.New()
	r.Path("/").Methods("POST").HandlerFunc(h.Handle)

	if err := http.ListenAndServe(bindAddr, r); err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}
