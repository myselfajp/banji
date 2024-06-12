package banji

import (
	"log"
	"os/exec"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	exclusionFiles       = []string{".git*"}
	exclusionDirectories = []string{"*.git*"}
	cmd                  *exec.Cmd
	mu                   sync.Mutex
	shouldExit           bool
	MainFuncDirectory    string
	Flags                []string
	commands             []string
)

func Run() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	setupSignalHandler()

	if err = generateCommand(); err != nil {
		println(err.Error())
		return
	}

	go watchFileSystem(watcher)

	go addWatchers(watcher)

	if err := execute(); err != nil {
		println(err)
		return
	}

	select {}
}
