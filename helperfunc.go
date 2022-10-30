package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
func sizeStr[T int | int16 | int32 | int64](size T) string {
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
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			return err
		}
	} else {
		// Delete the file
		answer := RepeatAsk("The file already exists, do you want to overwrite it? (y/n): ", []string{"y", "n"})
		if answer == "y" {
			err := os.WriteFile(path, data, 0644)
			if err != nil {
				return err
			}
		} else if answer == "n" {
			answer = RepeatAsk("Do you want to change the name of the file? (y/n): ", []string{"y", "n"})
			if answer == "y" {
				name := typeutils.Ask("Enter the name of the file: ")
				if strings.ToLower(AppConfig.Encoder) == "json" {
					name = name + ".json"
				} else if strings.ToLower(AppConfig.Encoder) == "gob" {
					name = name + ".gob"
				}
				err = os.WriteFile(EXE_DIR+"\\conf\\"+name, data, 0644)
				if err != nil {
					return err
				}
			} else if answer == "n" {
				os.Exit(1)
			}
		}
	}
	return nil
}

func gobEncode(dir Directory) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(dir)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(data []byte) (Directory, error) {
	var dir Directory
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&dir)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
}

func ReplaceNames(data []byte, name string) []byte {
	name_urlomitted := URLOmit(name)
	data = bytes.Replace(data, []byte("$$PROJECT_NAME$$"), []byte(name), -1)
	var re = regexp.MustCompile(`\$\$PROJECT_NAME\s*;\s*OMITURL\$\$`)
	data = re.ReplaceAll(data, []byte(name_urlomitted))
	return data
}

func ReplaceNamesString(data string, name string) string {
	return string(ReplaceNames([]byte(data), name))
}
