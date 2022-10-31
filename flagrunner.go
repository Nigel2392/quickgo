package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

type FlagRunner struct {
	importpath   *string // JSON/GOB directory config to import
	get_config   *string // Generate configuration from directory
	config_name  *string // Configuration name to use (-l to view available)
	proj_name    *string // Project name to use, or name of new config when using -get
	encoder      *string // Encoder type for new config when generating, or using
	list_configs *bool   // List available configurations
	view_config  *bool   // View configuration file/directory tree
	location     *bool   // Print current directory location, and executable location
	del_conf     *bool   // Delete configuration, only works when specifying a config with -use
	raw          *bool   // Don't replace $$PROJECT_NAME$$ with project name, keep as is
	serve        *bool   // Serve the configuration over http to preview
	openBrowser  *bool   // Open browser automatically when serving
}

func (fr *FlagRunner) Run() {
	if *fr.encoder != "" {
		switch strings.TrimSpace(strings.ToLower(*fr.encoder)) {
		case "json":
			AppConfig.Encoder = "json"
		case "gob":
			AppConfig.Encoder = "gob"
		default:
			fmt.Println(Craft(CMD_Red, "Invalid encoder type"))
			os.Exit(1)
		}
	}

	if *fr.importpath != "" {
		_, err := InitProjectConfig(*fr.importpath)
		if err != nil {
			panic(err)
		}
	} else if *fr.serve {
		dirnames := ListInternalConfigs()
		dirnames = append(dirnames, ListConfigs()...)
		viewer := NewViewer(dirnames, *fr.raw)
		if err := viewer.serve(*fr.openBrowser); err != nil {
			panic(err)
		}
	} else if *fr.config_name != "" {
		dir, err := GetDir(*fr.config_name, *fr.proj_name, *fr.raw)
		if err != nil {
			panic(err)
		}
		if *fr.view_config {
			ListFiles(dir, "")
			return
		} else if *fr.del_conf {
			err := DeleteConfig(*fr.config_name)
			if err != nil {
				panic(err)
			}
			return
		}
		_, err = InitProject(*fr.config_name, *fr.proj_name, dir)
		if err != nil {
			panic(err)
		}
	} else if *fr.list_configs {
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

	} else if *fr.get_config != "" {
		dir, err := GetDirFromPath(*fr.get_config)
		if err != nil {
			panic(err)
		}
		if *fr.proj_name == "" {
			*fr.proj_name = *fr.get_config
		}
		*fr.proj_name = URLOmit(*fr.proj_name)
		if strings.EqualFold(*fr.proj_name, "static") {
			*fr.proj_name = strings.Replace(strings.ToLower(*fr.proj_name), "static", "tpl_static", 1)
			fmt.Println(Craft(CMD_Red, "Warning: The project name contains 'static' which is reserved for static files when serving.\n The project name will be changed to: "+*fr.proj_name))
		}
		err = AppConfig.Serialize(dir, EXE_DIR+"\\conf\\"+*fr.proj_name)
		if err != nil {
			panic(err)
		}
	} else if *fr.location {
		PrintLocation()
	} else {
		PrintLogo()
		flag.CommandLine.Usage()
		os.Exit(1)
	}

}
