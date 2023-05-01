package altsrc

import "github.com/urfave/cli/v2"

type GenericWithImport interface {
	cli.Generic

	FromJson([]byte) error
}
