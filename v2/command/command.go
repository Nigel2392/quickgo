package command

type (
	Step interface {
		Name() string
		Execute() error
	}

	// List represents a list of steps to run.
	List interface {
		// Steps returns the list of steps.
		Steps() []Step

		// Execute runs the list of steps.
		Execute() error
	}
)
