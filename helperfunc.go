package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Nigel2392/typeutils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

func GetExeDIR() string {
	exe_file, err := os.Executable()
	if err != nil {
		panic(err)
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
	var urlomitted = strings.ReplaceAll(url, "\\", "/")
	if strings.Contains(urlomitted, "/") {
		s_url := strings.Split(urlomitted, "/")
		urlomitted = s_url[len(s_url)-1]
	}
	urlomitted = strings.ReplaceAll(urlomitted, "-", "_")
	urlomitted = strings.ReplaceAll(urlomitted, " ", "_")
	urlomitted = strings.ReplaceAll(urlomitted, ".", "_")
	return urlomitted
}
func sizeStr[T int | int16 | int32 | int64 | float32 | float64](size T) string {
	f_size := float64(size)
	if f_size < 1024 {
		return fmt.Sprintf("%d b", int(f_size))
	}
	f_size = f_size / 1024
	if f_size < 1024 {
		return fmt.Sprintf("%.1f KB", f_size)
	}
	f_size = f_size / 1024
	if f_size < 1024 {
		return fmt.Sprintf("%.1f MB", f_size)
	}
	f_size = f_size / 1024
	return fmt.Sprintf("%.1f GB", f_size)
}

func WriteConf(path string, data []byte) error {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the file
		if err = os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	} else {
		// Delete the file
		answer := RepeatAsk("The file already exists, do you want to overwrite it? (y/n): ", []string{"y", "n"})
		if answer == "y" {
			return os.WriteFile(path, data, 0644)
		} else if answer == "n" {
			answer = RepeatAsk("Do you want to change the name of the file? (y/n): ", []string{"y", "n"})
			if answer == "y" {
				// Ask for the new name
				name := typeutils.Ask("Enter the name of the file: ")
				// Make the name path safe
				name = URLOmit(AppConfig.GetName(name))
				// Write the file by recursion
				return WriteConf(EXE_DIR+"\\conf\\"+name, data)
			} else if answer == "n" {
				return fmt.Errorf("operation canceled by the user")
			}
		}
	}
	return nil
}

func GetDepth(dirs []Directory) int {
	depth := 1
	for _, dir := range dirs {
		if len(dir.Children) > 0 {
			depth = int(math.Max(float64(depth), float64(GetDepth(dir.Children))))
		}
	}
	return depth + 1
}

func ReplaceNames(data []byte, name string) []byte {
	name_urlomitted := URLOmit(name)
	data = bytes.Replace(data, []byte("$$PROJECT_NAME$$"), []byte(name), -1)
	var re = regexp.MustCompile(`\$\$PROJECT_NAME\s*;\s*OMITURL\$\$`)
	data = re.ReplaceAll(data, []byte(name_urlomitted))
	re = regexp.MustCompile(`\$\$PROJECT_NAME\s*;\s*URLOMIT\$\$`)
	data = re.ReplaceAll(data, []byte(name_urlomitted))
	return data
}

func ReplaceNamesString(data string, name string) string {
	return string(ReplaceNames([]byte(data), name))
}

func Markdownify(data []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.DefinitionList, extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(data, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
