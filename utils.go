package banji

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

func setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		shouldExit = true
		if err := execute(); err != nil {
			log.Println(err)
		} else {
			os.Exit(0)
		}
	}()
}

func isWatchable(file string) bool {
	for _, pattern := range exclusionFiles {
		if match, _ := filepath.Match(pattern, filepath.Base(file)); match {
			return false
		}
	}
	return true
}

func isWatchableDir(dir string) bool {
	parts := strings.Split(dir, string(os.PathSeparator))
	for _, part := range parts {
		for _, pattern := range exclusionDirectories {
			if match, _ := filepath.Match(pattern, part); match {
				return false
			}
		}
	}
	return true
}

func scanMainFunc() string {

	startPath := "."
	var mainFilePath string

	regexPattern := regexp.MustCompile(`(?s)package main.*func main\(\)`)

	err := filepath.WalkDir(startPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(content), "banji.Run()") {
				return nil
			}
			if regexPattern.MatchString(string(content)) {
				mainFilePath = path
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return mainFilePath
}

func generateCommand() error {

	if MainFuncDirectory == "" {
		MainFuncDirectory = scanMainFunc()
		if MainFuncDirectory == "" {
			return fmt.Errorf("couldNotFindMainFunction")
		}
	}

	commands = append(commands, "run", MainFuncDirectory)
	commands = append(commands, Flags...)
	return nil
}
