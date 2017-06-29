/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package github

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

// WebHook listen for events and list changed issues asynchronously
type WebHook struct {
	GithubKey string `json:"-"`
	Status    *StatusChange
}

// ServeHTTP receives the webhook, and process it
func (webhook WebHook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(webhook.GithubKey))
	if err != nil {
		glog.Error(err)
		http.Error(w, "Failed to validate payload", 400)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		glog.Error(err)
		http.Error(w, "Failed to parse event", 400)
		return

	}

	switch event := event.(type) {
	case *github.StatusEvent:
		if event.Commit != nil && event.Commit.SHA != nil {
			webhook.Status.CommitStatusChanged(*event.Commit.SHA)
		}
	}
}

// NewWebHookAndListen creates a new WebHook and listen to it
func NewWebHookAndListen(githubKey string, address string) *WebHook {
	webhook := WebHook{
		Status:    NewStatusChange(),
		GithubKey: githubKey,
	}

	go webhook.Listen(address)

	return &webhook
}

// Listen receives webhooks.
func (webhook *WebHook) Listen(address string) {
	http.Handle("/webhook", webhook)

	http.ListenAndServe(address, nil)
}

// UpdatePullRequest will add the pull-request's last commit
func (webhook *WebHook) UpdatePullRequest(id int, head string) {
	webhook.Status.UpdatePullRequestHead(id, head)
}

// PopIssues returns the list of issues that changed since last time it was called
func (webhook *WebHook) PopIssues() []int {
	return webhook.Status.PopChangedPullRequests()
}
