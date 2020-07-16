package repo

import "github.com/itchyny/github-migrator/github"

// ListPullReqCommits lists the commits of a pull request.
func (r *Repo) ListPullReqCommits(pullNumber int) github.Commits {
	return r.cli.ListPullReqCommits(r.path, pullNumber)
}
