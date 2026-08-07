package main

import (
	"encoding/base64"
	"strconv"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"

	"yunxiao-woodpecker-addon/internal"
)

// YunxiaoImage is a base64-encoded PNG placeholder for users without avatar.
const YunxiaoImage = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func convertRepository(repo *internal.YunxiaoRepository) *model.Repo {
	forgeRemoteID := model.ForgeRemoteID(strconv.Itoa(repo.ID))
	owner := repo.Namespace.Path
	if owner == "" {
		owner = repo.Owner.Username
	}
	return &model.Repo{
		ForgeRemoteID: forgeRemoteID,
		Name:          repo.Name,
		FullName:      repo.NameWithNamespace,
		Owner:         owner,
		ForgeURL:      repo.WebURL,
		Clone:         repo.HTTPUrlToRepo,
		CloneSSH:      repo.SSHUrlToRepo,
		Branch:        repo.DefaultBranch,
		PREnabled:     true,
	}
}

func convertBranchToModel(branch *internal.YunxiaoBranch) string {
	return branch.Name
}

func convertCommitToModel(commit *internal.YunxiaoCommit) *model.Commit {
	return &model.Commit{
		SHA:      commit.ID,
		ForgeURL: commit.WebURL,
	}
}

func convertChangeRequest(cr *internal.YunxiaoChangeRequest) *model.PullRequest {
	return &model.PullRequest{
		Index: model.ForgeRemoteID(strconv.Itoa(cr.LocalID)),
		Title: cr.Title,
	}
}

func convertFileContentToBytes(file *internal.YunxiaoFileContent) ([]byte, error) {
	if file.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(file.Content)
	}
	return []byte(file.Content), nil
}

func convertFileTreeEntries(entries []*internal.YunxiaoFileTreeEntry) []*types.FileMeta {
	var result []*types.FileMeta
	for _, entry := range entries {
		result = append(result, &types.FileMeta{
			Name: entry.Name,
		})
	}
	return result
}

func mapCommitStatus(status model.StatusValue) string {
	switch status {
	case model.StatusFailure, model.StatusError, model.StatusKilled, model.StatusDeclined, model.StatusBlocked:
		return internal.CommitStateFailure
	case model.StatusSuccess:
		return internal.CommitStateSuccess
	case model.StatusPending, model.StatusRunning:
		return internal.CommitStatePending
	default:
		return internal.CommitStateError
	}
}
