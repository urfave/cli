package altsrc

import "github.com/urfave/cli/v2"

// defaultInputSource creates a default cli.InputSourceContext.
func defaultInputSource() (cli.InputSourceContext, error) {
	return &MapInputSource{file: "", valueMap: map[interface{}]interface{}{}}, nil
}
