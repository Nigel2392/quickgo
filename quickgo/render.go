package quickgo

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
)

const (
	CMD_Blue          = "\033[34m"
	CMD_Cyan          = "\033[36m"
	CMD_Green         = "\033[32m"
	CMD_BRIGHT_Purple = "\033[35;1m"
	CMD_Purple        = "\033[35m"
	CMD_Red           = "\033[31m"
	CMD_Yellow        = "\033[33m"
	CMD_Bold          = "\033[1m"
	CMD_Underline     = "\033[4m"
	CMD_Reset         = "\033[0m"
)

func Craft(color, s any) string {
	return fmt.Sprintf("%s%v%s", color, s, CMD_Reset)
}

func BuildColorString(colors ...string) string {
	var s strings.Builder
	for _, color := range colors {
		s.WriteString(color)
	}
	s.WriteString(CMD_Reset)
	return s.String()
}

func PrintLogo() {
	// Quick GO logo.
	str := Craft(CMD_Cyan, " $$$$$$\\            $$\\           $$\\         "+Craft(CMD_Cyan, "  $$$$$$\\\n")) +
		Craft(CMD_Cyan, "$$  \033[31m__\033[36m$$\\           \033[31m\\__|\033[36m          $$ |      "+Craft(CMD_Cyan, "   $$  __$$\\ \n")) +
		Craft(CMD_Blue, "$$ \033[31m/\033[34m  $$ |$$\\   $$\\ $$\\  $$$$$$$\\ $$ |  $$\\ "+Craft(CMD_Cyan, "   $$ /  \\__| $$$$$$\\   ####\n")) +
		Craft(CMD_Blue, "$$ \033[31m|\033[34m  $$ |$$ |  $$ |$$ |$$  \033[31m_____|\033[34m$$ | $$  \033[31m|\033[34m  "+Craft(CMD_Cyan, " $$ |$$$$\\ $$  __$$\\\n")) +
		Craft(CMD_Blue, "$$ \033[31m|\033[34m  $$ |$$ |  $$ |$$ |$$ \033[31m/\033[34m      $$$$$$  \033[31m/\033[34m   "+Craft(CMD_Cyan, " $$ |\\_$$ |$$ /  $$ |   ######\n")) +
		Craft(CMD_Purple, "$$ $$\\$$ |$$ |  $$ |$$ |$$ \033[31m|\033[35m      $$  _$$<   "+Craft(CMD_Cyan, "  $$ |  $$ |$$ |  $$ |\n")) +
		Craft(CMD_Purple, "\\$$$$$$ / \\$$$$$$  |$$ |\\$$$$$$$\\ $$ | \\$$\\   "+Craft(CMD_Cyan, " \\$$$$$$  |\\$$$$$$  | #####\n")) +
		Craft(CMD_Red, " \\___"+CMD_Reset+Craft(CMD_Purple, "$$$")+Craft(CMD_Red, "\\  \\______/ \\__| \\_______|\\__|  \\__| ")+Craft(CMD_Cyan, "   \\______/  \\______/\n")) +
		Craft(CMD_Red, "     \\___|")
	fmt.Println(str)
	fmt.Println(Craft(CMD_Red, "\nCreated by: ") + Craft(CMD_Purple, "Nigel van Keulen"))

	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		fmt.Printf(Craft(CMD_Cyan, "Version: %s\n"), info.Main.Version)
	}

	// Example:
	// QUICKGO (but very big)
	// Created by: Nigel van Keulen
	// Version: vX.X.X
}

func wrapLog(colors ...string) func(l logger.LogLevel, s string) string {
	var s strings.Builder
	for _, color := range colors {
		s.WriteString(color)
	}
	var prefix = s.String()
	return func(l logger.LogLevel, s string) string {
		return Craft(prefix, s)
	}
}

func ColoredLogWrapper(l logger.LogLevel, s string) string {
	var fn, ok = logWrapperMap[l]
	if !ok {
		return s
	}
	return fn(l, s)
}

var logWrapperMap = map[logger.LogLevel]func(l logger.LogLevel, s string) string{
	logger.DebugLevel: wrapLog(CMD_Green),
	logger.InfoLevel:  wrapLog(CMD_Cyan),
	logger.WarnLevel:  wrapLog(CMD_Yellow),
	logger.ErrorLevel: wrapLog(CMD_Red, CMD_Bold),
}
