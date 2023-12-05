package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var rootPath string

func clearScreen() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin": // Linux and macOS
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		// Unsupported OS
		return
	}
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return
	}
}

func init() {
	// Command-line flags
	flag.StringVar(&rootPath, "path", "/Volumes/Austin Disk/", "Root path to search for Git repositories")
}
func main() {
	// Script Information
	fmt.Println("Running GitScript 1.0")
	time.Sleep(2 * time.Second)
	fmt.Println("Copyright (C) 2023 Abdon Morales")
	time.Sleep(2 * time.Second)
	fmt.Println("License: Free Research License")
	time.Sleep(2 * time.Second)

	clearScreen()

	flag.Parse()
	// Verify the rootPath is provided
	if rootPath == "" {
		log.Fatal("Please specify a root path using the -path flag")
	}
	initLogging()

	var wg sync.WaitGroup

	// Walk the directory tree

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				log.Printf("Skipping %s: %v\n", path, err)
				return filepath.SkipDir
			}
			return err
		}
		if isJunkFile(info.Name()) {
			return nil
		}
		if info.IsDir() && isGitRepo(path) {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				gitFetch(p)
			}(path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", rootPath, err)
	}
	// Wait for all goroutines to complete
	wg.Wait()
}

// isJunkFile checks if a file is a junk file
func isJunkFile(fileName string) bool {
	// List of junk files to ignore
	junkFiles := []string{".DS_Store", "._DS_Store"}
	for _, junk := range junkFiles {
		if strings.EqualFold(fileName, junk) {
			return true
		}
	}
	return false
}

// isGitRepo checks if a given path is a Git repository
func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return !os.IsNotExist(err)
}

// gitFetch performs a git fetch in the given directory and handles errors
func gitFetch(repoPath string) {
	var stderr bytes.Buffer
	cmd := exec.Command("git", "fetch")
	cmd.Dir = repoPath
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("Error fetching in repository %s: %v\n", repoPath, err)
		log.Printf("Git error: %s\n", stderr.String())
	} else {
		log.Printf("Successfully fetched in repository: %s\n", repoPath)
	}
}

func initLogging() {
	file, err := os.OpenFile("gitscript.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
}
