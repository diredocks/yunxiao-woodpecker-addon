// Forge implementation for the woodpecker CI forge addon interface.
// Includes repository management, webhook lifecycle, commit status reporting,
// user/organization operations, and type conversion helpers.
package internal

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/common"
	forgeTypes "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const YunxiaoImage = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

var _ forge.Forge = (*Forge)(nil)

type Forge struct {
	APIURL         string
	OrganizationID string
	WoodpeckerHost string
	HookSecret     string
	LoginPort      string
}

type ForgeOpts struct {
	APIURL         string
	OrganizationID string
	WoodpeckerHost string
	HookSecret     string
	LoginPort      string
}

func New(opts ForgeOpts) (*Forge, error) {
	f := &Forge{
		APIURL:         strings.TrimSuffix(opts.APIURL, "/"),
		OrganizationID: opts.OrganizationID,
		WoodpeckerHost: strings.TrimSuffix(opts.WoodpeckerHost, "/"),
		HookSecret:     opts.HookSecret,
		LoginPort:      opts.LoginPort,
	}

	if f.APIURL == "" {
		return nil, errors.New("must provide a value for YUNXIAO_API_URL")
	}
	if _, err := url.Parse(f.APIURL); err != nil {
		return nil, fmt.Errorf("must provide a valid YUNXIAO_API_URL value: %w", err)
	}

	return f, nil
}

func (f *Forge) newClient(ctx context.Context, u *model.User) *Client {
	token := ""
	if u != nil {
		token = u.AccessToken
	}
	return NewClient(ctx, f.APIURL, token, f.OrganizationID)
}

func (f *Forge) Name() string {
	return "yunxiao"
}

func (f *Forge) URL() string {
	return f.APIURL
}

func (f *Forge) Login(ctx context.Context, req *forgeTypes.OAuthRequest) (*model.User, string, error) {
	slog.Debug("Called Login")

	woodpeckerURL, _ := url.Parse(f.WoodpeckerHost)
	scheme := woodpeckerURL.Scheme
	if scheme == "" {
		scheme = "http"
	}
	loginHost := fmt.Sprintf("%s://%s:%s", scheme, woodpeckerURL.Hostname(), f.LoginPort)
	loginURL := fmt.Sprintf("%s/yunxiao/login?woodpecker_host=%s", loginHost, url.QueryEscape(f.WoodpeckerHost))

	if req == nil || req.Code == "" {
		return nil, loginURL, nil
	}

	client := NewClient(ctx, f.APIURL, req.Code, f.OrganizationID)
	userInfo, err := client.GetCurrentUser()
	if err != nil {
		return nil, loginURL, fmt.Errorf("invalid token: %w", err)
	}

	slog.Info("Logged in", "user", userInfo.Name, "id", userInfo.ID)

	return &model.User{
		ForgeRemoteID: model.ForgeRemoteID(userInfo.ID),
		Login:         userInfo.Name,
		Email:         userInfo.Email,
		Avatar:        YunxiaoImage,
		AccessToken:   req.Code,
	}, loginURL, nil
}

func (f *Forge) Auth(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (f *Forge) Teams(_ context.Context, _ *model.User, _ *model.ListOptions) ([]*model.Team, error) {
	return nil, nil
}

func (f *Forge) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
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

func (f *Forge) Repos(ctx context.Context, u *model.User, opts *model.ListOptions) ([]*model.Repo, error) {
	slog.Debug("Called Repos", "page", opts.Page, "perPage", opts.PerPage)

	client := f.newClient(ctx, u)
	repos, err := client.ListRepositories(opts.Page, opts.PerPage)
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

func (f *Forge) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, file string) ([]byte, error) {
	slog.Debug("Called File", "repo", r.ForgeRemoteID, "commit", b.Commit, "file", file)

	client := f.newClient(ctx, u)
	fileContent, err := client.GetFileContent(string(r.ForgeRemoteID), file, b.Commit)
	if err != nil {
		slog.Error("failed to get file", "error", err)
		return nil, err
	}
	return convertFileContentToBytes(fileContent)
}

func (f *Forge) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, dir string) ([]*forgeTypes.FileMeta, error) {
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
		if entry.Type == FileTypeBlob {
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

func (f *Forge) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, w *model.Workflow) error {
	slog.Debug("Called Status", "repo", r.ForgeRemoteID, "commit", b.Commit, "status", b.Status)

	if b.Status == model.StatusPending || b.Status == model.StatusRunning {
		return nil
	}

	statusState := mapCommitStatus(b.Status)
	statusDesc := fmt.Sprintf("Woodpecker CI pipeline #%d: %s", b.ID, b.Status)
	statusURL := f.WoodpeckerHost + common.GetPipelineStatusURL(r, b, w)

	req := &CreateCommitStatusRequest{
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

func (f *Forge) Netrc(u *model.User, _ *model.Repo) (*model.Netrc, error) {
	slog.Debug("Called Netrc")

	token := ""
	if u != nil {
		token = u.AccessToken
	}

	host := f.APIURL
	if parsed, err := url.Parse(f.APIURL); err == nil {
		host = parsed.Host
	}

	return &model.Netrc{
		Machine:  host,
		Login:    token,
		Password: "x-yunxiao-token",
	}, nil
}

func (f *Forge) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	slog.Debug("Called Activate", "repo", r.ForgeRemoteID, "link", link)

	client := f.newClient(ctx, u)

	webhookReq := &CreateWebhookRequest{
		URL:                   link,
		Token:                 f.HookSecret,
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

func (f *Forge) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
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

func (f *Forge) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
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

func (f *Forge) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	slog.Debug("Called BranchHead", "repo", r.ForgeRemoteID, "branch", branch)

	client := f.newClient(ctx, u)

	branchInfo, err := client.GetBranch(string(r.ForgeRemoteID), branch)
	if err != nil {
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

func (f *Forge) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
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

func (f *Forge) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	slog.Debug("Called Hook")

	if f.HookSecret != "" {
		token := r.Header.Get(WebhookTokenHeader)
		if token != f.HookSecret {
			return nil, nil, errors.New("invalid webhook token")
		}
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		return nil, nil, err
	}

	hookType := r.Header.Get(EventTypeHeaderKey)
	slog.Debug("hook event type", "type", hookType)

	switch hookType {
	case EventTypePush:
		return parsePushHook(payload)
	case EventTypeTagPush:
		return parseTagPushHook(payload)
	case EventTypeMergeRequest:
		return parseMergeRequestHook(payload)
	case EventTypeNote:
		return nil, nil, &forgeTypes.ErrIgnoreEvent{Event: hookType}
	default:
		return nil, nil, &forgeTypes.ErrIgnoreEvent{Event: hookType}
	}
}

func (f *Forge) Org(ctx context.Context, u *model.User, org string) (*model.Org, error) {
	slog.Debug("Called Org", "org", org)

	client := f.newClient(ctx, u)

	if org == "" || org == u.Login {
		return &model.Org{
			Name:   u.Login,
			IsUser: true,
		}, nil
	}

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

func (f *Forge) OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error) {
	slog.Debug("Called OrgMembership", "org", org, "user", u.Login)

	if org == "" || org == u.Login {
		return &model.OrgPerm{
			Member: true,
			Admin:  true,
		}, nil
	}

	client := f.newClient(ctx, u)

	var members []*YunxiaoOrganizationMember
	var err error

	if f.OrganizationID != "" {
		members, err = client.ListOrganizationMembers(f.OrganizationID, 1, 100)
	} else {
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
				Admin:  true,
			}, nil
		}
	}

	return &model.OrgPerm{
		Member: false,
		Admin:  false,
	}, nil
}

func convertRepository(repo *YunxiaoRepository) *model.Repo {
	forgeRemoteID := model.ForgeRemoteID(strconv.Itoa(repo.ID))
	owner := repo.Namespace.Path
	if owner == "" {
		owner = repo.Owner.Username
	}

	accessLevel := max(
		repo.Permissions.ProjectAccess.AccessLevel,
		repo.Permissions.GroupAccess.AccessLevel,
		repo.AccessLevel,
	)
	perm := &model.Perm{
		Pull:  accessLevel >= 10,
		Push:  accessLevel >= 30,
		Admin: accessLevel >= 40,
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
		Perm:          perm,
	}
}

func convertBranchToModel(branch *YunxiaoBranch) string {
	return branch.Name
}

func convertCommitToModel(commit *YunxiaoCommit) *model.Commit {
	return &model.Commit{
		SHA:      commit.ID,
		ForgeURL: commit.WebURL,
	}
}

func convertChangeRequest(cr *YunxiaoChangeRequest) *model.PullRequest {
	return &model.PullRequest{
		Index: model.ForgeRemoteID(strconv.Itoa(cr.LocalID)),
		Title: cr.Title,
	}
}

func convertFileContentToBytes(file *YunxiaoFileContent) ([]byte, error) {
	if file.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(file.Content)
	}
	return []byte(file.Content), nil
}

func convertFileTreeEntries(entries []*YunxiaoFileTreeEntry) []*forgeTypes.FileMeta {
	var result []*forgeTypes.FileMeta
	for _, entry := range entries {
		result = append(result, &forgeTypes.FileMeta{
			Name: entry.Name,
		})
	}
	return result
}

func mapCommitStatus(status model.StatusValue) string {
	switch status {
	case model.StatusFailure, model.StatusError, model.StatusKilled, model.StatusDeclined, model.StatusBlocked:
		return CommitStateFailure
	case model.StatusSuccess:
		return CommitStateSuccess
	case model.StatusPending, model.StatusRunning:
		return CommitStatePending
	default:
		return CommitStateError
	}
}
