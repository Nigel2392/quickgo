package quickgo

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nigel2392/quickgo/v2/quickgo/command"
	"github.com/Nigel2392/quickgo/v2/quickgo/config"
	"github.com/Nigel2392/quickgo/v2/quickgo/js"
	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
	"github.com/dop251/goja"
	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
)

func (a *App) ListJSFiles() ([]string, error) {
	var (
		dirPath = GetQuickGoPath(config.COMMANDS_DIR)
		files   = make([]string, 0)
	)

	dir, err := os.ReadDir(dirPath)
	if err != nil && os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", dirPath)
	}

	for _, d := range dir {
		if d.IsDir() {
			continue
		}

		if strings.HasSuffix(d.Name(), ".js") {
			files = append(files, d.Name()[0:len(d.Name())-3])
		}
	}

	return files, nil
}

func (a *App) SaveJS(path string) error {
	var (
		dirPath   = GetQuickGoPath(config.COMMANDS_DIR)
		filename  = filepath.Join(dirPath, filepath.Base(path))
		scriptSrc *os.File
		file      *os.File
		err       error
	)

	path = filepath.FromSlash(path)

	logger.Infof("Copying file %s to %s", path, filename)

	if s, err := os.Stat(path); err != nil || s.IsDir() {
		return errors.Wrapf(err, "file %s does not exist or is not a valid file", path)
	}

	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", dirPath)
	}

	file, err = os.Create(
		filename,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", path)
	}
	defer file.Close()

	scriptSrc, err = os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", path)
	}

	_, err = io.Copy(file, scriptSrc)
	return errors.Wrapf(err, "failed to copy file %s to %s", path, file.Name())
}

func (a *App) ExecJS(targetDir string, scriptName string, args map[string]any) (err error) {
	var (
		scriptPath = GetQuickGoPath(
			config.COMMANDS_DIR,
			fmt.Sprintf("%s.js", scriptName),
		)

		script []byte
		cmd    *js.Command
	)

	script, err = os.ReadFile(scriptPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read script %s", script)
	}

	var (
		projectName string
		projectPath string
	)
	if a.ProjectConfig != nil {
		projectName = a.ProjectConfig.Name
		projectPath = getTargetDirectory(
			projectName,
		)
	}

	cmd = js.NewScript(
		"main",
		js.WithGlobals(map[string]any{
			"quickgo": map[string]any{
				"app":         a,
				"project":     a.ProjectConfig,
				"config":      a.Config,
				"environ":     args,
				"projectName": projectName,
				"projectPath": projectPath,
			},
			"os": map[string]any{
				"args": os.Args,
				"getEnv": func(key string) string {
					return os.Getenv(key)
				},
				"getWd": func() string {
					return getTargetDirectory(targetDir)
				},
				"setEnv": func(key, value string) error {
					return os.Setenv(key, value)
				},
				"exec": func(cmd string, commandArgs ...string) error {

					// Parse the command arguments if the command is provided as a single string.
					// Example: echo "Hello, World!" -> echo, ["Hello, World!"]
					if strings.Index(cmd, " ") > 0 {
						var s = strings.SplitN(cmd, " ", 2)
						cmd = s[0]
						if len(commandArgs) == 0 {
							commandArgs, err = shellquote.Split(
								s[1],
							)
							if err != nil {
								return errors.Wrapf(
									err,
									"failed to split command arguments for: %s",
									scriptPath,
								)
							}
						} else {
							logger.Warnf(
								"command arguments are provided, but the command is already split for: %s",
								scriptPath,
							)
						}
					}

					var c = command.Step{
						Name:    fmt.Sprintf("%s %s", cmd, strings.Join(commandArgs, " ")),
						Command: cmd,
						Args:    commandArgs,
					}

					return c.Execute(nil)
				},
			},
		}),
	)

	cmd.AddFunc(func(vm *goja.Runtime) error {
		vm.Set("console", &js.JSConsole{
			Debug: logger.Debug,
			Log:   logger.Info,
			Info:  logger.Info,
			Warn:  logger.Warn,
			Error: logger.Error,
		})
		vm.Set("base64Encode", func(data goja.Value) string {
			if vm.InstanceOf(data, vm.Get("Uint8Array").ToObject(vm)) {
				var arr = data.Export().([]byte)
				return base64.StdEncoding.EncodeToString(arr)
			}

			if data, ok := data.Export().(goja.ArrayBuffer); ok {
				return base64.StdEncoding.EncodeToString(data.Bytes())
			}

			var s = data.String()
			return base64.StdEncoding.EncodeToString([]byte(s))
		})
		vm.Set("base64Decode", func(data string) goja.Value {
			var arr, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				logger.Error(err)
				return goja.Undefined()
			}

			var arrBuff = vm.NewArrayBuffer(arr)
			return vm.ToValue(arrBuff)
		})
		vm.Set("writeFile", func(path string, data goja.Value) error {
			var b []byte
			if vm.InstanceOf(data, vm.Get("Uint8Array").ToObject(vm)) {
				b = data.Export().([]byte)
			} else if d, ok := data.Export().(goja.ArrayBuffer); ok {
				b = d.Bytes()
			} else {
				b = []byte(data.String())
			}

			return os.WriteFile(path, b, os.ModePerm)
		})
		vm.Set("readFile", func(path string) goja.Value {
			var data []byte
			var err error
			if data, err = os.ReadFile(path); err != nil {
				logger.Error(err)
				return goja.Undefined()
			}

			var arr = vm.NewArrayBuffer(data)
			return vm.ToValue(arr)
		})
		vm.Set("readTextFile", func(path string) string {
			var data, err = os.ReadFile(path)
			if err != nil {
				logger.Error(err)
				return ""
			}
			return string(data)
		})

		return nil
	})

	return cmd.Run(string(script))
}
