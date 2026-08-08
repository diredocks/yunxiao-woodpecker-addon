package internal

import (
	"encoding/base64"
	"strconv"
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestConvertRepository(t *testing.T) {
	repo := &YunxiaoRepository{
		ID:          12345,
		Name:        "test-repo",
		Path:        "test-repo",
		NameWithNamespace: "test-ns / test-repo",
		PathWithNamespace: "test-ns/test-repo",
		Description:  "Test repo",
		DefaultBranch: "master",
		HTTPUrlToRepo: "https://codeup.aliyun.com/test-ns/test-repo.git",
		SSHUrlToRepo:  "git@codeup.aliyun.com:test-ns/test-repo.git",
		WebURL:        "https://codeup.aliyun.com/test-ns/test-repo",
		Namespace: struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Path        string `json:"path"`
			Description string `json:"description"`
			Avatar      string `json:"avatar"`
			Visibility  string `json:"visibility"`
			OwnerID     int    `json:"ownerId"`
			CreatedAt   string `json:"createdAt"`
			UpdatedAt   string `json:"updatedAt"`
		}{
			ID: 100, Name: "test-ns", Path: "test-ns",
		},
	}

	result := convertRepository(repo)

	if result.ForgeRemoteID != model.ForgeRemoteID("12345") {
		t.Errorf("ForgeRemoteID = %q, want %q", result.ForgeRemoteID, "12345")
	}
	if result.Name != "test-repo" {
		t.Errorf("Name = %q, want %q", result.Name, "test-repo")
	}
	if result.FullName != "test-ns / test-repo" {
		t.Errorf("FullName = %q, want %q", result.FullName, "test-ns / test-repo")
	}
	if result.Owner != "test-ns" {
		t.Errorf("Owner = %q, want %q", result.Owner, "test-ns")
	}
	if result.ForgeURL != "https://codeup.aliyun.com/test-ns/test-repo" {
		t.Errorf("ForgeURL = %q", result.ForgeURL)
	}
	if result.Clone != "https://codeup.aliyun.com/test-ns/test-repo.git" {
		t.Errorf("Clone = %q", result.Clone)
	}
	if result.CloneSSH != "git@codeup.aliyun.com:test-ns/test-repo.git" {
		t.Errorf("CloneSSH = %q", result.CloneSSH)
	}
	if result.Branch != "master" {
		t.Errorf("Branch = %q, want %q", result.Branch, "master")
	}
	if result.PREnabled != true {
		t.Error("PREnabled should be true")
	}
}

func TestConvertFileContentBase64(t *testing.T) {
	plaintext := "pipeline:\n  build:\n    image: alpine\n    commands: echo hello\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(plaintext))
	file := &YunxiaoFileContent{
		Content:  encoded,
		Encoding: "base64",
	}
	result, err := convertFileContentToBytes(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != plaintext {
		t.Errorf("content = %q, want %q", string(result), plaintext)
	}
}

func TestConvertFileContentText(t *testing.T) {
	plaintext := "plain text content"
	file := &YunxiaoFileContent{
		Content:  plaintext,
		Encoding: "text",
	}
	result, err := convertFileContentToBytes(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != plaintext {
		t.Errorf("content = %q, want %q", string(result), plaintext)
	}
}

func TestConvertFileTreeEntries(t *testing.T) {
	entries := []*YunxiaoFileTreeEntry{
		{Name: "file1.go", Type: FileTypeBlob},
		{Name: "src", Type: FileTypeTree},
	}
	result := convertFileTreeEntries(entries)
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].Name != "file1.go" {
		t.Errorf("result[0].Name = %q", result[0].Name)
	}
	if result[1].Name != "src" {
		t.Errorf("result[1].Name = %q", result[1].Name)
	}
}

func TestConvertChangeRequest(t *testing.T) {
	cr := &YunxiaoChangeRequest{
		LocalID: 42,
		Title:   "Fix critical bug",
	}
	result := convertChangeRequest(cr)
	if result.Index != model.ForgeRemoteID(strconv.Itoa(42)) {
		t.Errorf("Index = %q, want %q", result.Index, "42")
	}
	if result.Title != "Fix critical bug" {
		t.Errorf("Title = %q", result.Title)
	}
}

func TestMapCommitStatus(t *testing.T) {
	tests := []struct {
		status model.StatusValue
		want   string
	}{
		{model.StatusSuccess, CommitStateSuccess},
		{model.StatusFailure, CommitStateFailure},
		{model.StatusError, CommitStateFailure},
		{model.StatusKilled, CommitStateFailure},
		{model.StatusDeclined, CommitStateFailure},
		{model.StatusPending, CommitStatePending},
		{model.StatusRunning, CommitStatePending},
	}
	for _, tt := range tests {
		got := mapCommitStatus(tt.status)
		if got != tt.want {
			t.Errorf("mapCommitStatus(%v) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
