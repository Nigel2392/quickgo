package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Nigel2392/typeutils"
)

func GetExeDIR() string {
	exe_file, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exe_dir := filepath.Dir(exe_file)
	return exe_dir
}

func Craft(color, s any) string {
	return fmt.Sprintf("%s%v%s", color, s, CMD_Reset)
}

func RepeatAsk(s string, allowed []string) string {
	answer := typeutils.Ask(s)
	answer = strings.ToLower(answer)
	for _, a := range allowed {
		if answer == strings.ToLower(a) {
			return answer
		}
	}
	return RepeatAsk(s, allowed)
}

func OpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func SortDirs(dirs []Directory) []Directory {
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name < dirs[j].Name
	})
	return dirs
}

func SortFiles(files []File) []File {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	return files
}
