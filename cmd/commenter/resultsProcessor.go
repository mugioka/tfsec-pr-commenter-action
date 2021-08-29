package main

import (
  "encoding/json"
  "io/ioutil"
)

type CheckRange struct {
  Filename  string `json:"filename"`
  StartLine int    `json:"start_line"`
  EndLine   int    `json:"end_line"`
}

type Result struct {
  RuleID          string      `json:"rule_id"`
  LegacyRuleID    string      `json:"legacy_rule_id"`
  RuleDescription string      `json:"rule_description"`
  RuleProvider    string      `json:"rule_provider"`
  Links           []string    `json:"links"`
  Range           *CheckRange `json:"location"`
  Description     string      `json:"description"`
  RangeAnnotation string      `json:"-"`
  Severity        string      `json:"severity"`
}

const resultsFile = "results.json"

func loadResultsFile() ([]Result, error) {
  results := []Result{}

  file, err := ioutil.ReadFile(resultsFile)
  if err != nil {
    return nil, err
  }
  err = json.Unmarshal(file, &results)
  if err != nil {
    return nil, err
  }
  return results, nil
}
