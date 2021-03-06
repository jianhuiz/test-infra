/*
Copyright 2017 The Kubernetes Authors.

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

// Package wip will label a PR a work-in-progress if the author provides
// a prefix to their pull request title to the same effect. The submit-
// queue will not merge pull requests with the work-in-progress label.
// The label will be removed when the title changes to no longer begin
// with the prefix.
package wip

import (
	"fmt"
	"regexp"

	"github.com/Sirupsen/logrus"

	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/plugins"
)

const (
	label      = "do-not-merge/work-in-progress"
	pluginName = "wip"
)

var (
	titleRegex = regexp.MustCompile(`(?i)^(\[WIP\]|WIP)`)
)

type event struct {
	org         string
	repo        string
	number      int
	hasLabel    bool
	needsLabel  bool
	commentBody string
}

func init() {
	plugins.RegisterPullRequestHandler(pluginName, handlePullRequest)
}

// Strict subset of *github.Client methods.
type githubClient interface {
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	AddLabel(owner, repo string, number int, label string) error
	RemoveLabel(owner, repo string, number int, label string) error
	CreateComment(org, repo string, number int, comment string) error
}

func handlePullRequest(pc plugins.PluginClient, pe github.PullRequestEvent) error {
	// These are the only actions indicating the PR title may have changed.
	if pe.Action != github.PullRequestActionOpened && pe.Action != github.PullRequestActionEdited {
		return nil
	}

	var (
		org    = pe.PullRequest.Base.Repo.Owner.Login
		repo   = pe.PullRequest.Base.Repo.Name
		number = pe.PullRequest.Number
		author = pe.PullRequest.User.Login
		title  = pe.PullRequest.Title
	)

	currentLabels, err := pc.GitHubClient.GetIssueLabels(org, repo, number)
	if err != nil {
		return fmt.Errorf("could not get labels for PR %s/%s:%d in WIP plugin: %v", org, repo, number, err)
	}
	hasLabel := false
	for _, l := range currentLabels {
		if l.Name == label {
			hasLabel = true
		}
	}

	needsLabel := titleRegex.MatchString(title)

	commentBody := plugins.FormatResponse(
		author,
		fmt.Sprintf(`Your pull request title starts with %q, so the %s label will be added.`, titleRegex.FindString(title), label),
		`This label will ensure that your pull request will not be merged. Remove the prefix from your pull request title to trigger the removal of the label and allow for your pull request to be merged.`,
	)
	e := &event{
		org:         org,
		repo:        repo,
		number:      number,
		hasLabel:    hasLabel,
		needsLabel:  needsLabel,
		commentBody: commentBody,
	}
	return handle(pc.GitHubClient, pc.Logger, e)
}

// handle interacts with GitHub to drive the pull request to the
// proper state by adding and removing comments and labels. If a
// PR has a WIP prefix, it needs an explanatory comment and label.
// Otherwise, neither should be present.
func handle(gc githubClient, le *logrus.Entry, e *event) error {
	if e.needsLabel && !e.hasLabel {
		if err := gc.AddLabel(e.org, e.repo, e.number, label); err != nil {
			le.Warnf("error while adding label %q: %v", label, err)
			return err
		}
		if err := gc.CreateComment(e.org, e.repo, e.number, e.commentBody); err != nil {
			le.Warnf("error while adding comment %q: %v", e.commentBody, err)
			return err
		}
	} else if !e.needsLabel && e.hasLabel {
		if err := gc.RemoveLabel(e.org, e.repo, e.number, label); err != nil {
			le.Warnf("error while removing label %q: %v", label, err)
			return err
		}
	}
	return nil
}
