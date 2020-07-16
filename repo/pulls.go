package repo

import "github.com/itchyny/github-migrator/github"

// ListPullReqs lists the pull requests.
func (r *Repo) ListPullReqs() github.PullReqs {
	return r.cli.ListPullReqs(r.path, &github.ListPullReqsParams{
		State:     github.ListPullReqsParamStateAll,
		Direction: github.ListPullReqsParamDirectionAsc,
	})
}

// GetPullReq gets the pull request.
func (r *Repo) GetPullReq(pullNumber int) (*github.PullReq, error) {
	return r.cli.GetPullReq(r.path, pullNumber)
}
