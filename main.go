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
  "sync"
)

type Match struct { 
  File string `json:"file"`
  Line int `json:"line"`
  Text string `json:"text"`
}

func main() {
  startDir, exts, keywords, outputFile, workers := handleArguments()

  fmt.Printf("Searching files in '%s' with extention '%v' and keywords '%v'\n", startDir, exts, keywords)

  matches, err := scanFiles(startDir, exts, keywords, workers)

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

func handleArguments() (string, []string, []string, string, int) {
  startDir := flag.String("startDir", "/", "Starting directory where search")
  ext := flag.String("ext", ".log,.csv", "Extention to search")
  key := flag.String("keyword", "", "Keywords to search")
  outputFile := flag.String("output", "", "Path to save the result")
  workers := flag.Int("workers", 8, "Workers to process files")
  flag.Parse()

  exts := utils.SplitAndTrim(*ext)
  keywords := utils.SplitAndTrim(*key)
  if len(exts) == 0 || len(keywords) == 0 {
    fmt.Fprintf(os.Stderr, "Error: you must specify both exts and keywords\n")
    os.Exit(1)
  }

  return *startDir, exts, keywords, *outputFile, *workers
}

func scanFiles(startDir string, exts, keywords []string, numWorkers int) ([]Match, error) {
  var matches []Match
  var matchesMu sync.Mutex

  fileChan := make(chan string, 100)
  wg := sync.WaitGroup{}

  for i := 0; i<numWorkers; i++ {
    wg.Add(1)
    workerId := i
    go func() {
      defer wg.Done()
      for path := range fileChan {
        fmt.Printf("[Worker %d] Processing file: %s\n", workerId, path)
        processFile(path, keywords, &matches, &matchesMu)
      }
    }()
  }
  err := filepath.WalkDir(startDir, func(path string, d os.DirEntry, err error) error {
    if err != nil {
      return err
    }
    if d.IsDir() || !utils.HasValidExtention(d.Name(), exts) {
      return nil
    }
    fileChan <- path
    return nil
  })

  close(fileChan)
  wg.Wait()

  return matches, err
}

func processFile(path string, keywords []string, matches *[]Match, mu *sync.Mutex) {
  file, err := os.Open(path)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening the file %s:%v\n", path, err)
    return
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  lineNum := 1
  for scanner.Scan() {
    line := scanner.Text()
    if utils.ContainsKeyword(line, keywords) {
      mu.Lock()
      *matches = append(*matches, Match{
        File: path,
        Line: lineNum,
        Text: line,
      })
      mu.Unlock()
    }
    lineNum++
  }

  if err := scanner.Err(); err != nil {
    fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
  }
}

