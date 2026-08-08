// Test fixtures providing a mock yunxiao API server for integration testing.
package fixtures

import (
	"net/http"
	"strconv"
	"strings"
)

var (
	UserPayload = []byte(`{
		"createdAt": "2023-03-22T12:44:50.048Z",
		"email": "test@example.com",
		"id": "99d16124",
		"name": "Test User",
		"nickName": "Tester",
		"username": "testuser"
	}`)

	OrganizationsPayload = []byte(`[{
		"createdAt": "2023-08-31T03:59:16.201Z",
		"creatorId": "99d16124",
		"defaultRole": "member",
		"description": "Test Org",
		"id": "org123",
		"name": "test-org",
		"updateAt": "2023-08-31T03:59:16.201Z"
	}]`)

	RepositoryPayload = []byte(`{
		"accessLevel": 40,
		"defaultBranch": "master",
		"description": "Test repository",
		"httpUrlToRepo": "https://codeup.aliyun.com/test-ns/test-repo.git",
		"sshUrlToRepo": "git@codeup.aliyun.com:test-ns/test-repo.git",
		"id": 12345,
		"name": "test-repo",
		"nameWithNamespace": "test-ns / test-repo",
		"path": "test-repo",
		"pathWithNamespace": "test-ns/test-repo",
		"namespace": {
			"id": 100,
			"name": "test-ns",
			"path": "test-ns"
		},
		"owner": {
			"id": 1,
			"userId": "99d16124",
			"username": "testuser",
			"name": "Test User"
		},
		"visibility": "private",
		"webUrl": "https://codeup.aliyun.com/test-ns/test-repo",
		"archived": false
	}`)

	RepositoryPayloadNoRoute = []byte(`{"message": "Repository not found"}`)

	RepositoriesPayloadPage0 = []byte(`[{
		"accessLevel": 40,
		"defaultBranch": "master",
		"httpUrlToRepo": "https://codeup.aliyun.com/test-ns/repo1.git",
		"sshUrlToRepo": "git@codeup.aliyun.com:test-ns/repo1.git",
		"id": 12345,
		"name": "repo1",
		"nameWithNamespace": "test-ns / repo1",
		"path": "repo1",
		"pathWithNamespace": "test-ns/repo1",
		"namespace": {"id": 100, "name": "test-ns", "path": "test-ns"},
		"owner": {"id": 1, "userId": "99d16124", "username": "testuser", "name": "Test User"},
		"visibility": "private",
		"webUrl": "https://codeup.aliyun.com/test-ns/repo1",
		"archived": false
	}]`)

	BranchPayload = []byte(`{
		"commit": {
			"id": "45ede4680536406d793e0e629bc771cb9fcaa153",
			"shortId": "45ede468",
			"title": "commit title",
			"message": "commit message",
			"authorName": "Test User",
			"authorEmail": "test@example.com",
			"authoredDate": "2024-04-05T15:30:45Z",
			"committedDate": "2024-04-05T15:30:45Z",
			"committerName": "Test User",
			"committerEmail": "test@example.com",
			"parentIds": ["3fdaf119cf76539c1a47de0074ac02927ef4c8e1"]
		},
		"defaultBranch": false,
		"name": "feature-branch",
		"protected": false
	}`)

	BranchesPayload = []byte(`[{
		"commit": {
			"id": "abc123def456",
			"shortId": "abc123de",
			"title": "feat: add feature",
			"message": "feat: add feature",
			"authorName": "Test User",
			"authorEmail": "test@example.com",
			"authoredDate": "2024-04-05T15:30:45Z",
			"committedDate": "2024-04-05T15:30:45Z",
			"committerName": "Test User",
			"committerEmail": "test@example.com",
			"parentIds": []
		},
		"defaultBranch": true,
		"name": "master",
		"protected": true
	}, {
		"commit": {
			"id": "45ede4680536406d793e0e629bc771cb9fcaa153",
			"shortId": "45ede468",
			"title": "commit title",
			"message": "commit message",
			"authorName": "Test User",
			"authorEmail": "test@example.com",
			"authoredDate": "2024-04-05T15:30:45Z",
			"committedDate": "2024-04-05T15:30:45Z",
			"committerName": "Test User",
			"committerEmail": "test@example.com",
			"parentIds": []
		},
		"defaultBranch": false,
		"name": "feature-branch",
		"protected": false
	}]`)

	CommitsPayload = []byte(`[{
		"id": "abc123def456",
		"shortId": "abc123de",
		"title": "feat: add feature",
		"message": "feat: add feature\n\nDetailed description",
		"authorName": "Test User",
		"authorEmail": "test@example.com",
		"authoredDate": "2024-04-05T15:30:45Z",
		"committedDate": "2024-04-05T15:30:45Z",
		"committerName": "Test User",
		"committerEmail": "test@example.com",
		"parentIds": [],
		"webUrl": "https://codeup.aliyun.com/test-ns/test-repo/commits/abc123def456"
	}]`)

	FileContentPayload = []byte(`{
		"blobId": "b573bb50d56e8c19282593cbf5b081e211923a83",
		"commitId": "abc123def456",
		"content": "cGlwZWxpbmU6CiAgYnVpbGQ6CiAgICBpbWFnZTogYWxwaW5lCiAgICBjb21tYW5kczogZWNobyBoZWxsbwo=",
		"encoding": "base64",
		"fileName": ".woodpecker.yml",
		"filePath": ".woodpecker.yml",
		"lastCommitId": "abc123def456",
		"ref": "master",
		"size": 64
	}`)

	FileTreePayload = []byte(`[{
		"id": "b573bb50d56e8c19282593cbf5b081e211923a83",
		"isLFS": false,
		"mode": "100644",
		"name": ".woodpecker.yml",
		"path": ".woodpecker.yml",
		"type": "blob"
	}, {
		"id": "c684ec50e67f9d29292593dbf6c182f312034b94",
		"isLFS": false,
		"mode": "040000",
		"name": "src",
		"path": "src",
		"type": "tree"
	}]`)

	ChangeRequestsPayload = []byte(`[{
		"author": {
			"userId": "99d16124",
			"username": "testuser",
			"name": "Test User",
			"email": "test@example.com",
			"state": "active",
			"avatar": ""
		},
		"createdAt": "2024-10-05T15:30:45Z",
		"creationMethod": "WEB",
		"description": "Fix bug",
		"localId": 1,
		"projectId": 12345,
		"sourceBranch": "feature-branch",
		"sourceProjectId": 12345,
		"sourceType": "BRANCH",
		"state": "opened",
		"targetBranch": "master",
		"targetProjectId": 12345,
		"title": "Fix critical bug",
		"totalCommentCount": 0,
		"hasConflict": false,
		"workInProgress": false,
		"updatedAt": "2024-10-05T15:30:45Z",
		"webUrl": "https://codeup.aliyun.com/test-ns/test-repo/change/1"
	}]`)

	CreateWebhookPayload = []byte(`{
		"createdAt": "2024-10-05T15:30:45Z",
		"description": "Woodpecker CI webhook",
		"enableSSLVerification": false,
		"id": 1,
		"mergeRequestEvents": true,
		"noteEvents": false,
		"pushEvents": true,
		"repositoryId": 12345,
		"tagPushEvents": true,
		"token": "my-hook-secret",
		"url": "https://ci.example.com/api/hook"
	}`)

	ListWebhooksPayload = []byte(`[{
		"id": 1,
		"url": "https://ci.example.com/api/hook",
		"token": "my-hook-secret",
		"pushEvents": true,
		"tagPushEvents": true,
		"noteEvents": false,
		"mergeRequestEvents": true,
		"description": "Woodpecker CI webhook",
		"repositoryId": 12345,
		"enableSSLVerification": false
	}]`)

	DeleteWebhookPayload = []byte(`{"result": true}`)

	CommitStatusPayload = []byte(`{
		"id": 1,
		"sha": "abc123def456",
		"context": "woodpecker-ci",
		"state": "success",
		"description": "CI pipeline passed",
		"createdAt": "2024-10-05T15:30:45Z",
		"updatedAt": "2024-10-05T15:30:45Z",
		"author": {
			"id": "99d16124",
			"username": "testuser",
			"name": "Test User",
			"type": "User"
		}
	}`)

	EmptyArrayPayload = []byte(`[]`)

	MembersPayload = []byte(`[{
		"deptIds": ["dept1"],
		"email": "test@example.com",
		"id": "member1",
		"joined": "2023-08-31T03:59:16.201Z",
		"lastUpdated": "2023-08-31T03:59:16.201Z",
		"name": "Test User",
		"organizationId": "org123",
		"roleIds": ["role1"],
		"status": "ENABLED",
		"userId": "99d16124",
		"visited": "2023-08-31T03:59:16.201Z"
	}]`)

	HookPushPayload = []byte(`{
		"object_kind": "push",
		"before": "f2e2d577fab1562a6239b82721fd9827e05fdce6",
		"after": "eb63d0277e64684236ebf8394b919230c4b8a286",
		"ref": "refs/heads/master",
		"user_id": 4,
		"user_name": "Test User",
		"user_email": "test@example.com",
		"project_id": 15,
		"repository": {
			"name": "test-repo",
			"url": "git@codeup.aliyun.com:demo/test-repo.git",
			"description": "",
			"homepage": "https://codeup.aliyun.com/demo/test-repo",
			"git_http_url": "https://codeup.aliyun.com/demo/test-repo.git",
			"git_ssh_url": "git@codeup.aliyun.com:demo/test-repo.git",
			"visibility_level": 0
		},
		"commits": [{
			"id": "eb63d0277e64684236ebf8394b919230c4b8a286",
			"message": "Fixed readme",
			"timestamp": "2019-01-03T23:36:29+08:00",
			"url": "https://codeup.aliyun.com/demo/test-repo/commits/eb63d027",
			"author": {
				"name": "Test User",
				"email": "test@example.com"
			}
		}],
		"total_commits_count": 1
	}`)

	HookTagPushPayload = []byte(`{
		"object_kind": "tag_push",
		"ref": "refs/tags/v1.0.0",
		"before": "0000000000000000000000000000000000000000",
		"after": "eb63d0277e64684236ebf8394b919230c4b8a286",
		"user_id": 1,
		"user_name": "Test User",
		"project_id": 15,
		"repository": {
			"name": "test-repo",
			"url": "ssh://git@codeup.aliyun.com/demo/test-repo.git",
			"description": "",
			"homepage": "https://codeup.aliyun.com/demo/test-repo",
			"git_http_url": "https://codeup.aliyun.com/demo/test-repo.git",
			"git_ssh_url": "git@codeup.aliyun.com:demo/test-repo.git",
			"visibility_level": 0
		},
		"commits": [],
		"total_commits_count": 0
	}`)

	HookMergeRequestPayload = []byte(`{
		"object_kind": "merge_request",
		"user": {
			"name": "Test User",
			"username": "testuser",
			"avatar_url": ""
		},
		"repository": {
			"name": "test-repo",
			"url": "https://codeup.aliyun.com/demo/test-repo.git",
			"homepage": "https://codeup.aliyun.com/demo/test-repo",
			"git_http_url": "https://codeup.aliyun.com/demo/test-repo.git",
			"git_ssh_url": "git@codeup.aliyun.com:demo/test-repo.git"
		},
		"object_attributes": {
			"action": "open",
			"last_commit": {
				"author": {
					"name": "Test User",
					"email": "test@example.com"
				},
				"id": "b4f8b799e9698278efb7d23bfae4fff17252eba3",
				"message": "Fix 2.txt",
				"timestamp": "2023-06-13T10:09:53+08:00",
				"url": "https://codeup.aliyun.com/demo/test-repo/commits/b4f8b799"
			},
			"local_id": 247,
			"merge_status": "unchecked",
			"project_id": 15,
			"source": {
				"http_url": "https://codeup.aliyun.com/demo/test-repo.git",
				"name": "test-repo",
				"namespace": "demo",
				"ssh_url": "git@codeup.aliyun.com:demo/test-repo.git",
				"visibility_level": 0,
				"web_url": "https://codeup.aliyun.com/demo/test-repo"
			},
			"source_branch": "feature-branch",
			"source_project_id": 15,
			"source_type": "BRANCH",
			"state": "opened",
			"target": {
				"http_url": "https://codeup.aliyun.com/demo/test-repo.git",
				"name": "test-repo",
				"namespace": "demo",
				"ssh_url": "git@codeup.aliyun.com:demo/test-repo.git",
				"visibility_level": 0,
				"web_url": "https://codeup.aliyun.com/demo/test-repo"
			},
			"target_branch": "master",
			"target_project_id": 15,
			"title": "Fix 2.txt",
			"updated_at": "2023-07-10T14:25:30+08:00",
			"url": "https://codeup.aliyun.com/demo/test-repo/change/247",
			"work_in_progress": false
		}
	}`)

	NotFoundPayload = []byte(`{"message": "not found"}`)
)

func jsonResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func NewServer() *http.ServeMux {
	mux := http.NewServeMux()

	// Platform APIs
	mux.HandleFunc("GET /oapi/v1/platform/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-yunxiao-token") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			jsonResponse(w, []byte(`{"message":"missing token"}`))
			return
		}
		jsonResponse(w, UserPayload)
	})
	mux.HandleFunc("GET /oapi/v1/platform/organizations", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, OrganizationsPayload)
	})
	mux.HandleFunc("GET /oapi/v1/platform/organizations/{orgId}/members", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, MembersPayload)
	})
	mux.HandleFunc("GET /oapi/v1/platform/members", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, MembersPayload)
	})

	// Codeup APIs - region edition
	mux.HandleFunc("GET /oapi/v1/codeup/repositories", func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page > 1 {
			jsonResponse(w, EmptyArrayPayload)
			return
		}
		w.Header().Set("x-total", "1")
		w.Header().Set("x-total-pages", "1")
		w.Header().Set("x-page", "1")
		w.Header().Set("x-per-page", "20")
		jsonResponse(w, RepositoriesPayloadPage0)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/repositories/{repoId}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("repoId") == "99999" {
			w.WriteHeader(http.StatusNotFound)
			jsonResponse(w, NotFoundPayload)
			return
		}
		jsonResponse(w, RepositoryPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/repositories/{repoId}/branches", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, BranchesPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/repositories/{repoId}/branches/{branchName}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("branchName") == "nonexistent" {
			w.WriteHeader(http.StatusNotFound)
			jsonResponse(w, NotFoundPayload)
			return
		}
		jsonResponse(w, BranchPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/repositories/{repoId}/commits", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, CommitsPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/repositories/{repoId}/files/tree", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, FileTreePayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/repositories/{repoId}/files/{filePath...}", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.PathValue("filePath")
		if filePath == "nonexistent" {
			w.WriteHeader(http.StatusNotFound)
			jsonResponse(w, NotFoundPayload)
			return
		}
		jsonResponse(w, FileContentPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/changeRequests", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, ChangeRequestsPayload)
	})
	mux.HandleFunc("POST /oapi/v1/codeup/repositories/{repoId}/commits/{sha}/statuses", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, CommitStatusPayload)
	})
	mux.HandleFunc("POST /oapi/v1/codeup/repositories/{repoId}/webhooks", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, CreateWebhookPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/repositories/{repoId}/webhooks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-total", "1")
		w.Header().Set("x-total-pages", "1")
		jsonResponse(w, ListWebhooksPayload)
	})
	mux.HandleFunc("DELETE /oapi/v1/codeup/repositories/{repoId}/webhooks/{hookId}", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, DeleteWebhookPayload)
	})

	// Codeup APIs - center edition
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories", func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page > 1 {
			jsonResponse(w, EmptyArrayPayload)
			return
		}
		w.Header().Set("x-total", "1")
		w.Header().Set("x-total-pages", "1")
		w.Header().Set("x-page", "1")
		w.Header().Set("x-per-page", "20")
		jsonResponse(w, RepositoriesPayloadPage0)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("repoId") == "99999" {
			w.WriteHeader(http.StatusNotFound)
			jsonResponse(w, NotFoundPayload)
			return
		}
		jsonResponse(w, RepositoryPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/branches", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, BranchesPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/branches/{branchName}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("branchName") == "nonexistent" {
			w.WriteHeader(http.StatusNotFound)
			jsonResponse(w, NotFoundPayload)
			return
		}
		jsonResponse(w, BranchPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/commits", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, CommitsPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/files/tree", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, FileTreePayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/files/{filePath...}", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.PathValue("filePath")
		if filePath == "nonexistent" {
			w.WriteHeader(http.StatusNotFound)
			jsonResponse(w, NotFoundPayload)
			return
		}
		jsonResponse(w, FileContentPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/changeRequests", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, ChangeRequestsPayload)
	})
	mux.HandleFunc("POST /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/commits/{sha}/statuses", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, CommitStatusPayload)
	})
	mux.HandleFunc("POST /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/webhooks", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, CreateWebhookPayload)
	})
	mux.HandleFunc("GET /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/webhooks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-total", "1")
		w.Header().Set("x-total-pages", "1")
		jsonResponse(w, ListWebhooksPayload)
	})
	mux.HandleFunc("DELETE /oapi/v1/codeup/organizations/{orgId}/repositories/{repoId}/webhooks/{hookId}", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, DeleteWebhookPayload)
	})

	// Catch-all: some URLs have path-encoded segments that Go routing might not match
	// Fallback handler for paths containing encoded repository IDs
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// Handle files with full encoded paths (e.g. /oapi/v1/codeup/repositories/{id}/files/src%2Fmain.go)
		if strings.Contains(path, "/repositories/") && strings.Contains(path, "/files/") {
			jsonResponse(w, FileContentPayload)
			return
		}
		// Handle change requests with organization prefix
		if strings.Contains(path, "/changeRequests") && method == "GET" {
			jsonResponse(w, ChangeRequestsPayload)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		jsonResponse(w, NotFoundPayload)
	})

	return mux
}
