package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/addon"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/common"
	forgeTypes "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"k8s.io/utils/env"

	"yunxiao-woodpecker-addon/internal"
	"yunxiao-woodpecker-addon/pkg/version"
)

var _ forge.Forge = (*yunxiao)(nil)

type yunxiao struct {
	apiURL         string
	organizationID string
	woodpeckerHost string
	hookSecret     string
}

type yunxiaoOpts struct {
	APIURL         string
	OrganizationID string
	WoodpeckerHost string
	HookSecret     string
}

func main() {
	logLevel := env.GetString("LOG_LEVEL", "info")
	var slogLevel slog.Level
	switch logLevel {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(slogLevel)
	slog.Info("yunxiao-woodpecker-addon is starting",
		"version", version.GetVersion(),
		"revision", version.GetRevision(),
		"build_time", version.GetBuildTime())

	opts := yunxiaoOpts{
		APIURL:         env.GetString("YUNXIAO_API_URL", ""),
		OrganizationID: env.GetString("YUNXIAO_ORGANIZATION_ID", ""),
		WoodpeckerHost: env.GetString("WOODPECKER_HOST", ""),
		HookSecret:     env.GetString("YUNXIAO_HOOK_SECRET", ""),
	}

	f, err := newYunxiao(opts)
	if err != nil {
		slog.Error("failed to create yunxiao forge", "error", err)
		return
	}
	addon.Serve(f)
}

func newYunxiao(opts yunxiaoOpts) (*yunxiao, error) {
	f := &yunxiao{
		apiURL:         strings.TrimSuffix(opts.APIURL, "/"),
		organizationID: opts.OrganizationID,
		woodpeckerHost: strings.TrimSuffix(opts.WoodpeckerHost, "/"),
		hookSecret:     opts.HookSecret,
	}

	if f.apiURL == "" {
		return nil, errors.New("must provide a value for YUNXIAO_API_URL")
	}
	if _, err := url.Parse(f.apiURL); err != nil {
		return nil, fmt.Errorf("must provide a valid YUNXIAO_API_URL value: %w", err)
	}

	return f, nil
}

func (f *yunxiao) newClient(ctx context.Context, u *model.User) *internal.Client {
	token := ""
	if u != nil {
		token = u.AccessToken
	}
	return internal.NewClient(ctx, f.apiURL, token, f.organizationID)
}

// Name returns the string name of this driver.
func (f *yunxiao) Name() string {
	return "yunxiao"
}

// URL returns the root url of the configured forge.
func (f *yunxiao) URL() string {
	return f.apiURL
}

// Login authenticates the session and returns the forge user details.
// Yunxiao uses personal access tokens. Phase 1 returns a redirect URL
// where users can input their token. Phase 2 validates the token.
func (f *yunxiao) Login(ctx context.Context, req *forgeTypes.OAuthRequest) (*model.User, string, error) {
	slog.Debug("Called Login")

	loginURL := fmt.Sprintf("%s/authorize", f.woodpeckerHost)

	if req == nil || req.Code == "" {
		return nil, loginURL, nil
	}

	client := internal.NewClient(ctx, f.apiURL, req.Code, f.organizationID)
	userInfo, err := client.GetCurrentUser()
	if err != nil {
		return nil, loginURL, fmt.Errorf("invalid token: %w", err)
	}

	avatarData := YunxiaoImage

	slog.Info("Logged in", "user", userInfo.Username, "id", userInfo.ID)

	return &model.User{
		ForgeRemoteID: model.ForgeRemoteID(userInfo.ID),
		Login:         userInfo.Username,
		Avatar:        avatarData,
		AccessToken:   req.Code,
	}, loginURL, nil
}

// Auth authenticates the session and returns the forge user login for the given token and secret.
// Not used by Yunxiao since there is no OAuth process.
func (f *yunxiao) Auth(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

// Teams fetches a list of team memberships from the forge.
func (f *yunxiao) Teams(_ context.Context, _ *model.User, _ *model.ListOptions) ([]*model.Team, error) {
	return nil, nil
}

// Repo fetches the named repository from the forge.
func (f *yunxiao) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	slog.Debug("Called Repo", "remoteID", remoteID, "owner", owner, "name", name)

	repoID := name
	if remoteID.IsValid() {
		repoID = string(remoteID)
	}

	client := f.newClient(ctx, u)
	repo, err := client.GetRepository(repoID)
	if err != nil {
		slog.Error("failed to get repository", "error", err)
		return nil, err
	}
	return convertRepository(repo), nil
}

// Repos fetches a list of repos from the forge.
func (f *yunxiao) Repos(ctx context.Context, u *model.User, _ *model.ListOptions) ([]*model.Repo, error) {
	slog.Debug("Called Repos")

	client := f.newClient(ctx, u)
	repos, err := client.ListRepositories(1, 100)
	if err != nil {
		slog.Error("failed to list repositories", "error", err)
		return nil, err
	}

	var result []*model.Repo
	for _, repo := range repos {
		result = append(result, convertRepository(repo))
	}
	return result, nil
}

// File fetches a file from the forge repository and returns it in bytes.
func (f *yunxiao) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, file string) ([]byte, error) {
	slog.Debug("Called File", "repo", r.ForgeRemoteID, "commit", b.Commit, "file", file)

	client := f.newClient(ctx, u)
	fileContent, err := client.GetFileContent(string(r.ForgeRemoteID), file, b.Commit)
	if err != nil {
		slog.Error("failed to get file", "error", err)
		return nil, err
	}
	return convertFileContentToBytes(fileContent)
}

// Dir fetches a folder from the forge repository.
func (f *yunxiao) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, dir string) ([]*forgeTypes.FileMeta, error) {
	slog.Debug("Called Dir", "repo", r.ForgeRemoteID, "commit", b.Commit, "dir", dir)

	client := f.newClient(ctx, u)
	entries, err := client.ListFiles(string(r.ForgeRemoteID), dir, b.Commit)
	if err != nil {
		slog.Error("failed to list files", "error", err)
		return nil, err
	}

	var result []*forgeTypes.FileMeta
	for _, entry := range entries {
		content := []byte{}
		if entry.Type == internal.FileTypeBlob {
			fileContent, err := client.GetFileContent(string(r.ForgeRemoteID), entry.Path, b.Commit)
			if err != nil {
				slog.Error("failed to get file content", "path", entry.Path, "error", err)
				return nil, err
			}
			content, err = convertFileContentToBytes(fileContent)
			if err != nil {
				slog.Error("failed to decode file content", "path", entry.Path, "error", err)
				return nil, err
			}
		}
		result = append(result, &forgeTypes.FileMeta{
			Name: entry.Name,
			Data: content,
		})
	}
	return result, nil
}

// Status sends the commit status to the forge.
func (f *yunxiao) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, w *model.Workflow) error {
	slog.Debug("Called Status", "repo", r.ForgeRemoteID, "commit", b.Commit, "status", b.Status)

	if b.Status == model.StatusPending || b.Status == model.StatusRunning {
		return nil
	}

	statusState := mapCommitStatus(b.Status)
	statusDesc := fmt.Sprintf("Woodpecker CI pipeline #%d: %s", b.ID, b.Status)
	statusURL := f.woodpeckerHost + common.GetPipelineStatusURL(r, b, w)

	req := &internal.CreateCommitStatusRequest{
		State:       statusState,
		Context:     "woodpecker-ci",
		Description: statusDesc,
		TargetURL:   statusURL,
	}

	client := f.newClient(ctx, u)
	_, err := client.CreateCommitStatus(string(r.ForgeRemoteID), b.Commit, req)
	if err != nil {
		slog.Error("failed to create commit status", "error", err)
		return err
	}
	return nil
}

// Netrc returns a .netrc file that can be used to clone private repositories from a forge.
func (f *yunxiao) Netrc(u *model.User, _ *model.Repo) (*model.Netrc, error) {
	slog.Debug("Called Netrc")

	token := ""
	if u != nil {
		token = u.AccessToken
	}

	host := f.apiURL
	if parsed, err := url.Parse(f.apiURL); err == nil {
		host = parsed.Host
	}

	return &model.Netrc{
		Machine:  host,
		Login:    token,
		Password: "x-yunxiao-token",
	}, nil
}

// Activate activates a repository by creating the post-commit hook.
func (f *yunxiao) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	slog.Debug("Called Activate", "repo", r.ForgeRemoteID, "link", link)

	client := f.newClient(ctx, u)

	webhookReq := &internal.CreateWebhookRequest{
		URL:                   link,
		Token:                 f.hookSecret,
		PushEvents:            true,
		TagPushEvents:         true,
		MergeRequestsEvents:   true,
		NoteEvents:            false,
		EnableSSLVerification: false,
		Description:           "Woodpecker CI webhook",
	}

	_, err := client.CreateWebhook(string(r.ForgeRemoteID), webhookReq)
	if err != nil {
		slog.Error("failed to create webhook", "error", err)
		return fmt.Errorf("could not activate repository: %w", err)
	}
	return nil
}

// Deactivate deactivates a repository by removing the post-commit hook matching the given link.
func (f *yunxiao) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	slog.Debug("Called Deactivate", "repo", r.ForgeRemoteID, "link", link)

	client := f.newClient(ctx, u)

	hook, err := client.GetWebhookByURL(string(r.ForgeRemoteID), link)
	if err != nil {
		slog.Error("failed to find webhook", "error", err)
		return fmt.Errorf("could not deactivate repository: %w", err)
	}
	if hook == nil {
		return nil
	}

	_, err = client.DeleteWebhook(string(r.ForgeRemoteID), hook.ID)
	if err != nil {
		slog.Error("failed to delete webhook", "error", err)
		return fmt.Errorf("could not deactivate repository: %w", err)
	}
	return nil
}

// Branches returns the names of all branches for the named repository.
func (f *yunxiao) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	slog.Debug("Called Branches", "repo", r.ForgeRemoteID, "page", p.Page)

	client := f.newClient(ctx, u)
	branches, err := client.ListBranches(string(r.ForgeRemoteID), p.Page, p.PerPage)
	if err != nil {
		slog.Error("failed to list branches", "error", err)
		return nil, err
	}

	var result []string
	for _, branch := range branches {
		result = append(result, branch.Name)
	}
	return result, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch.
func (f *yunxiao) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	slog.Debug("Called BranchHead", "repo", r.ForgeRemoteID, "branch", branch)

	client := f.newClient(ctx, u)

	branchInfo, err := client.GetBranch(string(r.ForgeRemoteID), branch)
	if err != nil {
		// Fallback: try to get commits directly
		commits, commitErr := client.ListCommits(string(r.ForgeRemoteID), branch, 1, 1)
		if commitErr != nil {
			slog.Error("failed to get branch", "error", err)
			return nil, err
		}
		if len(commits) == 0 {
			return nil, errors.New("branch has no commits")
		}
		return &model.Commit{
			SHA:      commits[0].ID,
			ForgeURL: commits[0].WebURL,
		}, nil
	}

	return &model.Commit{
		SHA:      branchInfo.Commit.ID,
		ForgeURL: branchInfo.WebURL,
	}, nil
}

// PullRequests returns all pull requests for the named repository.
func (f *yunxiao) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	slog.Debug("Called PullRequests", "repo", r.ForgeRemoteID, "page", p.Page)

	client := f.newClient(ctx, u)
	changeRequests, err := client.ListChangeRequests(string(r.ForgeRemoteID), p.Page, p.PerPage)
	if err != nil {
		slog.Error("failed to list change requests", "error", err)
		return nil, err
	}

	var result []*model.PullRequest
	for _, cr := range changeRequests {
		result = append(result, convertChangeRequest(cr))
	}
	return result, nil
}

// Hook parses the incoming webhook and returns the repo and pipeline.
func (f *yunxiao) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	slog.Debug("Called Hook")

	// Verify X-Codeup-Token if hook secret is configured
	if f.hookSecret != "" {
		token := r.Header.Get(internal.WebhookTokenHeader)
		if token != f.hookSecret {
			return nil, nil, errors.New("invalid webhook token")
		}
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		return nil, nil, err
	}

	hookType := r.Header.Get(internal.EventTypeHeaderKey)
	slog.Debug("hook event type", "type", hookType)

	switch hookType {
	case internal.EventTypePush:
		return parsePushHook(payload)
	case internal.EventTypeTagPush:
		return parseTagPushHook(payload)
	case internal.EventTypeMergeRequest:
		return parseMergeRequestHook(payload)
	case internal.EventTypeNote:
		// Note hooks (comments) are not processed as pipelines
		return nil, nil, &forgeTypes.ErrIgnoreEvent{Event: hookType}
	default:
		return nil, nil, &forgeTypes.ErrIgnoreEvent{Event: hookType}
	}
}

// Org returns the organization.
func (f *yunxiao) Org(ctx context.Context, u *model.User, org string) (*model.Org, error) {
	slog.Debug("Called Org", "org", org)

	client := f.newClient(ctx, u)

	// If org is empty or equals the user's login, return user as org
	if org == "" || org == u.Login {
		return &model.Org{
			Name:   u.Login,
			IsUser: true,
		}, nil
	}

	// Try to find the organization
	orgs, err := client.ListOrganizations(1, 100)
	if err != nil {
		return &model.Org{
			Name:   u.Login,
			IsUser: true,
		}, nil
	}

	for _, o := range orgs {
		if o.Name == org || o.ID == org {
			return &model.Org{
				Name:   o.Name,
				IsUser: false,
			}, nil
		}
	}

	return &model.Org{
		Name:   u.Login,
		IsUser: true,
	}, nil
}

// OrgMembership returns if the user is a member of the organization and if the user is an admin.
func (f *yunxiao) OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error) {
	slog.Debug("Called OrgMembership", "org", org, "user", u.Login)

	// If looking up the user themselves, they are admin
	if org == "" || org == u.Login {
		return &model.OrgPerm{
			Member: true,
			Admin:  true,
		}, nil
	}

	client := f.newClient(ctx, u)

	var members []*internal.YunxiaoOrganizationMember
	var err error

	if f.organizationID != "" {
		members, err = client.ListOrganizationMembers(f.organizationID, 1, 100)
	} else {
		// Region edition: use platform members
		members, err = client.ListPlatformMembers(1, 100)
	}
	if err != nil {
		slog.Error("failed to list members", "error", err)
		return &model.OrgPerm{
			Member: false,
			Admin:  false,
		}, nil
	}

	for _, member := range members {
		if member.UserID == string(u.ForgeRemoteID) {
			return &model.OrgPerm{
				Member: true,
				Admin:  true, // Assume admin if they can authenticate
			}, nil
		}
	}

	return &model.OrgPerm{
		Member: false,
		Admin:  false,
	}, nil
}


