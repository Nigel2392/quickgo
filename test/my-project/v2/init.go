package quickgo

import (
	"encoding/gob"
	"os"
	"path/filepath"
)

var executableDir string

func init() {
	var executable, err = os.Executable()
	if err != nil {
		panic(err)
	}

	executableDir = filepath.Dir(executable)

	gob.Register(&App{})

}
