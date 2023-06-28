package cli

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var mprinter *message.Printer

func init() {
	mprinter = message.NewPrinter(language.AmericanEnglish)
}
