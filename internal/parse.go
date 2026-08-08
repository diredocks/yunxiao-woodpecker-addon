package internal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func parsePushHook(payload []byte) (*model.Repo, *model.Pipeline, error) {
	var hook HookPushPayload
	if err := json.Unmarshal(payload, &hook); err != nil {
		return nil, nil, err
	}

	repo := convertHookRepository(&hook.Repository, hook.ProjectID)

	ref := hook.Ref
	branch := strings.TrimPrefix(ref, "refs/heads/")

	var changedFiles []string
	message := ""
	author := hook.UserName
	for i, c := range hook.Commits {
		if i == 0 {
			message = c.Message
			if c.Author.Name != "" {
				author = c.Author.Name
			}
		}
	}

	pipeline := &model.Pipeline{
		Event:        model.EventPush,
		Branch:       branch,
		Commit:       hook.After,
		Ref:          ref,
		Message:      message,
		Author:       author,
		ChangedFiles: changedFiles,
	}

	return repo, pipeline, nil
}

func parseTagPushHook(payload []byte) (*model.Repo, *model.Pipeline, error) {
	var hook HookTagPushPayload
	if err := json.Unmarshal(payload, &hook); err != nil {
		return nil, nil, err
	}

	repo := convertHookRepository(&hook.Repository, hook.ProjectID)

	ref := hook.Ref
	tag := strings.TrimPrefix(ref, "refs/tags/")

	pipeline := &model.Pipeline{
		Event:  model.EventTag,
		Branch: tag,
		Commit: hook.After,
		Ref:    ref,
		Author: hook.UserName,
	}

	return repo, pipeline, nil
}

func parseMergeRequestHook(payload []byte) (*model.Repo, *model.Pipeline, error) {
	var hook HookMergeRequestPayload
	if err := json.Unmarshal(payload, &hook); err != nil {
		return nil, nil, err
	}

	repo := convertHookRepository(&hook.Repository, hook.ObjectAttributes.ProjectID)

	action := hook.ObjectAttributes.Action
	isOpen := action == "open" || action == "reopen"
	isClose := action == "close" || action == "merge"

	pipeline := &model.Pipeline{
		Event:  model.EventPull,
		Branch: hook.ObjectAttributes.TargetBranch,
		Commit: hook.ObjectAttributes.LastCommit.ID,
		Ref:    fmt.Sprintf("refs/merge-requests/%d/head", hook.ObjectAttributes.LocalID),
		Refspec: fmt.Sprintf("%s:%s",
			hook.ObjectAttributes.SourceBranch,
			hook.ObjectAttributes.TargetBranch),
		Message: hook.ObjectAttributes.Title,
		Author:  hook.ObjectAttributes.LastCommit.Author.Name,
		Title:   hook.ObjectAttributes.Title,
		AdditionalVariables: map[string]string{
			"change_request_local_id": strconv.Itoa(hook.ObjectAttributes.LocalID),
			"change_request_source":   hook.ObjectAttributes.SourceBranch,
			"change_request_target":   hook.ObjectAttributes.TargetBranch,
			"change_request_project":  strconv.Itoa(hook.ObjectAttributes.ProjectID),
		},
	}

	_ = action
	_ = isOpen
	_ = isClose

	return repo, pipeline, nil
}

func convertHookRepository(repo *HookRepository, projectID int) *model.Repo {
	return &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(strconv.Itoa(projectID)),
		Name:          repo.Name,
		FullName:      repo.Name,
		Owner:         extractOwnerFromURL(repo.Homepage),
		ForgeURL:      repo.Homepage,
		Clone:         repo.GitHTTPURL,
		CloneSSH:      repo.GitSSHURL,
		PREnabled:     true,
	}
}

func extractOwnerFromURL(homepage string) string {
	parts := strings.Split(strings.TrimRight(homepage, "/"), "/")
	if len(parts) >= 3 {
		return parts[len(parts)-2]
	}
	return ""
}
