package cli

import (
	"os"
)

// OS runtime
type Runtime struct {
	Args []string
	Env  []string
}

var DefaultRuntime = Runtime{
	Args: os.Args,
	Env:  os.Environ(),
}
