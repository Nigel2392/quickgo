package main

import (
	"fmt"
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
			panic(err)
		}
		defer f.Close()
		_, err = f.WriteString(file.Content)
		if err != nil {
			panic(err)
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
	name = AppConfig.GetName(name)
	return os.Remove(EXE_DIR + "\\conf\\" + name)
}

func InitProject(proj_name string, dir Directory) (Directory, error) {
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
	if proj_name == "" {
		proj_name = dir.Name
	}
	CreateProject(dir, proj_name)
	return dir, nil
}

func InitProjectConfig(path string) (Directory, error) {
	var dir Directory
	var err error

	file, err := os.ReadFile(path)
	if err != nil {
		return Directory{}, err
	}
	dir, err = AppConfig.Deserialize(file)
	if err != nil {
		return Directory{}, err
	}
	// write file to current directory
	fname := AppConfig.GetName(URLOmit(path))
	err = os.WriteFile(EXE_DIR+"\\conf\\"+fname, file, 0644)
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

func setupExclude(exclude []string) []string {
	if exclude == nil {
		return []string{}
	}
	for i, ex := range exclude {
		if strings.HasPrefix(ex, "./") {
			ex = strings.TrimPrefix(ex, ".")
		}
		ex = filepath.ToSlash(ex)
		ex = strings.TrimPrefix(ex, "/")
		exclude[i] = ex
	}
	return exclude
}

func anyFunc[T any](s []T, f func(int) bool) bool {
	if s == nil {
		return false
	}
	for i := range s {
		if f(i) {
			return true
		}
	}
	return false
}

func isExcluded(path string, exclude []string, excludeContains []string) bool {
	if exclude == nil && excludeContains == nil {
		return false
	}

	if anyFunc(excludeContains, func(i int) bool {
		return strings.Contains(path, excludeContains[i])
	}) {
		return true
	}
	if exclude != nil {
		if strings.HasPrefix(path, "./") {
			path = strings.TrimPrefix(path, ".")
		}
		path = filepath.ToSlash(path)
		path = strings.TrimPrefix(path, "/")
		if anyFunc(exclude, func(i int) bool {
			return strings.HasPrefix(path, exclude[i])
		}) {
			return true
		}
	}
	return false
}

func GetDirFromPath(path string, exclude []string, excludeContains []string, verbose bool) (Directory, error) {
	var dir Directory
	dir.Name = filepath.Base(path)
	files, err := os.ReadDir(path)
	if err != nil {
		return Directory{}, err
	}
	for _, f := range files {
		var fullPath = filepath.Join(path, f.Name())
		if isExcluded(fullPath, exclude, excludeContains) {
			if verbose {
				fmt.Printf("[excluded] %s\n", fullPath)
			}
			continue
		}
		if verbose {
			fmt.Printf("[reading] %s\n", fullPath)
		}
		if f.IsDir() {
			child, err := GetDirFromPath(fullPath, exclude, excludeContains, verbose)
			if err != nil {
				return Directory{}, err
			}
			dir.Children = append(dir.Children, child)
		} else {
			file, err := os.ReadFile(fullPath)
			if err != nil {
				return Directory{}, err
			}
			dir.Files = append(dir.Files, File{Name: f.Name(), Content: string(file)})
		}
	}
	return dir, nil
}
func GetDir(name string, project_name string, raw bool) (Directory, error) {
	name = AppConfig.GetName(name)
	file, err := os.ReadFile(EXE_DIR + "\\conf\\" + name)
	if err != nil {
		file, err = ConfFS.ReadFile("conf/" + name)
		if err != nil {
			return Directory{}, err
		}
	}
	dir, err := AppConfig.Deserialize(file)
	if !raw {
		dir = RenameDirData(dir, project_name)
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
		panic(err)
	}
	var filenames []string
	for _, f := range files {
		name := f.Name()
		name_ext := strings.Split(name, ".")
		if len(name) > 1 {
			name = name_ext[0]
			if AppConfig.IsEncType(name_ext[len(name_ext)-1]) {
				filenames = append(filenames, name)
			}
		}
	}
	return filenames
}

func ListInternalConfigs() []string {
	files, err := ConfFS.ReadDir("conf")
	if err != nil {
		panic(err)
	}
	var namelist []string
	for _, f := range files {
		name_ext := strings.Split(f.Name(), ".")
		if AppConfig.IsEncType(name_ext[len(name_ext)-1]) {
			namelist = append(namelist, name_ext[0])
		}
	}
	return namelist
}
