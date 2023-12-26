package cli

import (
	"os"
)

func init() {
	_ = os.Setenv("CLI_TEMPLATE_REPANIC", "1")
}
