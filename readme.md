# Why QuickGo?

* Quickgo is a simple and easy to use golang command line tool for creating projects of any language.  
* Easily import your existing templates into QuickGo with `quickgo -get (filename)`.
* Share your templates, and import them with `quickgo -import (filename.json)`.
* View your imported templates with `quickgo -v (name of project)`.
* List all your imported templates with `quickgo -l`.

## Installation:
```
go install github.com/Nigel2392/quickgo
```
## Available commands:

```
.\quickgo.exe -h
```

- **del**
  - Delete a config
- **dir string**
  - The directory to create the project in
- **get string**
  - Get the JSON config of the project
- **import string**
  - Path of the JSON file to be imported
- **l**    
  - List all the available configs
- **loc**
  - Location of the executable
- **n string**
  - Name of the project to be created
- **use string**
  - Path of the JSON file to use for creating templates
- **v**
  - View the config of the project