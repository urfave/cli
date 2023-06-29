package cli

import (
	_ "github.com/urfave/cli/v3/internal/translations"
	"golang.org/x/text/message"
)

var mprinter *message.Printer

func init() {
	// TBD add language detection logic and fallback to en-US
	mprinter = message.NewPrinter(message.MatchLanguage("en-US"))
}
