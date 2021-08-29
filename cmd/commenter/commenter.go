package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "os"
  "strconv"
  "strings"

  "github.com/mugioka/go-github-pr-commenter/commenter"
  "github.com/google/go-github/v32/github"
)

func main() {
  fmt.Println("Starting the github commenter...")

  token := os.Getenv("INPUT_GITHUB_TOKEN")
  if len(token) == 0 {
    fail("The INPUT_GITHUB_TOKEN has not been set")
  }

  githubRepository := os.Getenv("GITHUB_REPOSITORY")
  split := strings.Split(githubRepository, "/")
  if len(split) != 2 {
    fail(fmt.Sprintf("Expected value for split not found. Expected 2 in %v", split))
  }
  owner := split[0]
  repo := split[1]

  prNo, err := extractPullRequestNumber()
  if err != nil {
    fail(err.Error())
  }
  c, err := commenter.NewCommenter(token, owner, repo, prNo)
  if err != nil {
    fail(err.Error())
  }
  results, err := loadResultsFile()
  if err != nil {
    fail(err.Error())
  }
  var prReviewComments []commenter.PRReviewComment
  for _, result := range results {
    prReviewComment := generatePRReviewComment(result)
    prReviewComments = append(prReviewComments, prReviewComment)
  }
  draftPRReviewComments := c.CreateDraftPRReviewComments(prReviewComments)
  prReviewEvent := selectPRReviewEventBy(draftPRReviewComments)
  err = c.WritePRReview(draftPRReviewComments, prReviewEvent)
  if err != nil {
    fail(err.Error())
  } else {
    fmt.Printf("The PR review was written successfully.")
  }
}

func selectPRReviewEventBy(comments []*github.DraftReviewComment) string {
  if len(comments) > 0 {
    return commenter.RequestChanges
  } else {
    return commenter.Approve
  }
}

func generatePRReviewComment(result Result) commenter.PRReviewComment {
  fileName := strings.ReplaceAll(result.Range.Filename, fmt.Sprintf("%s/", os.Getenv("GITHUB_WORKSPACE")), "")
  body := generateErrorMessage(result)
  return commenter.PRReviewComment{
    FileName: fileName,
    StartLine: result.Range.StartLine,
    EndLine: result.Range.EndLine,
    Body: body,
  }
}

func generateErrorMessage(result Result) string {
  return fmt.Sprintf(
    "## result\ntfsec check %s failed.\n## severity\n⚠️%s\n## reason\n%s\n## how to ignore\n`#tfsec:ignore:%s`([refs](https://github.com/aquasecurity/tfsec#ignoring-warnings))\n\nFor more information, [see](%s)\n",
    result.RuleID,
    result.Severity,
    result.Description,
    result.LegacyRuleID,
    result.Links[0])
}

func extractPullRequestNumber() (int, error) {
  file, err := ioutil.ReadFile("/github/workflow/event.json")
  if err != nil {
    return -1, err
  }

  var data interface{}
  err = json.Unmarshal(file, &data)
  if err != nil {
    return -1, err
  }
  payload := data.(map[string]interface{})

  return strconv.Atoi(fmt.Sprintf("%v", payload["number"]))
}

func fail(err string) {
  fmt.Printf("The commenter failed with the following error:\n%s\n", err)
  os.Exit(-1)
}
