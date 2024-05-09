<center>
    <img src="https://github.com/Nigel2392/quickgo/blob/main/quickgo/_templates/static/quickgo.png?raw=true" alt="QuickGo Logo"/>
</center>

# Why QuickGo?

QuickGo is a simple and easy to use golang command line tool for creating projects of any language.

**Support for:**

- Saving project templates for easy reuse.
- Variable usage with custom delimiters defined in `quickgo.yaml` (uses Go's `text/template` package).
- Custom commands/scripts written in Javascript for more advanced management of more complex projects.
- Custom commands ran with `quickgo <command_name> <args>` defined in `quickgo.yaml`.
- Executing commands before and after copying project templates.
- Excluding files from the project templates.
- Serving over HTTP for easy viewing of projects.

# Installation

quickgo can be installed using `go install` with the following command:

```bash
go install github.com/Nigel2392/quickgo/v2@v2.4.9
```

Optionally we provide a binary for Linux, MacOS and Windows.

These can be found in the [releases](https://github.com/Nigel2392/quickgo/releases) section.

After downloading the binary, you should add it to your PATH.

# Usage

## List all the projects you've saved

To list all the projects which you've saved we provide the `-list` flag.

Example:

```bash
quickgo -list
```

## Configuring your project templates

### Example configuration

We provide a command to easily generate an example configuration file for quickgo.

This provides a good starting point for creating your own project configurations.

**It is however fully possible to skip creating a `quickgo.yaml` file and control everything via the command line.**

```bash
quickgo -example
```

Now you should have a `quickgo.yaml` file in your current directory.

### `quickgo.yaml`

The `quickgo.yaml` file is where you define your project templates and commands.

It allows for the following fields:

- `name`: The name of the project.
- `context`: The context to use for the project templates.
- `delimLeft`: The left delimiter for the project templates.
- `delimRight`: The right delimiter for the project templates.
- `exclude`: A list of files to exclude from the project in glob format. (e.g. `*.go`)
- `beforeCopy`: A list of commands to run before copying the project templates.
- `afterCopy`: A list of commands to run after copying the project templates.s
- `commands`: A list of commands to run before and after copying the project templates.

## Using the template engine

The template engine uses Go's `text/template` package.

This allows for the use of variables in your project files.

Variables are by default delimited by `{{` and `}}`, this can be changed in the `quickgo.yaml` file.

Example of a variable in a file:

Let's take the example `quickgo.yaml` file at the end of the README to demonstrate how this works.

```text
# {{ .Name }}

# This is a project template for {{ index .Context "CustomName" }}.

Read more:

{{ index .Context "Description" }}
```

### Note: Filepaths

Filepaths in the project template can also contain variables!

## Saving your project templates

After you have configured your project in `quickgo.yaml`, you can save them with the following commands.

```bash
# Save the project (this will be saved in $HOME/.quickgo/projects).
quickgo -save 

# Save the project, override the name and context:
quickgo -save -name my-custom-project / customContextKey=customContextValue

# Save the project, exclude all files matching the glob pattern `*.go` and `*.mod`.
quickgo -save -e '*.go' -e '*.mod'
```

## Using your project templates

Now that you have created some project templates, you can use them with the following command:

```bash
# Create a new project from the template.
# This will create a new project in the `my/target/directory/my-project` directory.
quickgo -use my-project -d my/target/directory
```

We also allow for a few overrides when using your project templates, you can provide extra context or change the project name.

This can be done by delimiting your context with a slash when running the above command.

Example:

```bash
# Create a new project from the template.
# Change the name, provide extra context and change the target directory.
# This will create a new project in the `my/target/directory/my-custom-project-name` directory.
quickgo -use my-project -d my/target/directory -name my-custom-project-name / customContextKey=customContextValue
```

## Serving your project templates

You can also serve your project templates over HTTP.

This can be done with the following command:

```bash
# Serve the project over HTTP.
quickgo -serve
```

You can also provide a custom host and port to serve the project on.

```bash
# Serve the project over HTTP on a custom host and port.
quickgo -serve -host 127.0.0.1 -port 8080
```

Or even serve the project over HTTPS.

```bash
# Serve the project over HTTPS.
quickgo -serve -port 443 -tls-cert /path/to/cert.pem -tls-key /path/to/key.pem
```

## Locking the project configuration.

If you want to lock the project configuration, you can do so by providing the `-lock` flag.

This will prevent the project configuration and files from being modified by QuickGo.

* 1 = Lock
* 0 = Unlock

Example:

```bash
# Lock the current project configuration.
quickgo -lock 1
```

## Using your defined project's commands

You can also run the commands defined in your `quickgo.yaml` file.

It is also possible to provide the `-d` flag to run the commands in the project's directory.

```bash
# Run the `echoName` command defined in the `quickgo.yaml` file.
# This will echo the project name.
quickgo echoName

# Run the `echoName` command defined in the `myconfig/quickgo.yaml` file.
# Note the shell delimiters! These are NOT defined in the `quickgo.yaml` file,
# but instead handled by the OS.
quickgo -d 'myconfig' echoName customProjectName="custom-${projectName}"
```

These are addressed by the label of the steps.

Example:

```bash
# Run the `echoName` command defined in the `quickgo.yaml` file.
# This will echo the project name.
quickgo echoName

# Run the `echoName` command defined in the `quickgo.yaml` file.
# Provide a custom project name to echo.
# This will echo the custom project name.
quickgo echoName customProjectName="custom-${projectName}"
```

## Global commands

### Saving global commands

You can also save global commands for your user.

These will be stored in `$HOME/.quickgo/commands`.

These should be javascript files.

Example:

```bash
# Save a global command for this user.
quickgo -save-command /path/to/my-command.js
```

Where the `my-command.js` file should look like this:

For a more advanced examples, see the [commands](./commands) directory.

```javascript
function main() {
    console.debug("Generating random number...");
    let rand = Math.floor(Math.random() * 10);
    if (rand > 5) {
        return Result(1, "Failed");
    }

    if (!quickgo.environ.myVar) {
        return Result(1, "myVar not set, please run `quickgo exec my-command myVar=myValue`");
    }

    return Result(0, "Success");
}
```

### Running global commands

You can run global commands with the following command:

```bash
# Run the global command.
quickgo exec my-command myVar=myValue
``` 

`my-command` is the name of the javascript file minus the `.js` extension.

### Listing global commands

You can list the global commands with the following command:

```bash
# List the global commands.
quickgo -list-commands
```


## All available application flags:

```bash
-d: The target directory to write the project to.
-delim-left: The left delimiter for the project templates.
-delim-right: The right delimiter for the project templates.
-e: A list of files to exclude from the project in glob format.
-example=false: Print an example project configuration.
-host=localhost: The host to run the server on.
-list=false: List the projects available for use.
-list-commands=false: List the commands available for all projects.
-lock=-1: Lock the project configuration. 1=Lock, 0=Unlock.
-name: The name of the project.
-port=8080: The port to run the server on.
-save=false: Import the project from the current directory.
-save-command: Save a global command for this user by providing a path to a JS file.
-serve=false: Serve the project over HTTP.
-tls-cert: The path to the TLS certificate.
-tls-key: The path to the TLS key.
-use: Use the specified project configuration.
-v: Enable verbose logging.
```

### Example `quickgo.yaml` configuration

See the example file below to get a better understanding of how to configure your project templates.

```yaml
# The name of your project template.
# This is how you want it to be adressed when using the template.
# It can optionally be overridden, example: `quickgo -get my-project -name my-custom-project-name`
name: my-project

# Optional extra context that can be used in text files throughout your project.
# These can also be used when running commands like beforeCopy, afterCopy and the project commands themselves.
# Example: `{{.Name}}` will be replaced with `my-project` in all files, if the leftDelim and rightDelim are set to `{{` and `}}`.
# Context variables can be addressed with `{{ index .Context "CustomName" }}`.
context:
    CustomName: My Project
    Description: This is a project template for My Project.

    # Version mapping for the command defined in commands/version.js
    # Example: `quickgo exec version v=1.0.0`
    versionMapping:
      README.md: "\\@v((?:(\\d+)\\.)?(?:(\\d+)\\.)?(\\*|\\d+))"
      
# The left and right delimiters for the template engine.
# These are used to replace the context variables in the project files.
# Example: `{{ .Name }}` will be replaced with `my-project` in all files, if the leftDelim and rightDelim are set to `{{` and `}}`.
delimLeft: '${{'
delimRight: '}}'

# A list of files and directories to exclude from the project template.
# These will not be saved; there is currently no way to exclude files after saving.
exclude:
    - '*node_modules*'
    - '*dist*'
    - .git

# A list of steps to run before copying the template to the destination.
# This can be used to prepare the project files, etc.
# The project is not yet copied to the destination at this point.
beforeCopy:
    # This will execute:
    #   echo $projectPath  
    steps:
        - name: Echo Project Path Before
          command: echo
          args:
            - $projectPath

# A list of steps to run after copying the template to the destination.
# This can be used to clean up, install dependencies, etc.
# The project is already copied to the destination at this point.
afterCopy:
    steps:
        - name: Echo Project Name After
          command: echo
          args:
            - $projectName

# A list of custom commands that can be run on the project.
# These can be used to automate tasks, etc.
# Example: `quickgo <command-name> <args>`
commands:

    # The label is the name of the command.
    echoName:

        # Description is only used for type annotations.
        description: Echo the project name, supply one with customProjectName=myValue

        # Arguments which can be setup before running the command.
        args:
            customProjectName: $projectName

        # The steps which will be executed when running the command.
        steps:
            - name: Echo Project Name
              command: echo
              args:
                - $customProjectName
```
