package main

import (
  "fmt"
  "os"
  "flag"
  "path/filepath"
  "bufio"
  "logscanner/utils"
  "encoding/json"
  "strings"
)

type Match struct { 
  File string `json:"file"`
  Line int `json:"line"`
  Text string `json:"text"`
}

func main() {
  startDir, exts, keywords, outputFile := handleArguments()

  fmt.Printf("Searching files in '%s' with extention '%v' and keywords '%v'\n", startDir, exts, keywords)

  matches, err := scanFiles(startDir, exts, keywords)

  if err != nil {
    fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
    os.Exit(1)
  }

  jsonData, err := json.MarshalIndent(matches, "", " ")
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error serializing json %v\n", err)
    os.Exit(1)
  }

  if outputFile != "" {
    if !strings.HasSuffix(outputFile, ".json") {
			outputFile += ".json"
		}

		err := os.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Results written to %s\n", outputFile)
  } else {
    fmt.Println(string(jsonData))
  }
}

func handleArguments() (string, []string, []string, string) {
  startDir := flag.String("startDir", "/", "Starting directory where search")
  ext := flag.String("ext", ".log,.csv", "Extention to search")
  key := flag.String("keyword", "", "Keywords to search")
  outputFile := flag.String("output", "", "Path to save the result")
  flag.Parse()

  exts := utils.SplitAndTrim(*ext)
  keywords := utils.SplitAndTrim(*key)

  if len(exts) == 0 || len(keywords) == 0 {
    fmt.Fprintf(os.Stderr, "Error: you must specify both exts and keywords\n")
    os.Exit(1)
  }

  return *startDir, exts, keywords, *outputFile
}

func scanFiles(startDir string, exts, keywords []string) ([]Match, error) {
  var matches []Match

  err := filepath.WalkDir(startDir, func(path string, d os.DirEntry, err error) error {
    if err != nil {
      return err
    }

    if d.IsDir() {
      return nil
    }

    if !utils.HasValidExtention(d.Name(), exts) {
      return nil
    }

    file, err := os.Open(path)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error opening the file in the path '%s': %v\n ", path, err)
      return nil
    }

    defer file.Close()

    scanner := bufio.NewScanner(file)
    lineNum := 1
    for scanner.Scan() {
      line := scanner.Text()
      if utils.ContainsKeyword(line, keywords) {
        matches = append(matches, Match{
          File: path,
          Line: lineNum,
          Text: line,
        })
      }
      lineNum++
    }

    if err := scanner.Err(); err != nil {
      fmt.Fprintf(os.Stderr, "Errore durante la lettura di '%s': %v\n", path, err)
    }

    return nil
  })

  return matches, err
}

