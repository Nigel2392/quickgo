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
	if *fr.encoder != "" { // Set encoder type
		err := AppConfig.SetEncoder(*fr.encoder)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if *fr.importpath != "" { // Import configuration
		_, err := InitProjectConfig(*fr.importpath)
		if err != nil {
			panic(err)
		}
	} else if *fr.del_conf { // Delete configuration
		if *fr.config_name == "" {
			fmt.Println("No configuration specified, use -use to specify a configuration")
			return
		}
		err := DeleteConfig(*fr.config_name)
		if err != nil {
			panic(err)
		}
		return
	} else if *fr.serve { // Serve configuration
		dirnames := ListInternalConfigs()
		dirnames = append(dirnames, ListConfigs()...)
		viewer := NewViewer(dirnames, *fr.raw)
		if err := viewer.serve(*fr.openBrowser); err != nil {
			panic(err)
		}
	} else if *fr.config_name != "" { // Use configuration
		dir, err := GetDir(*fr.config_name, *fr.proj_name, *fr.raw)
		if err != nil {
			panic(err)
		}
		if *fr.view_config { // View configuration
			ListFiles(dir, "")
			return
		}
		_, err = InitProject(*fr.config_name, *fr.proj_name, dir)
		if err != nil {
			panic(err)
		}
	} else if *fr.list_configs { // List configurations
		int_confs := ListInternalConfigs()
		ext_confs := ListConfigs()
		sort.Slice(int_confs, func(i, j int) bool {
			return strings.ToLower(int_confs[i]) < strings.ToLower(int_confs[j])
		})
		sort.Slice(ext_confs, func(i, j int) bool {
			return strings.ToLower(ext_confs[i]) < strings.ToLower(ext_confs[j])
		})
		for _, conf := range int_confs { // Print internal configurations
			fmt.Println(Craft(CMD_Purple, conf))
		}
		for _, conf := range ext_confs { // Print external configurations
			fmt.Println(Craft(CMD_Blue, conf))
		}

	} else if *fr.get_config != "" { // Generate configuration
		dir, err := GetDirFromPath(*fr.get_config)
		if err != nil {
			panic(err)
		}
		if *fr.proj_name == "" { // No project name specified
			*fr.proj_name = *fr.get_config
		}
		rp_name := URLOmit(*fr.proj_name)
		fr.proj_name = &rp_name
		if strings.EqualFold(*fr.proj_name, "static") {
			*fr.proj_name = strings.Replace(strings.ToLower(*fr.proj_name), "static", "tpl_static", 1)
			fmt.Println(Craft(CMD_Red, "Warning: The project name equals 'static' which is reserved for static files when serving.\n The project name will be changed to: "+*fr.proj_name))
		}
		err = AppConfig.Serialize(dir, EXE_DIR+"\\conf\\"+*fr.proj_name)
		if err != nil {
			panic(err)
		}
	} else if *fr.location { // Print current directory
		PrintLocation()
	} else { // Print help
		PrintLogo()
		flag.CommandLine.Usage()
		os.Exit(1)
	}

}
