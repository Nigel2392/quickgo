package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Nigel2392/typeutils"
)

func PrintLocation() {
	fmt.Println(Craft(CMD_Bold, Craft(CMD_Blue, "Current location: "+WORKING_DIR)))
	fmt.Println(Craft(CMD_Bold, Craft(CMD_Blue, "Executable location: "+EXE_DIR)))
}

func InitProjectConfig(path string) (Directory, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return Directory{}, err
	}
	dir, err := FileToDir(file)
	if err != nil {
		return Directory{}, err
	}
	// write file to current directory
	err = os.WriteFile(EXE_DIR+"\\conf\\", file, 0644)
	if err != nil {
		return Directory{}, err
	}
	// Delete the file
	err = os.Remove(path)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
}

func Loading(s string, l int) {
	fmt.Print(Craft(CMD_Bold, Craft(CMD_Red, s)))
	for i := 0; i < l; i++ {
		fmt.Print(Craft(CMD_Bold, Craft(CMD_Red, ".")))
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
}

func GetDir(name string, project_name string) (Directory, error) {
	name = name + ".json"
	file, err := os.ReadFile(EXE_DIR + "\\conf\\" + name)
	if err != nil {
		file, err = ConfFS.ReadFile("conf/" + name)
		if err != nil {
			return Directory{}, err
		}
	}
	if !*RAW {
		if project_name == "" {
			project_name = name
		}
		var project_name_urlomitted = project_name
		if strings.Contains(project_name, "/") {
			proj_vars := strings.Split(project_name, "/")
			project_name_urlomitted = proj_vars[len(proj_vars)-1]
		}
		if strings.Contains(project_name, "\\") {
			proj_vars := strings.Split(project_name, "\\")
			project_name_urlomitted = proj_vars[len(proj_vars)-1]
		}
		project_name_urlomitted = strings.ReplaceAll(project_name_urlomitted, "-", "_")
		file = bytes.Replace(file, []byte("$$PROJECT_NAME$$"), []byte(project_name), -1)
		var re = regexp.MustCompile(`\$\$PROJECT_NAME\s*;\s*OMITURL\$\$`)
		file = re.ReplaceAll(file, []byte(project_name_urlomitted))
	}

	var dir Directory
	err = json.Unmarshal(file, &dir)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
}

func InitProject(name string, proj_name string, dir Directory) (Directory, error) {
	Loading("Creating project from "+dir.Name, 3)
	if proj_name == "" {
		proj_name = dir.Name
	}
	if strings.Contains(proj_name, "/") {
		proj_vars := strings.Split(proj_name, "/")
		proj_name = proj_vars[len(proj_vars)-1]
	}
	// write file to current directory
	path, err := os.Getwd()
	if err != nil {
		return Directory{}, err
	}
	os.Mkdir(path+"\\"+proj_name, 0755)
	// os.Chdir(path + "\\" + proj_name)
	CreateProject(dir, proj_name)
	return dir, nil
}

func CreateProject(dir Directory, name string) {
	// create the directory
	os.Mkdir(name, 0755)
	// create the files
	for _, file := range dir.Files {
		f, err := os.Create(name + "\\" + file.Name)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err = f.WriteString(file.Content)
		if err != nil {
			log.Fatal(err)
		}
	}
	os.Chdir(".\\" + name)
	// create the children directories
	for _, child := range dir.Children {
		CreateProject(child, child.Name)
	}
	os.Chdir("..")
}

func ListFiles(dir Directory, indent string) {
	for _, file := range dir.Files {
		fmt.Println(indent + Craft(CMD_Green, file.Name))
	}
	for _, child := range dir.Children {
		fmt.Println(indent + Craft(CMD_Blue, child.Name))
		ListFiles(child, indent+"  ")
	}
}

func ListConfigs() []string {
	files, err := os.ReadDir(EXE_DIR + "\\conf\\")
	if err != nil {
		log.Fatal(err)
	}
	var filenames []string
	for _, f := range files {
		name := f.Name()
		name_ext := strings.Split(name, ".")
		if len(name) > 1 {
			name = name_ext[0]
			filenames = append(filenames, name)
		}
	}
	return filenames
}

func DeleteConfig(name string) error {
	name = name + ".json"
	return os.Remove(EXE_DIR + "\\conf\\" + name)
}

func GetConfFromDir(path string) (Directory, error) {
	var dir Directory
	dir.Name = filepath.Base(path)
	files, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}
	for _, f := range files {
		if f.IsDir() {
			child, err := GetConfFromDir(path + "\\" + f.Name())
			if err != nil {
				return Directory{}, err
			}
			dir.Children = append(dir.Children, child)
		} else {
			file, err := os.ReadFile(path + "\\" + f.Name())
			if err != nil {
				return Directory{}, err
			}
			file_data := string(file)
			dir.Files = append(dir.Files, File{Name: f.Name(), Content: file_data})
		}
	}
	return dir, nil
}

func WriteJSONConfig(dir Directory, path string) error {
	json_data, err := json.MarshalIndent(dir, "", "  ")
	if err != nil {
		return err
	}
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the file
		err = os.WriteFile(path, json_data, 0644)
		if err != nil {
			return err
		}
	} else {
		// Delete the file
		answer := RepeatAsk("The file already exists, do you want to overwrite it? (y/n): ", []string{"y", "n"})
		if answer == "y" {
			err := os.WriteFile(path, json_data, 0644)
			if err != nil {
				return err
			}
		} else if answer == "n" {
			answer = RepeatAsk("Do you want to change the name of the file? (y/n): ", []string{"y", "n"})
			if answer == "y" {
				name := typeutils.Ask("Enter the name of the file: ")
				err = os.WriteFile(EXE_DIR+"\\conf\\"+name+"json", json_data, 0644)
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

func InitLocalProject(conf_name string, project_name string) {
	conf, err := ConfFS.ReadFile("conf/" + conf_name + ".json")
	if err != nil {
		log.Fatal(err)
	}
	dir, err := FileToDir(conf)
	if err != nil {
		log.Fatal(err)
	}
	if project_name == "" {
		project_name = conf_name
	}
	_, err = InitProject(conf_name, project_name, dir)
	if err != nil {
		log.Fatal(err)
	}
}

func ListInternalConfigs() []string {
	files, err := ConfFS.ReadDir("conf")
	if err != nil {
		log.Fatal(err)
	}
	var namelist []string
	for _, f := range files {
		fname := strings.Split(f.Name(), ".")
		namelist = append(namelist, fname[0])
	}
	return namelist
}
