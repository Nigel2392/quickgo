package command

import "encoding/gob"

func init() {
	gob.Register(&StepList{})
	gob.Register(&Step{})
}
