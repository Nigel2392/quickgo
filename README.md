<center>
    <img src="https://github.com/Nigel2392/quickgo/blob/main/.github/assets/quickgo.png?raw=true" alt="QuickGo Logo"/>
</center>

# Why QuickGo?

QuickGo is a simple and easy to use golang command line tool for creating projects of any language.

**Support for:**

- Variable usage with custom delimiters defined in `quickgo.yaml`.
- Custom commands ran with `quickgo <command_name> <args>` defined in `quickgo.yaml`.
- Executing commands before and after copying project templates.
- Excluding files from the project templates.
- Serving over HTTP for easy viewing of projects.

# Installation

quickgo can be installed using the following command:

```bash
go install github.com/Nigel2392/quickgo@2.1.6
```

# Usage

(This part will be expanded upon in the near future)

```bash
 $$$$$$\            $$\           $$\           $$$$$$\
$$  __$$\           \__|          $$ |         $$  __$$\
$$ /  $$ |$$\   $$\ $$\  $$$$$$$\ $$ |  $$\    $$ /  \__| $$$$$$\   ####
$$ |  $$ |$$ |  $$ |$$ |$$  _____|$$ | $$  |   $$ |$$$$\ $$  __$$\
$$ |  $$ |$$ |  $$ |$$ |$$ /      $$$$$$  /    $$ |\_$$ |$$ /  $$ |   ######
$$ $$\$$ |$$ |  $$ |$$ |$$ |      $$  _$$<     $$ |  $$ |$$ |  $$ |
\$$$$$$ / \$$$$$$  |$$ |\$$$$$$$\ $$ | \$$\    \$$$$$$  |\$$$$$$  | #####
 \___$$$\  \______/ \__| \_______|\__|  \__|    \______/  \______/
     \___|


Created by: Nigel van Keulen
QuickGo: A simple project generator and server.
Usage: quickgo [flags | command] [?args]
Available application flags:
  -d
    d: The target directory to write the project to.
  -delim-left
    delim-left: The left delimiter for the project templates.
  -delim-right
    delim-right: The right delimiter for the project templates.
  -e
    e: A list of files to exclude from the project in glob format.
  -example=false
    example: Print an example project configuration.
  -get
    get: Import the project from the current directory.
  -host=localhost
    host: The host to run the server on.
  -list=false
    list: List the projects available for use.
  -name
    name: The name of the project.
  -port=8080
    port: The port to run the server on.
  -serve=false
    serve: Serve the project over HTTP.
  -tls-cert
    tls-cert: The path to the TLS certificate.
  -tls-key
    tls-key: The path to the TLS key.
  -use
    use: Use the specified project configuration.
  -v=false
    v: Enable verbose logging.
```
