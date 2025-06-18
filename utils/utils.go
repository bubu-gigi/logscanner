package utils

import "strings"

func SplitAndTrim(input string) []string {
  parts := strings.Split(input, ",")
  var result []string
  for _, p := range parts {
    trimmed := strings.TrimSpace(p)
    if trimmed != "" {
      result = append(result, trimmed)
    }
  }
  return result
}

func HasValidExtention(filename string, exts []string) bool {
  for _, ext := range exts {
    if strings.HasSuffix(filename, ext) {
      return true
    }
  }
  return false
}

func ContainsKeyword(line string, keywords []string) bool {
  lower := strings.ToLower(line)
  for _, keyword := range keywords {
    if strings.Contains(lower, strings.ToLower(keyword)) {
      return true
    }
  }
  return false
}



