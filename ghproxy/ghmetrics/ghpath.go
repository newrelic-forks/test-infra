/*
Copyright 2019 The Kubernetes Authors.

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

package ghmetrics

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetSimplifiedPath returns a variable-free path that can be used as label for prometheus metrics
func GetSimplifiedPath(path string) string {
	tree := l("", // shadow element mimicing the root
		l("repos",
			v("owner",
				v("repo",
					l("branches", v("branch", l("protection",
						l("restrictions", l("users"), l("teams")),
						l("required_status_checks", l("contexts")),
						l("required_pull_request_reviews"),
						l("required_signatures"),
						l("enforce_admins")))),
					l("issues",
						l("comments", v("commentId")),
						l("events", v("eventId")),
						v("issueId",
							l("lock"),
							l("comments"),
							l("events"),
							l("labels", v("labelId")))),
					l("keys", v("keyId")),
					l("labels", v("labelId")),
					l("milestones", v("milestone")),
					l("pulls", v("pullId")),
					l("releases", v("releaseId")),
					l("statuses", v("statusId")),
					l("subscribers", v("subscriberId")),
					l("assignees", v("assigneeId")),
					l("archive", v("zip")),
					l("collaborators", v("collaboratorId")),
					l("comments", v("commentId")),
					l("compare", v("sha")),
					l("contents", v("contentId")),
					l("commits", v("sha")),
					l("git",
						l("commits", v("sha")),
						l("ref", v("refId")),
						l("tags", v("tagId")),
						l("trees", v("sha")),
						l("refs", l("heads", v("ref")))),
					l("stars"),
					l("merges"),
					l("stargazers"),
					l("notifications"),
					l("hooks"),
					l("deployments"),
					l("downloads"),
					l("events"),
					l("forks"),
					l("topics"),
					l("vulnerability-alerts"),
					l("automated-security-fixes"),
					l("contributors"),
					l("languages"),
					l("teams"),
					l("tags"),
					l("transfer")))),
		l("user",
			l("following", v("userId")),
			l("keys", v("keyId")),
			l("email", l("visibility")),
			l("emails"),
			l("public_emails"),
			l("followers"),
			l("starred"),
			l("issues")),
		l("users",
			v("username",
				l("followers", v("username")),
				l("repos"),
				l("hovercard"),
				l("following"))),
		l("orgs",
			v("orgname",
				l("credential-authorizations", v("credentialId")),
				l("repos"),
				l("issues"),
				l("invitations"),
				l("members"),
				l("teams"))),
		l("organizations",
			v("orgId",
				l("members"),
				l("teams"))),
		l("issues", v("issueId")),
		l("search",
			l("repositories"),
			l("commits"),
			l("code"),
			l("issues"),
			l("users"),
			l("topics"),
			l("labels")),
		l("gists",
			l("public"),
			l("starred")),
		l("notifications", l("threads", v("threadId", l("subscription")))),
		l("repositories"),
		l("emojis"),
		l("events"),
		l("feeds"),
		l("hub"),
		l("rate_limit"),
		l("teams"),
		l("licenses"))

	splitPath := strings.FieldsFunc(path, splitFunc)
	resolvedPath, matches := resolve(tree, splitPath)
	if !matches {
		logrus.WithField("path", path).Warning("Path not handled. This is a bug in GHProxy, please open an issue against the kubernetes/test-infra repository with this error message.")
		return path
	}
	return resolvedPath
}

func splitFunc(c rune) bool {
	return c == '/'
}

type node struct {
	PathFragment
	children []node
}

// PathFragment Interface for tree leafs to help resolve paths
type PathFragment interface {
	Matches(part string) bool
	Represent() string
}

type literal string

func (l literal) Matches(part string) bool {
	return string(l) == part
}

func (l literal) Represent() string {
	return string(l)
}

type variable string

func (v variable) Matches(part string) bool {
	return true
}

func (v variable) Represent() string {
	return ":" + string(v)
}

func l(fragment string, children ...node) node {
	return node{
		PathFragment: literal(fragment),
		children:     children,
	}
}

func v(fragment string, children ...node) node {
	return node{
		PathFragment: variable(fragment),
		children:     children,
	}
}

func resolve(parent node, path []string) (string, bool) {
	if len(path) == 0 {
		return "", true
	}
	if parent.Represent() == "" {
		for _, child := range parent.children {
			suffix, matched := resolve(child, path)
			if matched {
				if suffix != "" {
					return fmt.Sprintf("/%s", suffix), true
				}
			}
		}
	}
	if !parent.Matches(path[0]) {
		return "", false
	}
	representation := parent.Represent()
	if len(parent.children) == 0 {
		return representation, true
	}

	for _, child := range parent.children {
		suffix, matched := resolve(child, path[1:])
		if matched {
			if suffix != "" {
				return fmt.Sprintf("%s/%s", representation, suffix), true
			}
			return representation, true
		}
	}
	return "", false
}
