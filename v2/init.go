package quickgo

import (
	"embed"
	"encoding/gob"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

var (
	executableDir string

	//go:embed _templates
	embedFS embed.FS

	staticFS fs.FS
)

func init() {
	var executable, err = os.Executable()
	if err != nil {
		panic(err)
	}

	executableDir = filepath.Dir(executable)

	gob.Register(&App{})

	staticFS, err = fs.Sub(embedFS, "_templates/static")
	if err != nil {
		panic(
			errors.Wrap(err, "failed to create staticFS"),
		)
	}

}
