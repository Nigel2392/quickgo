# Why QuickGo?

* Quickgo is a simple and easy to use golang command line tool for creating projects of any language.  
* Easily import your existing templates into QuickGo with `quickgo -get (filename)`.
* Share your templates, and import them with `quickgo -import (filename.json)`.
* View your imported templates with `quickgo -v (name of project)`.
* List all your imported templates with `quickgo -l`.
* Define a name for each project, display it by using `$$PROJECT_NAME$$` in your templates, then use the template.
* Quickly preview any template you've made via `quickgo -serve -o` (`-o` Is optional to open browser right away.)

## Information:
When using `-l`, the default embedded configs are highlighted in cyan. 
These default configs are only available via the "-d (template name)" argument.

## Warning:
* QuickGo uses JSON serialization to save the file to disk, which means you cannot use some binary data in your templates. 
  * This however does mean that you can use any text editor to edit your templates, and easily share them with others.
  * You have to explicitly set the encoding to `gob` in the `config.json` stored in the same directory as the executable.

* When using -serve, QuickGo might use a lot of memory depending on the size of your template configurations. 
  We load all the files into memory before serving them, this is to save on loading times.


## Installation:
```
go install github.com/Nigel2392/quickgo
```
## Available commands:
```
.\quickgo.exe -h
```
- **-serve**
  - Serve the project files in the browser.
- **-del**
  - Delete a config (usage: quickgo -use (config name) -del)
- **-get**
  - Get the JSON config of the project (usage: quickgo -get (relative directory name))
- **-import**
  - Path of the JSON file to be imported (usage: quickgo -import (path))
- **-l**    
  - List all the available configs (usage: quickgo -l)
- **-loc**
  - Location of the executable (usage: quickgo -loc (path))
- **-n**
  - Name of the project to be created (usage: quickgo -use (config name) (optional: -n (name of project))
    - Replaces the `$$PROJECT_NAME$$` in the template with the name of the project.
    - Replaces the `$$PROJECT_NAME; OMITURL$$` in the template with the name of the project, assuming the name is the last part of the URL passsed into -n
      - This is useful when creating Golang packages
- **-use**
  - Path of the JSON file to use for creating templates (usage: quickgo -use (config name))
- **-v**
  - View the config of the project (usage: quickgo -use (config name) -v)
- **-raw**
  - Write the raw config as a project, will not replace `$$PROJECT_NAME$$`
- **-enc **
        Encoder to use for the project (json/gob). Can also be set in the `config.json`