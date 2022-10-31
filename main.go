package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

var (
	CMD_Bold          = "\033[1m"
	CMD_Black         = "\033[30m"
	CMD_BRIGHT_Blue   = "\033[34;1m"
	CMD_Blue          = "\033[34m"
	CMD_BRIGHT_Cyan   = "\033[36;1m"
	CMD_Cyan          = "\033[36m"
	CMD_BRIGHT_Gray   = "\033[37;1m"
	CMD_Gray          = "\033[37m"
	CMD_BRIGHT_Green  = "\033[32;1m"
	CMD_Green         = "\033[32m"
	CMD_BRIGHT_Purple = "\033[35;1m"
	CMD_Purple        = "\033[35m"
	CMD_Underline     = "\033[4m"
	CMD_BRIGHT_Red    = "\033[31;1m"
	CMD_Red           = "\033[31m"
	CMD_Reset         = "\033[0m"
	CMD_White         = "\033[97m"
	CMD_BRIGHT_Yellow = "\033[33;1m"
	CMD_Yellow        = "\033[33m"
)

//go:embed conf/*
var ConfFS embed.FS

var WORKING_DIR, _ = os.Getwd()
var EXE_DIR = GetExeDIR()
var AppConfig *Configuration

func PrintLogo() {
	str := Craft(CMD_Cyan, " $$$$$$\\            $$\\           $$\\         "+Craft(CMD_Cyan, "     $$$$$$\\            \n")) +
		Craft(CMD_Cyan, "$$  __$$\\           \\__|          $$ |         "+Craft(CMD_Cyan, "   $$  __$$\\           \n")) +
		Craft(CMD_Blue, "$$ /  $$ |$$\\   $$\\ $$\\  $$$$$$$\\ $$ |  $$\\ "+Craft(CMD_Cyan, "      $$ /  \\__| $$$$$$\\  \n")) +
		Craft(CMD_Blue, "$$ |  $$ |$$ |  $$ |$$ |$$  _____|$$ | $$  |     "+Craft(CMD_Cyan, " $$ |$$$$\\ $$  __$$\\ \n")) +
		Craft(CMD_Blue, "$$ |  $$ |$$ |  $$ |$$ |$$ /      $$$$$$  /      "+Craft(CMD_Cyan, " $$ |\\_$$ |$$ /  $$ |\n")) +
		Craft(CMD_Purple, "$$ $$\\$$ |$$ |  $$ |$$ |$$ |      $$  _$$<      "+Craft(CMD_Cyan, "  $$ |  $$ |$$ |  $$ |\n")) +
		Craft(CMD_Purple, "\\$$$$$$ / \\$$$$$$  |$$ |\\$$$$$$$\\ $$ | \\$$\\   "+Craft(CMD_Cyan, "    \\$$$$$$  |\\$$$$$$  |\n")) +
		Craft(CMD_BRIGHT_Purple, " \\___"+CMD_Reset+Craft(CMD_Purple, "$$$")+Craft(CMD_BRIGHT_Purple, "\\  \\______/ \\__| \\_______|\\__|  \\__| ")+Craft(CMD_Cyan, "      \\______/  \\______/ \n")) +
		Craft(CMD_BRIGHT_Purple, "     \\___|                                         "+Craft(CMD_Cyan, "                   \n"))
	fmt.Println(str)
	fmt.Println(Craft(CMD_Red, "\nCreated by: "+Craft(CMD_Purple, "Nigel van Keulen")))
}

func init() {
	var err error
	// Get user configurations
	if _, err = os.Stat(EXE_DIR + "\\conf"); os.IsNotExist(err) {
		os.Mkdir(EXE_DIR+"\\conf", 0755)
	}
	// Get application configuration
	if AppConfig, err = AppConfig.GetConfig(EXE_DIR + "\\config.json"); err != nil {
		log.Fatal(err)
	}
}

func main() {

	importpath := flag.String("import", "", "Path of the JSON/GOB file to be imported")
	get_config := flag.String("get", "", "Get the JSON config of the project")
	config_name := flag.String("use", "", "Path of the JSON file to use for creating templates")
	list_configs := flag.Bool("l", false, "List all the available configs")
	proj_name := flag.String("n", "", "Name of the project to be created")
	view_config := flag.Bool("v", false, "View the config of the project")
	location := flag.Bool("loc", false, "Location of the executable")
	del_conf := flag.Bool("del", false, "Delete a config")
	RAW := flag.Bool("raw", false, "Output raw project from json")
	serve := flag.Bool("serve", false, "Serve the project files over http to preview (optional -o)")
	openBrowser := flag.Bool("o", false, "Open the browser after serving the project")
	encoder := flag.String("enc", "", "Encoder to use for the project (json/gob). Can also be set in the config.json")

	if len(os.Args) == 1 {
		PrintLogo()
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	flag.Parse()

	if *encoder != "" {
		switch strings.TrimSpace(strings.ToLower(*encoder)) {
		case "json":
			AppConfig.Encoder = "json"
		case "gob":
			AppConfig.Encoder = "gob"
		default:
			fmt.Println(Craft(CMD_Red, "Invalid encoder type"))
			os.Exit(1)
		}
	}

	if *importpath != "" {
		_, err := InitProjectConfig(*importpath)
		if err != nil {
			log.Fatal(err)
		}
	} else if *serve {
		dirnames := ListInternalConfigs()
		dirnames = append(dirnames, ListConfigs()...)
		viewer := NewViewer(dirnames, *RAW)
		if err := viewer.serve(*openBrowser); err != nil {
			log.Fatal(err)
		}
	} else if *config_name != "" {
		dir, err := GetDir(*config_name, *proj_name, *RAW)
		if err != nil {
			log.Fatal(err)
		}
		if *view_config {
			ListFiles(dir, "")
			return
		} else if *del_conf {
			err := DeleteConfig(*config_name)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		_, err = InitProject(*config_name, *proj_name, dir)
		if err != nil {
			log.Fatal(err)
		}
	} else if *list_configs {
		int_confs := ListInternalConfigs()
		ext_confs := ListConfigs()
		sort.Slice(int_confs, func(i, j int) bool {
			return strings.ToLower(int_confs[i]) < strings.ToLower(int_confs[j])
		})
		sort.Slice(ext_confs, func(i, j int) bool {
			return strings.ToLower(ext_confs[i]) < strings.ToLower(ext_confs[j])
		})
		for _, conf := range int_confs {
			fmt.Println(Craft(CMD_Purple, conf))
		}
		for _, conf := range ext_confs {
			fmt.Println(Craft(CMD_Blue, conf))
		}

	} else if *get_config != "" {
		dir, err := GetDirFromPath(*get_config)
		if err != nil {
			log.Fatal(err)
		}
		if *proj_name == "" {
			*proj_name = *get_config
		}
		*proj_name = URLOmit(*proj_name)
		if strings.EqualFold(*proj_name, "static") {
			*proj_name = strings.Replace(strings.ToLower(*proj_name), "static", "tpl_static", 1)
			fmt.Println(Craft(CMD_Red, "Warning: The project name contains 'static' which is reserved for static files when serving.\n The project name will be changed to: "+*proj_name))
		}
		err = AppConfig.Serialize(dir, EXE_DIR+"\\conf\\"+*proj_name)
		if err != nil {
			log.Fatal(err)
		}
	} else if *location {
		PrintLocation()
	} else {
		PrintLogo()
		flag.CommandLine.Usage()
		os.Exit(1)
	}
}
