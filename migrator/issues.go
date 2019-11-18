package migrator

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/itchyny/github-migrator/github"
)

var (
	beforeImportIssueDuration      = 1 * time.Second
	waitImportIssueInitialDuration = 3 * time.Second
)

func (m *migrator) migrateIssues() error {
	sourceRepo, err := m.getSourceRepo()
	if err != nil {
		return err
	}
	targetRepo, err := m.getTargetRepo()
	if err != nil {
		return err
	}
	sourceIssues := m.source.ListIssues()
	targetIssuesBuffer := newIssuesBuffer(m.target.ListIssues())
	commentFilters := newCommentFilters(
		newRepoURLFilter(sourceRepo, targetRepo),
		newUserMappingFilter(m.userMapping),
	)
	for {
		issue, err := sourceIssues.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		result, err := m.migrateIssue(sourceRepo, targetRepo, commentFilters, issue, targetIssuesBuffer)
		if err != nil {
			return err
		}
		if result != nil {
			if err := m.waitImportIssue(result.ID, issue); err != nil {
				return fmt.Errorf("importing %s failed: %w", issue.HTMLURL, err)
			}
		}
	}
	return nil
}

func (m *migrator) migrateIssue(
	sourceRepo, targetRepo *github.Repo, commentFilters commentFilters,
	sourceIssue *github.Issue, targetIssuesBuffer *issuesBuffer,
) (*github.ImportResult, error) {
	fmt.Printf("migrating: %s\n", sourceIssue.HTMLURL)
	targetIssue, err := targetIssuesBuffer.get(sourceIssue.Number)
	if err != nil {
		return nil, err
	}
	if targetIssue != nil {
		fmt.Printf("skipping: %s (already exists)\n", targetIssue.HTMLURL)
		return nil, nil
	}
	comments, err := github.CommentsToSlice(m.source.ListComments(sourceIssue.Number))
	if err != nil {
		return nil, err
	}
	events, err := github.EventsToSlice(m.source.ListEvents(sourceIssue.Number))
	if err != nil {
		return nil, err
	}
	var sourcePullReq *github.PullReq
	var commits []*github.Commit
	var commitDiff string
	var reviews []*github.Review
	var reviewComments []*github.ReviewComment
	if sourceIssue.PullRequest != nil {
		sourcePullReq, err = m.source.GetPullReq(sourceIssue.Number)
		if err != nil {
			return nil, err
		}
		commits, err = github.CommitsToSlice(m.source.ListPullReqCommits(sourceIssue.Number))
		if err != nil {
			return nil, err
		}
		commitDiff, err = m.source.NewPath(sourcePullReq.Base.Repo.FullName).
			GetCompare(sourcePullReq.Base.SHA, sourcePullReq.Head.SHA)
		if err != nil {
			return nil, err
		}
		reviews, err = github.ReviewsToSlice(m.source.ListReviews(sourceIssue.Number))
		if err != nil {
			return nil, err
		}
		reviewComments, err = github.ReviewCommentsToSlice(m.source.ListReviewComments(sourceIssue.Number))
		if err != nil {
			return nil, err
		}
	}
	members, err := m.listTargetMembers()
	if err != nil {
		return nil, err
	}
	time.Sleep(beforeImportIssueDuration)
	imp, err := buildImport(
		sourceRepo, targetRepo, commentFilters,
		sourceIssue, sourcePullReq, comments, events,
		commits, commitDiff, reviews, reviewComments, members,
	)
	if err != nil {
		return nil, err
	}
	return m.target.Import(imp)
}

func (m *migrator) waitImportIssue(id int, issue *github.Issue) error {
	var retry int
	duration := waitImportIssueInitialDuration
	for {
		time.Sleep(duration)
		if retry > 1 {
			duration *= 2
		}
		res, err := m.target.GetImport(id)
		if err != nil {
			return err
		}
		fmt.Printf("status check: %s (importing %s)\n", res.Status, issue.HTMLURL)
		switch res.Status {
		case "imported":
			return nil
		case "failed":
			return errors.New("failed status")
		}
		retry++
		if retry >= 5 {
			return errors.New("reached maximum retry count")
		}
	}
}
