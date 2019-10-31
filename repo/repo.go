package repo

import "github.com/itchyny/github-migrator/github"

// Repo represents a GitHub repository.
type Repo interface {
	Get() (*github.Repo, error)
	Update(*github.UpdateRepoParams) (*github.Repo, error)
	ListIssues() github.Issues
	ListComments(int) github.Comments
	ListPullReqs() github.PullReqs
	Import(*github.Import) error
}

// New creates a new Repo.
func New(cli github.Client, path string) Repo {
	return &repo{cli: cli, path: path}
}

type repo struct {
	cli  github.Client
	path string
}
