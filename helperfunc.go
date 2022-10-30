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
	"time"

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

func Loading(s string, l int) {
	fmt.Print(Craft(CMD_Bold, Craft(CMD_Red, s)))
	for i := 0; i < l; i++ {
		fmt.Print(Craft(CMD_Bold, Craft(CMD_Red, ".")))
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
}

func URLOmit(url string) string {
	var urlomitted = url
	if strings.Contains(url, "/") {
		s_url := strings.Split(url, "/")
		urlomitted = s_url[len(s_url)-1]
	}
	if strings.Contains(url, "\\") {
		s_url := strings.Split(url, "\\")
		urlomitted = s_url[len(s_url)-1]
	}
	urlomitted = strings.ReplaceAll(urlomitted, "-", "_")
	urlomitted = strings.ReplaceAll(urlomitted, " ", "_")
	urlomitted = strings.ReplaceAll(urlomitted, ".", "_")
	return urlomitted
}
