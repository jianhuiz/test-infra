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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/test-infra/prow/github"
)

const (
	upstreamCommentRe    = `^\s*/upstream\s+https://github.com/(.+)/(.+)/pull/(\d+)`
	doNotUpstreamLabel   = "do-not-upstream"
	requireUpstreamLabel = "require-upstream"
	mergedUpstreamLabel  = "merged-upstream"
)

func flagOptions() options {
	o := options{}
	flag.StringVar(&o.org, "org", "Huawei-Paas", "GitHub repository orgination")
	flag.StringVar(&o.repo, "repo", "kubernetes", "GitHub repository name")
	flag.IntVar(&o.pull, "pull", 0, "GitHub pull request number")
	flag.StringVar(&o.endpoint, "endpoint", "https://api.github.com", "GitHub's API endpoint")
	flag.StringVar(&o.token, "token", "", "Path to github token")
	flag.Parse()
	return o
}

type options struct {
	org      string
	repo     string
	pull     int
	endpoint string
	token    string
}

type pullRequest struct {
	org    string
	repo   string
	pull   int
	merged bool
}

func prKey(pr pullRequest) string {
	return fmt.Sprintf("%s/%s#%d", pr.org, pr.repo, pr.pull)
}

func hasLabel(label string, labels []github.Label) bool {
	for _, l := range labels {
		if l.Name == label {
			return true
		}
	}
	return false
}
func removeLabelIfExist(c *github.Client, org string, repo string, pull int, label string, labels []github.Label) error {
	if hasLabel(label, labels) {
		return c.RemoveLabel(org, repo, pull, label)
	}
	return nil
}
func ensureLabel(c *github.Client, org string, repo string, pull int, label string, labels []github.Label) error {
	if !hasLabel(label, labels) {
		return c.AddLabel(org, repo, pull, label)
	}
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	o := flagOptions()

	_, err := url.Parse(o.endpoint)
	if err != nil {
		log.Fatal("Must specify a valid --endpoint URL.")
	}

	if o.pull == 0 {
		log.Fatal("Must specify a valid --pull number.")
	}

	b, err := ioutil.ReadFile(o.token)
	if err != nil {
		log.Fatalf("cannot read --token: %v", err)
	}
	tok := strings.TrimSpace(string(b))
	c := github.NewClient(tok, o.endpoint)

	_, err = c.GetPullRequest(o.org, o.repo, o.pull)
	if err != nil {
		log.Fatalf("Failed getting PR %s/%s#%d, err: %v", o.org, o.repo, o.pull, err)
	}

	labels, err := c.GetIssueLabels(o.org, o.repo, o.pull)
	for _, l := range labels {
		if l.Name == doNotUpstreamLabel {
			log.Printf("PR has %v label, upstream checking succeeded", doNotUpstreamLabel)
			removeLabelIfExist(c, o.org, o.repo, o.pull, requireUpstreamLabel, labels)
			removeLabelIfExist(c, o.org, o.repo, o.pull, mergedUpstreamLabel, labels)
			return
		}
	}

	comments, err := c.ListIssueComments(o.org, o.repo, o.pull)
	if err != nil {
		log.Fatalf("Failed getting PR comments %v/%v#%v, err: %v", o.org, o.repo, o.pull, err)
	}

	upstreamPrs := map[string]*pullRequest{}
	re := regexp.MustCompile(upstreamCommentRe)
	for i := range comments {
		matches := re.FindAllStringSubmatch(comments[i].Body, -1)
		for _, m := range matches {
			pull, _ := strconv.Atoi(m[3])
			pr := pullRequest{m[1], m[2], pull, false}
			upstreamPrs[prKey(pr)] = &pr
		}
	}

	if len(upstreamPrs) == 0 {
		ensureLabel(c, o.org, o.repo, o.pull, requireUpstreamLabel, labels)
		removeLabelIfExist(c, o.org, o.repo, o.pull, mergedUpstreamLabel, labels)
		log.Fatalf("FAIL: No upstream PR created. Please create the upstream PR(s) and reference it(them) by add a comment contains `/upstream <full-upstream-pull-request-url>` and re-run the test. Reference to each upstream PR should be on their own lines")
	}

	for k, pr := range upstreamPrs {
		prInfo, err := c.GetPullRequest(pr.org, pr.repo, pr.pull)
		if err != nil {
			log.Printf("Error getting status of upstream PR %s/%s#%d", pr.org, pr.repo, pr.pull)
		}
		upstreamPrs[k].merged = prInfo.Merged
	}
	// at least one upstream PR not merged
	for _, pr := range upstreamPrs {
		if !pr.merged {
			ensureLabel(c, o.org, o.repo, o.pull, requireUpstreamLabel, labels)
			removeLabelIfExist(c, o.org, o.repo, o.pull, mergedUpstreamLabel, labels)
			log.Printf("Upstream PR %s/%s#%d not merged", pr.org, pr.repo, pr.pull)
			return
		}
	}
	// all upstream PRs merged
	ensureLabel(c, o.org, o.repo, o.pull, mergedUpstreamLabel, labels)
	removeLabelIfExist(c, o.org, o.repo, o.pull, requireUpstreamLabel, labels)
}
