package banji

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func watchFileSystem(watcher *fsnotify.Watcher) {

	var debounceTimer *time.Timer

	executeDebounced := func() {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(3*time.Second, func() {

			if err := execute(); err != nil {
				log.Println("error starting server:", err)
			}

		})
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if isWatchable(event.Name) {
				executeDebounced()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func addWatchers(watcher *fsnotify.Watcher) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && isWatchableDir(path) {
			log.Println("Watching directory:", path)
			return watcher.Add(path)
		}
		return nil
	})
}

func execute() error {
	mu.Lock()
	defer mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT); err != nil {
			return err
		}
		cmd.Process.Wait()
	}

	if shouldExit {
		return nil
	}

	cmd = exec.Command("go", commands...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cmd = nil
		return err
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			cmd = nil
		}
	}()
	return nil
}
