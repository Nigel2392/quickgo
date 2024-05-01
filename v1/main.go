package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
)

const (
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
		panic(err)
	}
}

func main() {
	var exclude = arrayFlags(make([]string, 0))
	var excludeContains = arrayFlags(make([]string, 0))
	fr := FlagRunner{
		exclude:         &exclude,
		excludeContains: &excludeContains,
	}
	fr.importpath = flag.String("import", "", "Path of the JSON/GOB file to be imported")
	flag.Var(fr.exclude, "x", "Exclude files and directories from the project, if the path starts with this string it will be excluded")
	flag.Var(fr.excludeContains, "xcon", "Exclude files and directories from the project, if the path contains this string it will be excluded")
	fr.verbose = flag.Bool("verbose", false, "Verbose output")
	fr.get_config = flag.String("get", "", "Get the JSON config of the project")
	fr.config_name = flag.String("use", "", "Path of the JSON file to use for creating templates")
	fr.list_configs = flag.Bool("l", false, "List all the available configs")
	fr.proj_name = flag.String("n", "", "Name of the project to be created")
	fr.view_config = flag.Bool("v", false, "View the config of the project")
	fr.location = flag.Bool("loc", false, "Location of the executable")
	fr.del_conf = flag.Bool("del", false, "Delete a config")
	fr.raw = flag.Bool("raw", false, "Output raw project from json")
	fr.serve = flag.Bool("serve", false, "Serve the project files over http to preview (optional -o)")
	fr.host = flag.String("host", "", "Host to serve on")
	fr.port = flag.String("port", "", "Port to serve on")
	fr.openBrowser = flag.Bool("o", false, "Open the browser after serving the project")
	fr.encoder = flag.String("enc", "", "Encoder to use for the project (json/gob). Can also be set in the config.json")

	if len(os.Args) == 1 {
		PrintLogo()
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	flag.Parse()
	fr.Run()
}
