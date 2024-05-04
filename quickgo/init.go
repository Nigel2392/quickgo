package quickgo

import (
	"embed"
	"encoding/gob"
	"io/fs"
	"os"

	"github.com/pkg/errors"
)

var (
	quickGoConfigDir string

	//go:embed _templates
	embedFS embed.FS

	staticFS fs.FS
)

func init() {
	var err error
	quickGoConfigDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	gob.Register(&App{})

	staticFS, err = fs.Sub(embedFS, "_templates/static")
	if err != nil {
		panic(
			errors.Wrap(err, "failed to create staticFS"),
		)
	}

}
