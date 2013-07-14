package cli

import "fmt"
import "os"
import "text/tabwriter"

var HelpCommand = Command{
	Name:      "help",
	ShortName: "h",
	Usage:     "View help topics",
	Action:    ShowHelp,
}

var ShowHelp = func(name string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Printf("Usage: %v [global-options] COMMAND [command-options]\n\n", Name)
	if Commands != nil {
		fmt.Printf("The most commonly used %v commands are:\n", Name)
		for _, c := range Commands {
			fmt.Fprintln(w, "   "+c.Name+"\t"+c.Usage)
		}
		w.Flush()
	}
}
