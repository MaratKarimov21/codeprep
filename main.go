package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	dir     = flag.String("dir", ".", "Directory to process")
	include = flag.String("include", "", "Comma-separated list of include patterns")
	exclude = flag.String("exclude", "", "Comma-separated list of exclude patterns")
	output  = flag.String("output", "context.txt", "Output file name")
)

func main() {
	flag.Parse()

	includePatterns := splitPatterns(*include)
	excludePatterns := splitPatterns(*exclude)

	var files []string
	allEntries := make(map[string]bool)

	err := filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(*dir, path)
		if err != nil {
			return err
		}

		if !isIncluded(relPath, includePatterns, excludePatterns) {
			return nil
		}

		files = append(files, relPath)
		dirPath := filepath.Dir(relPath)
		components := strings.Split(dirPath, string(filepath.Separator))
		current := ""
		for _, comp := range components {
			if comp == "." {
				continue
			}
			current = filepath.Join(current, comp)
			allEntries[current] = true
		}
		allEntries[relPath] = false

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No matching files found")
		os.Exit(0)
	}

	var paths []string
	for path := range allEntries {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var treeBuilder strings.Builder
	for _, path := range paths {
		depth := strings.Count(path, string(filepath.Separator))
		indent := strings.Repeat("    ", depth)
		name := filepath.Base(path)
		if allEntries[path] {
			treeBuilder.WriteString(fmt.Sprintf("%s%s/\n", indent, name))
		} else {
			treeBuilder.WriteString(fmt.Sprintf("%s%s\n", indent, name))
		}
	}

	var contentBuilder strings.Builder
	for _, file := range files {
		content, err := ioutil.ReadFile(filepath.Join(*dir, file))
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file, err)
			continue
		}
		contentBuilder.WriteString(fmt.Sprintf("\n=== File: %s ===\n", file))
		contentBuilder.Write(content)
		contentBuilder.WriteString("\n")
	}

	outputContent := treeBuilder.String() + "\n" + contentBuilder.String()
	outputPath := filepath.Join(*dir, *output)
	err = ioutil.WriteFile(outputPath, []byte(outputContent), 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	absOutputPath, _ := filepath.Abs(outputPath)
	fmt.Printf("Output file created at: %s\n", absOutputPath)
}

func splitPatterns(s string) []string {
	if s == "" {
		return nil
	}
	patterns := strings.Split(s, ",")
	for i, p := range patterns {
		patterns[i] = strings.TrimSpace(p)
	}
	return patterns
}

func isIncluded(relPath string, include, exclude []string) bool {
	fileName := filepath.Base(relPath)

	// Check exclude patterns
	for _, pattern := range exclude {
		if matches(pattern, relPath, fileName) {
			return false
		}
	}

	// Include all if no include patterns
	if len(include) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range include {
		if matches(pattern, relPath, fileName) {
			return true
		}
	}

	return false
}

func matches(pattern, fullPath, fileName string) bool {
	if strings.Contains(pattern, "/") {
		match, _ := filepath.Match(pattern, fullPath)
		return match
	}
	match, _ := filepath.Match(pattern, fileName)
	return match
}
