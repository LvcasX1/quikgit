package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
)

type Repository struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	CloneURL    string    `json:"clone_url"`
	SSHURL      string    `json:"ssh_url"`
	Language    string    `json:"language"`
	Stars       int       `json:"stargazers_count"`
	Forks       int       `json:"forks_count"`
	UpdatedAt   time.Time `json:"updated_at"`
	License     string    `json:"license"`
	Private     bool      `json:"private"`
	Owner       string    `json:"owner"`
	Topics      []string  `json:"topics"`
}

type SearchOptions struct {
	Query        string
	Language     string
	Sort         string // stars, forks, updated
	Order        string // asc, desc
	User         string
	Organization string
	Topic        string
	Limit        int
	Page         int
}

type Client struct {
	client *github.Client
}

func NewClient(githubClient *github.Client) *Client {
	return &Client{client: githubClient}
}

func (c *Client) SearchRepositories(ctx context.Context, opts SearchOptions) ([]*Repository, int, error) {
	if c.client == nil {
		return nil, 0, fmt.Errorf("GitHub client not initialized")
	}

	query := buildSearchQuery(opts)

	searchOpts := &github.SearchOptions{
		Sort:  opts.Sort,
		Order: opts.Order,
		ListOptions: github.ListOptions{
			Page:    opts.Page,
			PerPage: opts.Limit,
		},
	}

	if searchOpts.ListOptions.PerPage == 0 {
		searchOpts.ListOptions.PerPage = 30
	}

	result, _, err := c.client.Search.Repositories(ctx, query, searchOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search repositories: %w", err)
	}

	repositories := make([]*Repository, 0, len(result.Repositories))
	for _, repo := range result.Repositories {
		repositories = append(repositories, convertRepository(repo))
	}

	return repositories, result.GetTotal(), nil
}

func (c *Client) GetUserRepositories(ctx context.Context, username string, page, perPage int) ([]*Repository, error) {
	if c.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	opts := &github.RepositoryListOptions{
		Type: "all",
		Sort: "updated",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	if perPage == 0 {
		opts.ListOptions.PerPage = 30
	}

	var repos []*github.Repository
	var err error

	if username == "" {
		repos, _, err = c.client.Repositories.List(ctx, "", opts)
	} else {
		publicOpts := &github.RepositoryListOptions{
			Type: "public",
			Sort: "updated",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}
		repos, _, err = c.client.Repositories.List(ctx, username, publicOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}

	repositories := make([]*Repository, 0, len(repos))
	for _, repo := range repos {
		repositories = append(repositories, convertRepository(repo))
	}

	return repositories, nil
}

func (c *Client) GetOrganizationRepositories(ctx context.Context, org string, page, perPage int) ([]*Repository, error) {
	if c.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	opts := &github.RepositoryListByOrgOptions{
		Type: "all",
		Sort: "updated",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	if perPage == 0 {
		opts.ListOptions.PerPage = 30
	}

	repos, _, err := c.client.Repositories.ListByOrg(ctx, org, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization repositories: %w", err)
	}

	repositories := make([]*Repository, 0, len(repos))
	for _, repo := range repos {
		repositories = append(repositories, convertRepository(repo))
	}

	return repositories, nil
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	if c.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return convertRepository(repository), nil
}

func (c *Client) GetUserOrganizations(ctx context.Context) ([]*github.Organization, error) {
	if c.client == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	opts := &github.ListOptions{PerPage: 100}

	var allOrgs []*github.Organization
	for {
		orgs, resp, err := c.client.Organizations.List(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get organizations: %w", err)
		}

		allOrgs = append(allOrgs, orgs...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allOrgs, nil
}

func buildSearchQuery(opts SearchOptions) string {
	var parts []string

	if opts.Query != "" {
		parts = append(parts, opts.Query)
	}

	if opts.Language != "" {
		parts = append(parts, fmt.Sprintf("language:%s", opts.Language))
	}

	if opts.User != "" {
		parts = append(parts, fmt.Sprintf("user:%s", opts.User))
	}

	if opts.Organization != "" {
		parts = append(parts, fmt.Sprintf("org:%s", opts.Organization))
	}

	if opts.Topic != "" {
		parts = append(parts, fmt.Sprintf("topic:%s", opts.Topic))
	}

	if len(parts) == 0 {
		parts = append(parts, "is:public")
	}

	return strings.Join(parts, " ")
}

func convertRepository(repo *github.Repository) *Repository {
	r := &Repository{
		ID:          repo.GetID(),
		Name:        repo.GetName(),
		FullName:    repo.GetFullName(),
		Description: repo.GetDescription(),
		CloneURL:    repo.GetCloneURL(),
		SSHURL:      repo.GetSSHURL(),
		Language:    repo.GetLanguage(),
		Stars:       repo.GetStargazersCount(),
		Forks:       repo.GetForksCount(),
		UpdatedAt:   repo.GetUpdatedAt().Time,
		Private:     repo.GetPrivate(),
		Topics:      repo.Topics,
	}

	if repo.Owner != nil {
		r.Owner = repo.Owner.GetLogin()
	}

	if repo.License != nil {
		r.License = repo.License.GetName()
	}

	return r
}

// Languages contains common programming languages for filtering
var Languages = []string{
	"Go", "JavaScript", "TypeScript", "Python", "Java", "C", "C++", "C#",
	"Ruby", "PHP", "Swift", "Kotlin", "Rust", "Scala", "Shell", "HTML",
	"CSS", "Dart", "R", "Perl", "Haskell", "Clojure", "Elixir", "Erlang",
	"F#", "OCaml", "Lua", "Julia", "Nim", "Crystal", "Zig", "V",
}

func GetLanguageOptions() []string {
	options := make([]string, len(Languages))
	copy(options, Languages)
	sort.Strings(options)
	return options
}

func GetSortOptions() []string {
	return []string{"stars", "forks", "updated", "created"}
}
