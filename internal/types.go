package internal

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	EventTypeHeaderKey = "Codeup-Event"
	WebhookTokenHeader = "X-Codeup-Token"

	EventTypePush        = "Push Hook"
	EventTypeTagPush     = "Tag Push Hook"
	EventTypeNote        = "Note Hook"
	EventTypeMergeRequest = "Merge Request Hook"

	ObjectKindPush         = "push"
	ObjectKindTagPush      = "tag_push"
	ObjectKindNote         = "note"
	ObjectKindMergeRequest = "merge_request"

	FileTypeTree = "tree"
	FileTypeBlob = "blob"

	CommitStateError   = "error"
	CommitStateFailure = "failure"
	CommitStatePending = "pending"
	CommitStateSuccess = "success"

	AppJsonType = "application/json"
)

// ---------- Pagination ----------

type ListOpts struct {
	Page    int
	PerPage int
}

func (o *ListOpts) Encode() string {
	params := url.Values{}
	if o.Page > 0 {
		params.Set("page", strconv.Itoa(o.Page))
	}
	if o.PerPage > 0 {
		params.Set("perPage", strconv.Itoa(o.PerPage))
	}
	return params.Encode()
}

// ---------- User ----------

type YunxiaoUser struct {
	ID             string   `json:"id"`
	Username       string   `json:"username"`
	Name           string   `json:"name"`
	NickName       string   `json:"nickName"`
	Email          string   `json:"email"`
	StaffID        string   `json:"staffId"`
	SysDeptIDs     []string `json:"sysDeptIds"`
	LastOrganization string `json:"lastOrganization"`
	CreatedAt      string   `json:"createdAt"`
	DeletedAt      string   `json:"deletedAt"`
}

// ---------- Organization ----------

type YunxiaoOrganization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatorID   string `json:"creatorId"`
	DefaultRole string `json:"defaultRole"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updateAt"`
}

// ---------- Repository ----------

type YunxiaoRepository struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Path             string `json:"path"`
	NameWithNamespace string `json:"nameWithNamespace"`
	PathWithNamespace string `json:"pathWithNamespace"`
	Description      string `json:"description"`
	DefaultBranch    string `json:"defaultBranch"`
	HTTPUrlToRepo    string `json:"httpUrlToRepo"`
	SSHUrlToRepo     string `json:"sshUrlToRepo"`
	WebURL           string `json:"webUrl"`
	Visibility       string `json:"visibility"`
	Archived         bool   `json:"archived"`
	NamespaceID      int    `json:"namespaceId"`
	CreatorID        int    `json:"creatorId"`
	AccessLevel      int    `json:"accessLevel"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
	LastActivityAt   string `json:"lastActivityAt"`
	AvatarURL        string `json:"avatarUrl"`
	StarCount        int    `json:"starCount"`
	Starred          bool   `json:"starred"`
	ForkCount        int    `json:"forkCount"`
	DemoProject      bool   `json:"demoProject"`
	Encrypted        bool   `json:"encrypted"`

	Namespace struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Path        string `json:"path"`
		Description string `json:"description"`
		Avatar      string `json:"avatar"`
		Visibility  string `json:"visibility"`
		OwnerID     int    `json:"ownerId"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	} `json:"namespace"`

	Owner struct {
		ID       int    `json:"id"`
		UserID   string `json:"userId"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		AvatarURL string `json:"avatarUrl"`
		State    string `json:"state"`
		WebURL   string `json:"webUrl"`
	} `json:"owner"`

	Permissions struct {
		ProjectAccess struct {
			AccessLevel       int `json:"accessLevel"`
			NotificationLevel int `json:"notificationLevel"`
		} `json:"projectAccess"`
		GroupAccess struct {
			AccessLevel       int `json:"accessLevel"`
			NotificationLevel int `json:"notificationLevel"`
		} `json:"groupAccess"`
	} `json:"permissions"`

	AdminSettingLanguage        string `json:"adminSettingLanguage"`
	AllowPush                   bool   `json:"allowPush"`
	CloneDownloadControlGray    bool   `json:"cloneDownloadControlGray"`
	EnableCloneDownloadControl  bool   `json:"enableCloneDownloadControl"`
	OpenCloneDownloadControl    bool   `json:"openCloneDownloadControl"`
	ProjectType                 int    `json:"projectType"`
}

// ---------- Branch ----------

type YunxiaoBranch struct {
	Name           string              `json:"name"`
	DefaultBranch  bool                `json:"defaultBranch"`
	Protected      bool                `json:"protected"`
	WebURL         string              `json:"webUrl"`
	Commit         YunxiaoBranchCommit `json:"commit"`
}

type YunxiaoBranchCommit struct {
	ID             string   `json:"id"`
	ShortID        string   `json:"shortId"`
	Title          string   `json:"title"`
	Message        string   `json:"message"`
	AuthorName     string   `json:"authorName"`
	AuthorEmail    string   `json:"authorEmail"`
	AuthoredDate   string   `json:"authoredDate"`
	CommitterName  string   `json:"committerName"`
	CommitterEmail string   `json:"committerEmail"`
	CommittedDate  string   `json:"committedDate"`
	ParentIDs      []string `json:"parentIds"`
	Stats          struct {
		Additions int `json:"additions"`
		Deletions int `json:"deletions"`
		Total     int `json:"total"`
	} `json:"stats"`
}

// ---------- Commit ----------

type YunxiaoCommit struct {
	ID             string   `json:"id"`
	ShortID        string   `json:"shortId"`
	Title          string   `json:"title"`
	Message        string   `json:"message"`
	AuthorName     string   `json:"authorName"`
	AuthorEmail    string   `json:"authorEmail"`
	AuthoredDate   string   `json:"authoredDate"`
	CommitterName  string   `json:"committerName"`
	CommitterEmail string   `json:"committerEmail"`
	CommittedDate  string   `json:"committedDate"`
	ParentIDs      []string `json:"parentIds"`
	WebURL         string   `json:"webUrl"`
	Stats          struct {
		Additions int `json:"additions"`
		Deletions int `json:"deletions"`
		Total     int `json:"total"`
	} `json:"stats"`
}

// ---------- File ----------

type YunxiaoFileTreeEntry struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Path  string `json:"path"`
	Type  string `json:"type"`
	IsLFS bool   `json:"isLFS"`
	Mode  string `json:"mode"`
}

type YunxiaoFileContent struct {
	BlobID       string `json:"blobId"`
	CommitID     string `json:"commitId"`
	Content      string `json:"content"`
	Encoding     string `json:"encoding"`
	FileName     string `json:"fileName"`
	FilePath     string `json:"filePath"`
	LastCommitID string `json:"lastCommitId"`
	Ref          string `json:"ref"`
	Size         int    `json:"size"`
}

// ---------- Change Request (Merge Request) ----------

type YunxiaoChangeRequest struct {
	LocalID                int                     `json:"localId"`
	ProjectID              int                     `json:"projectId"`
	Title                  string                  `json:"title"`
	Description            string                  `json:"description"`
	State                  string                  `json:"state"`
	Status                 string                  `json:"status"`
	SourceBranch           string                  `json:"sourceBranch"`
	SourceProjectID        int                     `json:"sourceProjectId"`
	SourceType             string                  `json:"sourceType"`
	TargetBranch           string                  `json:"targetBranch"`
	TargetProjectID        int                     `json:"targetProjectId"`
	TargetType             string                  `json:"targetType"`
	CreatedAt              string                  `json:"createdAt"`
	UpdatedAt              string                  `json:"updatedAt"`
	WebURL                 string                  `json:"webUrl"`
	DetailURL              string                  `json:"detailUrl"`
	SSHURL                 string                  `json:"sshUrl"`
	MergedRevision         string                  `json:"mergedRevision"`
	HasConflict            bool                    `json:"hasConflict"`
	CreationMethod         string                  `json:"creationMethod"`
	WorkInProgress         bool                    `json:"workInProgress"`
	TotalCommentCount      int                     `json:"totalCommentCount"`
	UnResolvedCommentCount int                     `json:"unResolvedCommentCount"`
	SupportMergeFFOnly     bool                    `json:"supportMergeFFOnly"`
	Author                 YunxiaoUserSimple       `json:"author"`
	Reviewers              []YunxiaoReviewer       `json:"reviewers"`
}

// ---------- Change Request Detail ----------

type YunxiaoChangeRequestDetail struct {
	LocalID                       int                      `json:"localId"`
	ProjectID                     int                      `json:"projectId"`
	Title                         string                   `json:"title"`
	Description                   string                   `json:"description"`
	Status                        string                   `json:"status"`
	SourceBranch                  string                   `json:"sourceBranch"`
	SourceProjectID               int                      `json:"sourceProjectId"`
	TargetBranch                  string                   `json:"targetBranch"`
	TargetProjectID               int                      `json:"targetProjectId"`
	TargetProjectPathWithNamespace string                  `json:"targetProjectPathWithNamespace"`
	TargetProjectNameWithNamespace string                  `json:"targetProjectNameWithNamespace"`
	CreateTime                    string                   `json:"createTime"`
	UpdateTime                    string                   `json:"updateTime"`
	CreateFrom                    string                   `json:"createFrom"`
	WebURL                        string                   `json:"webUrl"`
	DetailURL                     string                   `json:"detailUrl"`
	MergedRevision                string                   `json:"mergedRevision"`
	Ahead                         int                      `json:"ahead"`
	Behind                        int                      `json:"behind"`
	ConflictCheckStatus           string                   `json:"conflictCheckStatus"`
	AllRequirementsPass           bool                     `json:"allRequirementsPass"`
	CanRevertOrCherryPick         bool                     `json:"canRevertOrCherryPick"`
	HasReverted                   bool                     `json:"hasReverted"`
	MrType                        string                   `json:"mrType"`
	SupportMergeFastForwardOnly   bool                     `json:"supportMergeFastForwardOnly"`
	TotalCommentCount             int                      `json:"totalCommentCount"`
	UnResolvedCommentCount        int                      `json:"unResolvedCommentCount"`
	Author                        YunxiaoUserSimple        `json:"author"`
	Reviewers                     []YunxiaoReviewer        `json:"reviewers"`
	Subscribers                   []YunxiaoUserSimple      `json:"subscribers"`
}

type YunxiaoUserSimple struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	State    string `json:"state"`
}

type YunxiaoReviewer struct {
	UserID              string `json:"userId"`
	Username            string `json:"username"`
	Name                string `json:"name"`
	Email               string `json:"email"`
	Avatar              string `json:"avatar"`
	State               string `json:"state"`
	HasCommented        bool   `json:"hasCommented"`
	HasReviewed         bool   `json:"hasReviewed"`
	ReviewOpinionStatus string `json:"reviewOpinionStatus"`
	ReviewTime          string `json:"reviewTime"`
}

// ---------- Commit Status ----------

type YunxiaoCommitStatus struct {
	ID          int                  `json:"id"`
	SHA         string               `json:"sha"`
	Context     string               `json:"context"`
	State       string               `json:"state"`
	Description string               `json:"description"`
	TargetURL   string               `json:"targetUrl"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
	Author      YunxiaoStatusAuthor  `json:"author"`
}

type YunxiaoStatusAuthor struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	Type     string `json:"type"`
}

type CreateCommitStatusRequest struct {
	State       string `json:"state"`
	Context     string `json:"context,omitempty"`
	Description string `json:"description,omitempty"`
	TargetURL   string `json:"targetUrl,omitempty"`
}

// ---------- Webhook ----------

type YunxiaoWebhook struct {
	ID                   int    `json:"id"`
	URL                  string `json:"url"`
	Token                string `json:"token"`
	PushEvents           bool   `json:"pushEvents"`
	TagPushEvents        bool   `json:"tagPushEvents"`
	NoteEvents           bool   `json:"noteEvents"`
	MergeRequestEvents   bool   `json:"mergeRequestEvents"`
	Description          string `json:"description"`
	EnableSSLVerification bool  `json:"enableSSLVerification"`
	RepositoryID         int    `json:"repositoryId"`
	CreatedAt            string `json:"createdAt"`
	UpdatedAt            string `json:"updatedAt"`
}

type CreateWebhookRequest struct {
	URL                     string `json:"url"`
	Token                   string `json:"token,omitempty"`
	PushEvents              bool   `json:"pushEvents"`
	TagPushEvents           bool   `json:"tagPushEvents"`
	NoteEvents              bool   `json:"noteEvents"`
	MergeRequestsEvents     bool   `json:"mergeRequestsEvents"`
	EnableSSLVerification   bool   `json:"enableSslVerification"`
	Description             string `json:"description,omitempty"`
}

type DeleteWebhookResponse struct {
	Result bool `json:"result"`
}

// ---------- Webhook Payload Types ----------

type HookPushPayload struct {
	ObjectKind       string                    `json:"object_kind"`
	Before           string                    `json:"before"`
	After            string                    `json:"after"`
	Ref              string                    `json:"ref"`
	UserID           int                       `json:"user_id"`
	UserName         string                    `json:"user_name"`
	UserEmail        string                    `json:"user_email"`
	ProjectID        int                       `json:"project_id"`
	Repository       HookRepository            `json:"repository"`
	Commits          []HookCommit              `json:"commits"`
	TotalCommitsCount int                      `json:"total_commits_count"`
}

type HookTagPushPayload struct {
	ObjectKind       string         `json:"object_kind"`
	Before           string         `json:"before"`
	After            string         `json:"after"`
	Ref              string         `json:"ref"`
	UserID           int            `json:"user_id"`
	UserName         string         `json:"user_name"`
	ProjectID        int            `json:"project_id"`
	Repository       HookRepository `json:"repository"`
	Commits          []HookCommit   `json:"commits"`
	TotalCommitsCount int           `json:"total_commits_count"`
}

type HookMergeRequestPayload struct {
	ObjectKind       string                `json:"object_kind"`
	User             HookUser              `json:"user"`
	Repository       HookRepository        `json:"repository"`
	ObjectAttributes HookMergeRequestAttr  `json:"object_attributes"`
}

type HookMergeRequestAttr struct {
	Action              string           `json:"action"`
	AuthorAliyunPK      string           `json:"author_aliyun_pk"`
	AuthorID            int              `json:"author_id"`
	BizID               string           `json:"biz_id"`
	CreatedAt           string           `json:"created_at"`
	Description         string           `json:"description"`
	IsUpdateByPush      bool             `json:"is_update_by_push"`
	LastCommit          HookLastCommit   `json:"last_commit"`
	LocalID             int              `json:"local_id"`
	MergeStatus         string           `json:"merge_status"`
	ProjectID           int              `json:"project_id"`
	Source              HookMRSide       `json:"source"`
	SourceBranch        string           `json:"source_branch"`
	SourceProjectID     int              `json:"source_project_id"`
	SourceType          string           `json:"source_type"`
	State               string           `json:"state"`
	Target              HookMRSide       `json:"target"`
	TargetBranch        string           `json:"target_branch"`
	TargetProjectID     int              `json:"target_project_id"`
	Title               string           `json:"title"`
	UpdatedAt           string           `json:"updated_at"`
	URL                 string           `json:"url"`
	WorkInProgress      bool             `json:"work_in_progress"`
}

type HookLastCommit struct {
	Author    HookCommitAuthor `json:"author"`
	ID        string           `json:"id"`
	Message   string           `json:"message"`
	Timestamp string           `json:"timestamp"`
	URL       string           `json:"url"`
}

type HookMRSide struct {
	HTTPURL         string `json:"http_url"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	SSHURL          string `json:"ssh_url"`
	VisibilityLevel int    `json:"visibility_level"`
	WebURL          string `json:"web_url"`
}

type HookRepository struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	GitHTTPURL  string `json:"git_http_url"`
	GitSSHURL   string `json:"git_ssh_url"`
	VisibilityLevel int `json:"visibility_level"`
}

type HookCommit struct {
	ID        string           `json:"id"`
	Message   string           `json:"message"`
	Timestamp string           `json:"timestamp"`
	URL       string           `json:"url"`
	Author    HookCommitAuthor `json:"author"`
}

type HookCommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type HookUser struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

// ---------- Organization Member ----------

type YunxiaoOrganizationMember struct {
	ID             string   `json:"id"`
	UserID         string   `json:"userId"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	OrganizationID string   `json:"organizationId"`
	DeptIDs        []string `json:"deptIds"`
	RoleIDs        []string `json:"roleIds"`
	Status         string   `json:"status"`
	Joined         string   `json:"joined"`
	LastUpdated    string   `json:"lastUpdated"`
	Visited        string   `json:"visited"`
}

// ---------- Error ----------

type Error struct {
	Status int
	URL    string
	Method string
	Body   struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e Error) Error() string {
	if len(e.Body.Message) > 0 {
		return e.Body.Message
	}
	return fmt.Sprintf("an unknown error occurred with the forge - %s %s %d", e.Method, e.URL, e.Status)
}
