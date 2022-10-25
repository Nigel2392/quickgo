# Why QuickGo?

* Quickgo is a simple and easy to use golang command line tool for creating projects of any language.  
* Easily import your existing templates into QuickGo with `quickgo -get (filename)`.
* Share your templates, and import them with `quickgo -import (filename.json)`.
* View your imported templates with `quickgo -v (name of project)`.
* List all your imported templates with `quickgo -l`.
* Define a name for each project, display it by using `$$PROJECT_NAME$$` in your templates, then use the template.

## Information:
When using `-l`, the default embedded configs are highlighted in purple. 
These default configs are only available via the "-d (template name)" argument.

## Installation:
```
go install github.com/Nigel2392/quickgo
```
## Available commands:
```
.\quickgo.exe -h
```

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