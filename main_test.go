package main

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"

	"yunxiao-woodpecker-addon/fixtures"
	"yunxiao-woodpecker-addon/internal"
)

func setupTest(t *testing.T) (*httptest.Server, *yunxiao) {
	t.Helper()
	mux := fixtures.NewServer()
	srv := httptest.NewServer(mux)
	f := &yunxiao{
		apiURL:         srv.URL,
		organizationID: "",
		woodpeckerHost: "https://ci.example.com",
		hookSecret:     "my-hook-secret",
	}
	return srv, f
}

func TestNewYunxiaoError(t *testing.T) {
	_, err := newYunxiao(yunxiaoOpts{APIURL: ""})
	if err == nil {
		t.Error("expected error for empty API URL")
	}
}

func TestName(t *testing.T) {
	_, f := setupTest(t)
	if f.Name() != "yunxiao" {
		t.Errorf("Name() = %q, want %q", f.Name(), "yunxiao")
	}
}

func TestURL(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()
	if f.URL() != srv.URL {
		t.Errorf("URL() = %q, want %q", f.URL(), srv.URL)
	}
}

func TestRepo(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}

	t.Run("by ForgeRemoteID", func(t *testing.T) {
		repo, err := f.Repo(context.Background(), user, "12345", "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.Name != "test-repo" {
			t.Errorf("Name = %q, want %q", repo.Name, "test-repo")
		}
		if repo.ForgeRemoteID != "12345" {
			t.Errorf("ForgeRemoteID = %q", repo.ForgeRemoteID)
		}
		if repo.Branch != "master" {
			t.Errorf("Branch = %q", repo.Branch)
		}
	})

	t.Run("by name", func(t *testing.T) {
		repo, err := f.Repo(context.Background(), user, "", "", "12345")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.Name != "test-repo" {
			t.Errorf("Name = %q", repo.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := f.Repo(context.Background(), user, "", "", "99999")
		if err == nil {
			t.Error("expected error for not found repo")
		}
	})
}

func TestRepos(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repos, err := f.Repos(context.Background(), user, &model.ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("len = %d, want 1", len(repos))
	}
	if repos[0].Name != "repo1" {
		t.Errorf("Name = %q", repos[0].Name)
	}
}

func TestBranches(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}
	branches, err := f.Branches(context.Background(), user, repo, &model.ListOptions{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("len = %d, want 2", len(branches))
	}
	if branches[0] != "master" {
		t.Errorf("branches[0] = %q, want %q", branches[0], "master")
	}
	if branches[1] != "feature-branch" {
		t.Errorf("branches[1] = %q", branches[1])
	}
}

func TestBranchHead(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}

	t.Run("existing branch", func(t *testing.T) {
		commit, err := f.BranchHead(context.Background(), user, repo, "feature-branch")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if commit.SHA != "45ede4680536406d793e0e629bc771cb9fcaa153" {
			t.Errorf("SHA = %q", commit.SHA)
		}
	})

	t.Run("fallback to commits for unknown branch", func(t *testing.T) {
		commit, err := f.BranchHead(context.Background(), user, repo, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if commit.SHA != "abc123def456" {
			t.Errorf("SHA = %q, want abc123def456", commit.SHA)
		}
	})
}

func TestFile(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}
	pipeline := &model.Pipeline{Commit: "abc123def456"}
	data, err := f.File(context.Background(), user, repo, pipeline, ".woodpecker.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "pipeline:\n  build:\n    image: alpine\n    commands: echo hello\n"
	if string(data) != expected {
		t.Errorf("data = %q", string(data))
	}
}

func TestDir(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}
	pipeline := &model.Pipeline{Commit: "abc123def456"}
	files, err := f.Dir(context.Background(), user, repo, pipeline, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("len = %d, want 2", len(files))
	}
	if files[0].Name != ".woodpecker.yml" {
		t.Errorf("files[0].Name = %q", files[0].Name)
	}
	if files[1].Name != "src" {
		t.Errorf("files[1].Name = %q", files[1].Name)
	}
}

func TestPullRequests(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}
	prs, err := f.PullRequests(context.Background(), user, repo, &model.ListOptions{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("len = %d, want 1", len(prs))
	}
	if prs[0].Title != "Fix critical bug" {
		t.Errorf("Title = %q", prs[0].Title)
	}
}

func TestStatus(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}

	t.Run("skip pending", func(t *testing.T) {
		pipeline := &model.Pipeline{Status: model.StatusPending}
		err := f.Status(context.Background(), user, repo, pipeline, nil)
		if err != nil {
			t.Errorf("expected nil error for pending status, got: %v", err)
		}
	})

	t.Run("skip running", func(t *testing.T) {
		pipeline := &model.Pipeline{Status: model.StatusRunning}
		err := f.Status(context.Background(), user, repo, pipeline, nil)
		if err != nil {
			t.Errorf("expected nil error for running status, got: %v", err)
		}
	})

	t.Run("send success", func(t *testing.T) {
		pipeline := &model.Pipeline{ID: 1, Commit: "abc123def456", Status: model.StatusSuccess}
		err := f.Status(context.Background(), user, repo, pipeline, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestActivate(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}
	err := f.Activate(context.Background(), user, repo, "https://ci.example.com/api/hook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeactivate(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	repo := &model.Repo{ForgeRemoteID: "12345"}
	err := f.Deactivate(context.Background(), user, repo, "https://ci.example.com/api/hook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHook(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	t.Run("push hook", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/hook", bytes.NewReader(fixtures.HookPushPayload))
		req.Header.Set(internal.EventTypeHeaderKey, internal.EventTypePush)
		req.Header.Set(internal.WebhookTokenHeader, "my-hook-secret")

		repo, pipeline, err := f.Hook(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pipeline.Event != model.EventPush {
			t.Errorf("Event = %v, want %v", pipeline.Event, model.EventPush)
		}
		if pipeline.Branch != "master" {
			t.Errorf("Branch = %q", pipeline.Branch)
		}
		if pipeline.Commit != "eb63d0277e64684236ebf8394b919230c4b8a286" {
			t.Errorf("Commit = %q", pipeline.Commit)
		}
		if repo.Name != "test-repo" {
			t.Errorf("repo.Name = %q", repo.Name)
		}
	})

	t.Run("tag push hook", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/hook", bytes.NewReader(fixtures.HookTagPushPayload))
		req.Header.Set(internal.EventTypeHeaderKey, internal.EventTypeTagPush)
		req.Header.Set(internal.WebhookTokenHeader, "my-hook-secret")

		repo, pipeline, err := f.Hook(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pipeline.Event != model.EventTag {
			t.Errorf("Event = %v, want %v", pipeline.Event, model.EventTag)
		}
		if pipeline.Branch != "v1.0.0" {
			t.Errorf("Branch = %q", pipeline.Branch)
		}
		if repo.Name != "test-repo" {
			t.Errorf("repo.Name = %q", repo.Name)
		}
	})

	t.Run("merge request hook", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/hook", bytes.NewReader(fixtures.HookMergeRequestPayload))
		req.Header.Set(internal.EventTypeHeaderKey, internal.EventTypeMergeRequest)
		req.Header.Set(internal.WebhookTokenHeader, "my-hook-secret")

		repo, pipeline, err := f.Hook(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pipeline.Event != model.EventPull {
			t.Errorf("Event = %v, want %v", pipeline.Event, model.EventPull)
		}
		if pipeline.Branch != "master" {
			t.Errorf("Branch = %q", pipeline.Branch)
		}
		if pipeline.Title != "Fix 2.txt" {
			t.Errorf("Title = %q", pipeline.Title)
		}
		if repo.Name != "test-repo" {
			t.Errorf("repo.Name = %q", repo.Name)
		}
		if pipeline.AdditionalVariables["change_request_local_id"] != "247" {
			t.Errorf("change_request_local_id = %q", pipeline.AdditionalVariables["change_request_local_id"])
		}
	})

	t.Run("reject invalid token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/hook", bytes.NewReader(fixtures.HookPushPayload))
		req.Header.Set(internal.EventTypeHeaderKey, internal.EventTypePush)
		req.Header.Set(internal.WebhookTokenHeader, "wrong-secret")

		_, _, err := f.Hook(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid webhook token")
		}
	})

	t.Run("ignore note hook", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/hook", bytes.NewReader([]byte(`{}`)))
		req.Header.Set(internal.EventTypeHeaderKey, internal.EventTypeNote)
		req.Header.Set(internal.WebhookTokenHeader, "my-hook-secret")

		_, _, err := f.Hook(context.Background(), req)
		if err == nil {
			t.Error("expected error for note hook (ErrIgnoreEvent)")
		}
	})
}

func TestOrg(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token", Login: "testuser"}
	org, err := f.Org(context.Background(), user, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org.Name != "testuser" {
		t.Errorf("Name = %q", org.Name)
	}
	if !org.IsUser {
		t.Error("IsUser should be true")
	}
}

func TestOrgMembership(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token", Login: "testuser", ForgeRemoteID: "99d16124"}
	perm, err := f.OrgMembership(context.Background(), user, user.Login)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !perm.Member {
		t.Error("Member should be true")
	}
	if !perm.Admin {
		t.Error("Admin should be true")
	}
}

func TestAuth(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	auth, err := f.Auth(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth != "" {
		t.Errorf("Auth() = %q, want empty", auth)
	}
}

func TestTeams(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	teams, err := f.Teams(context.Background(), nil, &model.ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if teams != nil {
		t.Error("Teams should return nil")
	}
}

func TestNetrc(t *testing.T) {
	srv, f := setupTest(t)
	defer srv.Close()

	user := &model.User{AccessToken: "test-token"}
	netrc, err := f.Netrc(user, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if netrc.Login != "test-token" {
		t.Errorf("Login = %q", netrc.Login)
	}
	if netrc.Password != "x-yunxiao-token" {
		t.Errorf("Password = %q", netrc.Password)
	}
}
