package commenter

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/go-github/v38/github"
)

// Commenter is the main commenter struct
type Commenter struct {
	ghConnector      *connector
	existingComments []*existingComment
	files            []*CommitFileInfo
}

type CommitFileInfo struct {
	fileName      string
	hunkStartLine int
	hunkEndLine   int
	sha           string
}

type PRReviewComment struct {
	FileName  string
	StartLine int
	EndLine   int
	Body      string
}

var (
	patchRegex     = regexp.MustCompile(`^@@.*\+(\d+),(\d+).+?@@`)
	commitRefRegex = regexp.MustCompile(".+ref=(.+)")
)

const (
	Approve            = "APPROVE"
	RequestChanges     = "REQUEST_CHANGES"
	ApproveBody        = "Approve:tada:"
	RequestChangesBody = "Request changes:rotating_light:"
)

// NewCommenter creates a Commenter for updating PR with comments
func NewCommenter(token, owner, repo string, prNumber int) (*Commenter, error) {

	if len(token) == 0 {
		return nil, errors.New("the GITHUB_TOKEN has not been set")
	}

	ghConnector, err := createConnector(token, owner, repo, prNumber)
	if err != nil {
		return nil, err
	}

	commitFileInfos, existingComments, err := ghConnector.getPRInfo()
	if err != nil {
		return nil, err
	}

	return &Commenter{
		ghConnector:      ghConnector,
		existingComments: existingComments,
		files:            commitFileInfos,
	}, nil
}

func (c *Commenter) CreateDraftPRReviewComments(comments []PRReviewComment) []*github.DraftReviewComment {
	var draftReviewComments []*github.DraftReviewComment
	for i := range comments {
		comment := comments[i]
		if c.checkCommentRelevant(comment.FileName, comment.StartLine, comment.EndLine) {
			reviewCommentSide := "RIGHT"
			draftReviewComment := &github.DraftReviewComment{
				Body: &comment.Body,
				Path: &comment.FileName,
				Line: &comment.EndLine,
				Side: &reviewCommentSide,
			}
			if comment.StartLine < comment.EndLine {
				reviewCommentStartSide := "RIGHT"
				draftReviewComment.StartLine = &comment.StartLine
				draftReviewComment.StartSide = &reviewCommentStartSide
			}
			draftReviewComments = append(draftReviewComments, draftReviewComment)
		}
	}
	return draftReviewComments
}

func (c *Commenter) checkCommentRelevant(filename string, startLine int, endLine int) bool {
	for _, file := range c.files {
		if file.fileName == filename && startLine >= file.hunkStartLine && startLine <= file.hunkEndLine && endLine >= file.hunkStartLine && endLine <= file.hunkEndLine {
			return true
		}
	}
	return false
}

func (c *Commenter) WritePRReview(comments []*github.DraftReviewComment, event string) error {

	ctx := context.Background()
	errs := c.removeAlreadyExistComments(ctx)
	for _, err := range errs {
		fmt.Printf("%s\n", err)
	}
	body, err := selectBodyBy(event)
	if err != nil {
		return err
	}
	err = c.ghConnector.CreatePRReview(ctx, event, body, comments)
	return err
}

func (c *Commenter) removeAlreadyExistComments(ctx context.Context) []error {
	var errs []error
	for _, comment := range c.existingComments {
		err := c.ghConnector.DeletePRReviewComment(ctx, comment.commentId)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func selectBodyBy(event string) (string, error) {
	switch event {
	case Approve:
		return ApproveBody, nil
	case RequestChanges:
		return RequestChangesBody, nil
	default:
		return "", fmt.Errorf("this event type is not supported")
	}
}
