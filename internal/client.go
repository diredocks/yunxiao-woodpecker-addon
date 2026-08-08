// HTTP client for the yunxiao platform and codeup APIs.
// Handles authentication, pagination, and JSON serialization.
package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

const (
	codeupBasePath   = "/oapi/v1/codeup"
	platformBasePath = "/oapi/v1/platform"
)

type Client struct {
	*http.Client
	base           string
	ctx            context.Context
	token          string
	organizationID string
}

func NewClient(ctx context.Context, baseURL string, token string, organizationID string) *Client {
	return &Client{
		Client:         http.DefaultClient,
		base:           baseURL,
		ctx:            ctx,
		token:          token,
		organizationID: organizationID,
	}
}

// codeupPath builds a codeup API path. For center edition, it includes the organization ID.
func (c *Client) codeupPath(template string, args ...interface{}) string {
	var path string
	if c.organizationID != "" {
		path = fmt.Sprintf(codeupBasePath+"/organizations/%s"+template, append([]interface{}{c.organizationID}, args...)...)
	} else {
		path = fmt.Sprintf(codeupBasePath+template, args...)
	}
	return path
}

// platformPath builds a platform API path. For center edition, it includes the organization ID if needed.
func (c *Client) platformPath(template string, args ...interface{}) string {
	return fmt.Sprintf(platformBasePath+template, args...)
}

// ---------- User ----------

func (c *Client) GetCurrentUser() (*YunxiaoUser, error) {
	out := new(YunxiaoUser)
	uri := c.platformPath("/user")
	_, err := c.do(uri, http.MethodGet, nil, out)
	return out, err
}

// ---------- Organizations ----------

func (c *Client) ListOrganizations(page, perPage int) ([]*YunxiaoOrganization, error) {
	uri := c.platformPath("/organizations") + "?" + (&ListOpts{Page: page, PerPage: perPage}).Encode()
	out := new([]*YunxiaoOrganization)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

// ---------- Repositories ----------

func (c *Client) ListRepositories(page, perPage int) ([]*YunxiaoRepository, error) {
	p := page
	if p <= 0 {
		p = 1
	}
	pp := perPage
	if pp <= 0 {
		pp = 100
	}
	uri := c.codeupPath("/repositories") + "?" + (&ListOpts{Page: p, PerPage: pp}).Encode()
	out := new([]*YunxiaoRepository)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

func (c *Client) GetRepository(repositoryID string) (*YunxiaoRepository, error) {
	out := new(YunxiaoRepository)
	uri := c.codeupPath("/repositories/%s", repositoryID)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return out, err
}

// ---------- Branches ----------

func (c *Client) ListBranches(repositoryID string, page, perPage int) ([]*YunxiaoBranch, error) {
	uri := c.codeupPath("/repositories/%s/branches", repositoryID) + "?" + (&ListOpts{Page: page, PerPage: perPage}).Encode()
	out := new([]*YunxiaoBranch)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

func (c *Client) GetBranch(repositoryID, branchName string) (*YunxiaoBranch, error) {
	out := new(YunxiaoBranch)
	uri := c.codeupPath("/repositories/%s/branches/%s", repositoryID, url.PathEscape(branchName))
	_, err := c.do(uri, http.MethodGet, nil, out)
	return out, err
}

// ---------- Commits ----------

func (c *Client) ListCommits(repositoryID, refName string, page, perPage int) ([]*YunxiaoCommit, error) {
	params := (&ListOpts{Page: page, PerPage: perPage}).Encode()
	if refName != "" {
		if params != "" {
			params += "&"
		}
		params += "refName=" + url.QueryEscape(refName)
	}
	uri := c.codeupPath("/repositories/%s/commits", repositoryID) + "?" + params
	out := new([]*YunxiaoCommit)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

func (c *Client) GetCommit(repositoryID, sha string) (*YunxiaoCommit, error) {
	out := new(YunxiaoCommit)
	uri := c.codeupPath("/repositories/%s/commits/%s", repositoryID, sha)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return out, err
}

// ---------- Files ----------

func (c *Client) ListFiles(repositoryID, path, ref string) ([]*YunxiaoFileTreeEntry, error) {
	params := ""
	if path != "" {
		params += "path=" + url.QueryEscape(path)
	}
	if ref != "" {
		if params != "" {
			params += "&"
		}
		params += "ref=" + url.QueryEscape(ref)
	}
	uri := c.codeupPath("/repositories/%s/files/tree", repositoryID)
	if params != "" {
		uri += "?" + params
	}
	out := new([]*YunxiaoFileTreeEntry)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

func (c *Client) GetFileContent(repositoryID, filePath, ref string) (*YunxiaoFileContent, error) {
	out := new(YunxiaoFileContent)
	uri := c.codeupPath("/repositories/%s/files/%s", repositoryID, url.PathEscape(filePath)) + "?ref=" + url.QueryEscape(ref)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return out, err
}

// ---------- Change Requests ----------

func (c *Client) ListChangeRequests(projectID string, page, perPage int) ([]*YunxiaoChangeRequest, error) {
	uri := c.codeupPath("/changeRequests") + "?" + (&ListOpts{Page: page, PerPage: perPage}).Encode()
	if projectID != "" {
		uri += "&projectIds=" + projectID
	}
	out := new([]*YunxiaoChangeRequest)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

func (c *Client) GetChangeRequest(repositoryID string, localID int) (*YunxiaoChangeRequestDetail, error) {
	out := new(YunxiaoChangeRequestDetail)
	uri := c.codeupPath("/repositories/%s/changeRequests/%d", repositoryID, localID)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return out, err
}

// ---------- Commit Status ----------

func (c *Client) CreateCommitStatus(repositoryID, sha string, req *CreateCommitStatusRequest) (*YunxiaoCommitStatus, error) {
	out := new(YunxiaoCommitStatus)
	uri := c.codeupPath("/repositories/%s/commits/%s/statuses", repositoryID, sha)
	_, err := c.do(uri, http.MethodPost, req, out)
	return out, err
}

// ---------- Webhooks ----------

func (c *Client) ListWebhooks(repositoryID string, page, perPage int) ([]*YunxiaoWebhook, error) {
	uri := c.codeupPath("/repositories/%s/webhooks", repositoryID) + "?" + (&ListOpts{Page: page, PerPage: perPage}).Encode()
	out := new([]*YunxiaoWebhook)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

func (c *Client) CreateWebhook(repositoryID string, req *CreateWebhookRequest) (*YunxiaoWebhook, error) {
	out := new(YunxiaoWebhook)
	uri := c.codeupPath("/repositories/%s/webhooks", repositoryID)
	_, err := c.do(uri, http.MethodPost, req, out)
	return out, err
}

func (c *Client) DeleteWebhook(repositoryID string, hookID int) (*DeleteWebhookResponse, error) {
	out := new(DeleteWebhookResponse)
	uri := c.codeupPath("/repositories/%s/webhooks/%d", repositoryID, hookID)
	_, err := c.do(uri, http.MethodDelete, nil, out)
	return out, err
}

func (c *Client) GetWebhookByURL(repositoryID, hookURL string) (*YunxiaoWebhook, error) {
	all, err := c.ListWebhooks(repositoryID, 1, 100)
	if err != nil {
		return nil, err
	}
	for _, hook := range all {
		if hook.URL == hookURL {
			return hook, nil
		}
	}
	return nil, nil
}

// ---------- Organization Members ----------

func (c *Client) ListOrganizationMembers(orgID string, page, perPage int) ([]*YunxiaoOrganizationMember, error) {
	uri := c.platformPath("/organizations/%s/members", orgID) + "?" + (&ListOpts{Page: page, PerPage: perPage}).Encode()
	out := new([]*YunxiaoOrganizationMember)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

func (c *Client) ListPlatformMembers(page, perPage int) ([]*YunxiaoOrganizationMember, error) {
	uri := c.platformPath("/members") + "?" + (&ListOpts{Page: page, PerPage: perPage}).Encode()
	out := new([]*YunxiaoOrganizationMember)
	_, err := c.do(uri, http.MethodGet, nil, out)
	return *out, err
}

// ---------- Core HTTP ----------

func (c *Client) do(rawpath, method string, in, out interface{}) (*http.Header, error) {
	uri, err := url.Parse(c.base + rawpath)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(in); err != nil {
			return nil, err
		}
	}

	slog.Debug(fmt.Sprintf("%s %s", method, uri.String()))

	req, err := http.NewRequestWithContext(c.ctx, method, uri.String(), buf)
	if err != nil {
		return nil, err
	}
	if in != nil {
		req.Header.Set("Content-Type", AppJsonType)
	}
	req.Header.Set("x-yunxiao-token", c.token)

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		err := Error{}
		err.URL = uri.String()
		err.Method = method
		_ = json.NewDecoder(resp.Body).Decode(&err)
		err.Status = resp.StatusCode
		return nil, err
	}

	respHdr := resp.Header

	if out != nil {
		return &respHdr, json.NewDecoder(resp.Body).Decode(out)
	}

	return &respHdr, nil
}

func getTotalPages(h *http.Header) int {
	if h == nil {
		return 0
	}
	v := h.Get("x-total-pages")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}
