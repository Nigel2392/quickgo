package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nigel2392/typeutils"
)

var (
	CMD_Bold      = "\033[1m"
	CMD_Black     = "\033[30m"
	CMD_Blue      = "\033[34m"
	CMD_Cyan      = "\033[36m"
	CMD_Gray      = "\033[37m"
	CMD_Green     = "\033[32m"
	CMD_Purple    = "\033[35m"
	CMD_Underline = "\033[4m"
	CMD_Red       = "\033[31m"
	CMD_Reset     = "\033[0m"
	CMD_White     = "\033[97m"
	CMD_Yellow    = "\033[33m"
)

func GetExeDIR() string {
	exe_file, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exe_dir := filepath.Dir(exe_file)
	return exe_dir
}

var WORKING_DIR, _ = os.Getwd()
var EXE_DIR = GetExeDIR()

func PrintLocation() {
	fmt.Println(Craft(CMD_Bold, Craft(CMD_Blue, "Current location: "+WORKING_DIR)))
	fmt.Println(Craft(CMD_Bold, Craft(CMD_Blue, "Executable location: "+EXE_DIR)))
}

func Craft(color, s any) string {
	return fmt.Sprintf("%s%v%s", color, s, CMD_Reset)
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Directory struct {
	Name     string      `json:"name"`
	Children []Directory `json:"directory"`
	Files    []File      `json:"files"`
}

func ImportJSONConfig(path string) (Directory, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return Directory{}, err
	}
	json_data := string(file)
	var dir Directory
	err = json.Unmarshal([]byte(json_data), &dir)
	if err != nil {
		return Directory{}, err
	}
	// write file to current directory
	err = ioutil.WriteFile(EXE_DIR+"\\conf\\", file, 0644)
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

func GetDir(name string) (Directory, error) {
	name = name + ".json"
	file, err := ioutil.ReadFile(EXE_DIR + "\\conf\\" + name)
	if err != nil {
		return Directory{}, err
	}
	json_data := string(file)
	var dir Directory
	err = json.Unmarshal([]byte(json_data), &dir)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
}

func ImportJSON(name string, proj_name string, dir_to_use string, dir Directory) (Directory, error) {
	Loading("Creating project from "+dir.Name, 3)
	if proj_name == "" {
		proj_name = name
	}
	// write file to current directory
	path, err := os.Getwd()
	if err != nil {
		return Directory{}, err
	}
	if dir_to_use != "" {
		path = dir_to_use
	}
	os.Mkdir(path+"\\"+proj_name, 0755)
	os.Chdir(path + "\\" + proj_name)
	CreateProject(dir, dir.Name)
	return dir, nil
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

func ListConfigs() {
	files, err := ioutil.ReadDir(EXE_DIR + "\\conf\\")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		name := f.Name()
		name_ext := strings.Split(name, ".")
		if len(name) > 1 {
			name = name_ext[0]
		}
		fmt.Println(Craft(CMD_Blue, name))
	}
}

func DeleteConfig(name string) error {
	name = name + ".json"
	return os.Remove(EXE_DIR + "\\conf\\" + name)
}

func CreateProject(dir Directory, name string) {
	// create the directory
	os.Mkdir(name, 0755)
	// create the files
	for _, file := range dir.Files {
		f, err := os.Create(dir.Name + "\\" + file.Name)
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

func GetConfFromDir(path string) (Directory, error) {
	var dir Directory
	dir.Name = filepath.Base(path)
	files, err := ioutil.ReadDir(path)
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
			file, err := ioutil.ReadFile(path + "\\" + f.Name())
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
		err = ioutil.WriteFile(path, json_data, 0644)
		if err != nil {
			return err
		}
	} else {
		// Delete the file
		answer := RepeatAsk("The file already exists, do you want to overwrite it? (y/n): ", []string{"y", "n"})
		if answer == "y" {
			err := ioutil.WriteFile(path, json_data, 0644)
			if err != nil {
				return err
			}
		} else if answer == "n" {
			answer = RepeatAsk("Do you want to change the name of the file? (y/n): ", []string{"y", "n"})
			if answer == "y" {
				name := typeutils.Ask("Enter the name of the file: ")
				err = ioutil.WriteFile(EXE_DIR+"\\conf\\"+name+"json", json_data, 0644)
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

func PrintLogo() {
	str := Craft(CMD_Cyan, " $$$$$$\\            $$\\           $$\\         "+Craft(CMD_Cyan, "     $$$$$$\\            \n")) +
		Craft(CMD_Cyan, "$$  __$$\\           \\__|          $$ |         "+Craft(CMD_Cyan, "   $$  __$$\\           \n")) +
		Craft(CMD_Purple, "$$ /  $$ |$$\\   $$\\ $$\\  $$$$$$$\\ $$ |  $$\\ "+Craft(CMD_Cyan, "      $$ /  \\__| $$$$$$\\  \n")) +
		Craft(CMD_Purple, "$$ |  $$ |$$ |  $$ |$$ |$$  _____|$$ | $$  |     "+Craft(CMD_Cyan, " $$ |$$$$\\ $$  __$$\\ \n")) +
		Craft(CMD_Red, "$$ |  $$ |$$ |  $$ |$$ |$$ /      $$$$$$  /      "+Craft(CMD_Cyan, " $$ |\\_$$ |$$ /  $$ |\n")) +
		Craft(CMD_Red, "$$ $$\\$$ |$$ |  $$ |$$ |$$ |      $$  _$$<      "+Craft(CMD_Cyan, "  $$ |  $$ |$$ |  $$ |\n")) +
		Craft(CMD_Red, "\\$$$$$$ / \\$$$$$$  |$$ |\\$$$$$$$\\ $$ | \\$$\\   "+Craft(CMD_Cyan, "    \\$$$$$$  |\\$$$$$$  |\n")) +
		Craft(CMD_Purple, " \\___"+Craft(CMD_Red, "$$$")+Craft(CMD_Purple, "\\  \\______/ \\__| \\_______|\\__|  \\__| ")+Craft(CMD_Cyan, "      \\______/  \\______/ \n")) +
		Craft(CMD_Purple, "     \\___|                                         "+Craft(CMD_Cyan, "                   \n"))
	fmt.Println(str)
	fmt.Println(Craft(CMD_Red, "\nCreated by: "+Craft(CMD_Purple, "Nigel van Keulen")))
}

func init() {
	if _, err := os.Stat(EXE_DIR + "\\conf"); os.IsNotExist(err) {
		os.Mkdir(EXE_DIR+"\\conf", 0755)
	}
}

func main() {
	importpath := flag.String("import", "", "Path of the JSON file to be imported")
	get_config := flag.String("get", "", "Get the JSON config of the project")
	use_dir := flag.String("dir", "", "The directory to create the project in")
	config_name := flag.String("use", "", "Path of the JSON file to use for creating templates")
	list_configs := flag.Bool("l", false, "List all the available configs")
	proj_name := flag.String("n", "", "Name of the project to be created")
	view_config := flag.Bool("v", false, "View the config of the project")
	location := flag.Bool("loc", false, "Location of the executable")
	delete := flag.Bool("del", false, "Delete a config")

	if len(os.Args) == 1 {
		PrintLogo()
		flag.CommandLine.Usage()
		Loading("Exiting", 3)
	}

	flag.Parse()
	if *importpath != "" {
		_, err := ImportJSONConfig(*importpath)
		if err != nil {
			log.Fatal(err)
		}
	} else if *config_name != "" {
		dir, err := GetDir(*config_name)
		if err != nil {
			log.Fatal(err)
		}
		if *view_config {
			ListFiles(dir, "")
			return
		}
		if *delete {
			err := DeleteConfig(*config_name)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		_, err = ImportJSON(*config_name, *proj_name, *use_dir, dir)
		if err != nil {
			log.Fatal(err)
		}

	} else if *list_configs {
		ListConfigs()
	} else if *get_config != "" {
		dir, err := GetConfFromDir(*get_config)
		if err != nil {
			log.Fatal(err)
		}
		err = WriteJSONConfig(dir, EXE_DIR+"\\conf\\"+*get_config+".json")
		if err != nil {
			log.Fatal(err)
		}
	} else if *location {
		PrintLocation()
	}
}
