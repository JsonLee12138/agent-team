package internal

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

const defaultGitHubAPIBaseURL = "https://api.github.com"

type RoleRepoGitHubClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func NewRoleRepoGitHubClient() *RoleRepoGitHubClient {
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		token = strings.TrimSpace(os.Getenv("GH_TOKEN"))
	}
	return &RoleRepoGitHubClient{
		httpClient: http.DefaultClient,
		baseURL:    defaultGitHubAPIBaseURL,
		token:      token,
	}
}

func NewRoleRepoGitHubClientForTest(baseURL string, httpClient *http.Client, token string) *RoleRepoGitHubClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &RoleRepoGitHubClient{httpClient: httpClient, baseURL: strings.TrimRight(baseURL, "/"), token: strings.TrimSpace(token)}
}

type RoleRepoGitHubError struct {
	StatusCode int
	Message    string
}

func (e *RoleRepoGitHubError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("github api error (%d)", e.StatusCode)
	}
	return fmt.Sprintf("github api error (%d): %s", e.StatusCode, e.Message)
}

func (e *RoleRepoGitHubError) IsAuthOrRateLimit() bool {
	return e.StatusCode == 401 || e.StatusCode == 403
}

func (c *RoleRepoGitHubClient) doJSON(ctx context.Context, method, requestPath string, out any) error {
	fullURL := c.baseURL + requestPath
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "agent-team-role-repo")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(body))
		if strings.HasPrefix(msg, "{") {
			var payload struct {
				Message string `json:"message"`
			}
			if json.Unmarshal(body, &payload) == nil && payload.Message != "" {
				msg = payload.Message
			}
		}
		return &RoleRepoGitHubError{StatusCode: resp.StatusCode, Message: msg}
	}

	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *RoleRepoGitHubClient) getDefaultBranch(ctx context.Context, source RoleRepoSource) (string, error) {
	var payload struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/repos/"+source.FullName(), &payload); err != nil {
		return "", err
	}
	if payload.DefaultBranch == "" {
		return "", fmt.Errorf("repository %s has no default branch", source.FullName())
	}
	return payload.DefaultBranch, nil
}

type roleRepoTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
}

func (c *RoleRepoGitHubClient) getRepoTree(ctx context.Context, source RoleRepoSource, ref string) ([]roleRepoTreeEntry, error) {
	q := url.Values{}
	q.Set("recursive", "1")
	path := fmt.Sprintf("/repos/%s/git/trees/%s?%s", source.FullName(), url.PathEscape(ref), q.Encode())
	var payload struct {
		Tree []roleRepoTreeEntry `json:"tree"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, &payload); err != nil {
		return nil, err
	}
	if payload.Tree == nil {
		payload.Tree = []roleRepoTreeEntry{}
	}
	return payload.Tree, nil
}

func (c *RoleRepoGitHubClient) getBlobContent(ctx context.Context, source RoleRepoSource, sha string) ([]byte, error) {
	var payload struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	path := fmt.Sprintf("/repos/%s/git/blobs/%s", source.FullName(), url.PathEscape(sha))
	if err := c.doJSON(ctx, http.MethodGet, path, &payload); err != nil {
		return nil, err
	}
	if payload.Encoding != "base64" {
		return nil, fmt.Errorf("unsupported blob encoding: %s", payload.Encoding)
	}
	content := strings.ReplaceAll(payload.Content, "\n", "")
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("decode blob %s: %w", sha, err)
	}
	return decoded, nil
}

func hashRoleRepoTreeFiles(files []roleRepoTreeEntry) string {
	lines := make([]string, 0, len(files))
	for _, f := range files {
		if f.Type != "blob" {
			continue
		}
		lines = append(lines, f.Path+"\t"+f.SHA)
	}
	sort.Strings(lines)
	return roleRepoSHA256Hex(strings.Join(lines, "\n"))
}

func roleRepoSHA256Hex(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}
