package quickgo

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		dirPath    = GetQuickGoPath(config.COMMANDS_DIR)
		scriptName = filepath.Base(path)
		outputName = filepath.Join(dirPath, scriptName)
		scriptSrc  *os.File
		file       *os.File
		err        error
	)

	path = filepath.FromSlash(path)

	logger.Infof("Copying file %s to %s", path, outputName)

	if s, err := os.Stat(path); err != nil || s.IsDir() {
		return errors.Wrapf(err, "file %s does not exist or is not a valid file", path)
	}

	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", dirPath)
	}

	file, err = os.Create(
		outputName,
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
	if err != nil {
		return errors.Wrapf(err, "failed to copy file %s to %s", path, file.Name())
	}

	logger.Infof("Command '%s' saved to '%s'", scriptName, outputName)

	return nil
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

	logger.Debugf("Reading script '%s'", scriptPath)

	script, err = os.ReadFile(scriptPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read script %s", script)
	}

	logger.Debugf("'%s' has access to project config: %v", scriptName, a.ProjectConfig != nil)
	var (
		projectName string
		projectPath string
	)
	if a.ProjectConfig != nil {
		projectName = a.ProjectConfig.Name
		projectPath = getTargetDirectory(
			targetDir,
		)
	}

	var (
		vm            = goja.New()
		quickgoModule = map[string]any{
			"app":         a,
			"project":     a.ProjectConfig,
			"config":      a.Config,
			"environ":     args,
			"projectName": projectName,
			"projectPath": projectPath,
			"projectLocked": func() bool {
				return config.IsLocked(targetDir) == nil
			},
		}
		fsModule = map[string]any{
			"cleanPath": filepath.Clean,
			"joinPath":  filepath.Join,
			"absPath": func(path string) string {
				path, err = filepath.Abs(path)
				if err != nil {
					logger.Error(err)
					return ""
				}
				return path
			},
			"getWd": func() string {
				return getTargetDirectory(targetDir)
			},
			"writeFile": func(path string, data goja.Value) error {
				var b []byte
				if vm.InstanceOf(data, vm.Get("Uint8Array").ToObject(vm)) {
					b = data.Export().([]byte)
				} else if d, ok := data.Export().(goja.ArrayBuffer); ok {
					b = d.Bytes()
				} else {
					b = []byte(data.String())
				}

				err = os.WriteFile(path, b, os.ModePerm)
				if err != nil {
					logger.Error(err)
					return err
				}
				logger.Debugf("File '%s' written by '%s'", path, scriptPath)
				return nil
			},
			"readFile": func(path string) goja.Value {
				var data []byte
				var err error
				if data, err = os.ReadFile(path); err != nil {
					logger.Error(err)
					return goja.Undefined()
				}

				logger.Debugf("File '%s' read by '%s'", path, scriptPath)

				var arr = vm.NewArrayBuffer(data)
				return vm.ToValue(arr)
			},
			"readTextFile": func(path string) string {
				var data, err = os.ReadFile(path)
				if err != nil {
					logger.Error(err)
					return ""
				}

				logger.Debugf("File '%s' read by '%s'", path, scriptPath)

				return string(data)
			},
			"readDir": func(path string) goja.Value {

				path, err = filepath.Abs(path)
				if err != nil {
					logger.Error(err)
					return goja.Undefined()
				}

				var dir, err = os.ReadDir(path)
				if err != nil {
					logger.Error(err)
					return goja.Undefined()
				}

				var d = make([]any, len(dir))
				for i, f := range dir {
					d[i] = map[string]any{
						"name":        f.Name(),
						"isDirectory": f.IsDir(),
						"isFile":      f.Type().IsRegular(),
						"path": filepath.Join(
							path, f.Name(),
						),
					}
				}

				return vm.ToValue(
					vm.NewArray(d...),
				)
			},
		}
		osModule = map[string]any{
			"args": os.Args,
			"getEnv": func(key string) string {
				return os.Getenv(key)
			},
			"setEnv": func(key, value string) error {
				logger.Debugf("Setting environment variable '%s'='%s'", key, value)
				return os.Setenv(key, value)
			},
			"exec": func(cmd string, commandArgs ...string) goja.Value {

				logger.Debugf(
					"Executing command from '%s': %s %s",
					scriptPath, cmd, strings.Join(commandArgs, " "),
				)

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
							logger.Errorf(
								"failed to split command arguments for: %s: %s",
								scriptPath, err,
							)
							return vm.ToValue(map[string]any{
								"stdout":   "",
								"exitCode": 1,
								"error": fmt.Sprintf(
									"failed to split command arguments: %s", err,
								),
							})
						}
					} else {
						logger.Warnf(
							"command arguments are provided, but the command is already split for: %s",
							scriptPath,
						)
					}
				}

				var command = exec.Command(cmd, commandArgs...)
				var s = new(strings.Builder)

				command.Stdout = s
				command.Stderr = s

				if err = command.Run(); err != nil {
					logger.Errorf(
						"failed to execute command: %s: %s",
						scriptPath, err,
					)
					return vm.ToValue(map[string]any{
						"stdout":   s.String(),
						"exitCode": 1,
						"error":    err,
					})
				}

				logger.Debugf(
					"Command executed from '%s': %s %s",
					scriptPath, cmd, strings.Join(commandArgs, " "),
				)

				return vm.ToValue(map[string]any{
					"stdout":   s.String(),
					"exitCode": command.ProcessState.ExitCode(),
					"error":    err,
				})
			},
		}
	)

	vm.SetParserOptions()

	cmd = js.NewScript(
		"main",
		js.WithVM(vm),
		js.WithGlobals(map[string]any{
			"quickgo": quickgoModule,
			"os":      osModule,
			"fs":      fsModule,
		}),
	)

	cmd.AddFunc(func(vm *goja.Runtime) error {
		vm.Set("console", &js.JSConsole{
			Debug: logger.Debug,
			Log:   logger.Info,
			Info:  logger.Info,
			Warn:  logger.Warn,
			Error: logger.Error,
			Fatal: func(a ...any) {
				logger.Fatal(1, a...)
			},
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

		return nil
	})

	logger.Debugf("Executing script '%s'", scriptPath)

	result, err := cmd.Run(string(script))
	if err != nil {
		if _, ok := err.(*js.CommandError); !ok {
			return errors.Wrapf(err, "failed to execute script '%s'", scriptName)
		}
	}

	var s = fmt.Sprintf("Script '%s' executed with exit code %d", scriptName, result.Importance)
	if result.Message != "" {
		s = fmt.Sprintf("%s: %s", s, result.Message)
	}

	if result.Importance == 0 {
		logger.Info(s)
	} else {
		logger.Error(s)
	}

	return js.ErrExitCode
}
