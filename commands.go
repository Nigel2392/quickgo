package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func PrintLocation() {
	fmt.Println(Craft(CMD_Bold, Craft(CMD_Blue, "Current location: "+WORKING_DIR)))
	fmt.Println(Craft(CMD_Bold, Craft(CMD_Blue, "Executable location: "+EXE_DIR)))
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

func DeleteConfig(name string) error {
	if strings.ToLower(AppConfig.Encoder) == "json" {
		name = name + ".json"
	} else if strings.ToLower(AppConfig.Encoder) == "gob" {
		name = name + ".gob"
	}
	return os.Remove(EXE_DIR + "\\conf\\" + name)
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
func GetDir(name string, project_name string) (Directory, error) {
	if strings.ToLower(AppConfig.Encoder) == "json" {
		name = name + ".json"
	} else if strings.ToLower(AppConfig.Encoder) == "gob" {
		name = name + ".gob"
	}
	file, err := os.ReadFile(EXE_DIR + "\\conf\\" + name)
	if err != nil {
		file, err = ConfFS.ReadFile("conf/" + name)
		if err != nil {
			return Directory{}, err
		}
	}
	if !*RAW && strings.ToLower(AppConfig.Encoder) == "json" {
		if project_name == "" {
			project_name = name
		}
		file = ReplaceNames(file, project_name)
	}

	dir, err := DeSerializeDir(file)
	if strings.ToLower(AppConfig.Encoder) == "gob" {
		dir := RenameDirData(dir, project_name)
		return dir, nil
	}
	return dir, err
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

func WriteConfig(dir Directory, path string) error {
	if strings.ToLower(AppConfig.Encoder) == "json" {
		return WriteJSONConfig(dir, path+".json")
	} else if strings.ToLower(AppConfig.Encoder) == "gob" {
		return WriteGOBConfig(dir, path+".gob")
	}
	return fmt.Errorf("invalid encoder")

}
func WriteJSONConfig(dir Directory, path string) error {
	json_data, err := json.MarshalIndent(dir, "", "  ")
	if err != nil {
		return err
	}
	err = WriteConf(path, json_data)
	if err != nil {
		return err
	}
	return nil
}

func WriteGOBConfig(dir Directory, path string) error {
	gob_data, err := gobEncode(dir)
	if err != nil {
		return err
	}
	err = WriteConf(path, gob_data)
	if err != nil {
		return err
	}
	return nil
}
