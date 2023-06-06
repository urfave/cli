// Package cli provides a minimal framework for creating and organizing command line
// Go applications. cli is designed to be easy to understand and write, the most simple
// cli application can be written as follows:
//
//	func main() {
//		(&cli.Command{}).Run(context.Background(), os.Args)
//	}
//
// Of course this application does not do much, so let's make this an actual application:
//
//	func main() {
//		cmd := &cli.Command{
//	  		Name: "greet",
//	  		Usage: "say a greeting",
//	  		Action: func(c *cli.Context) error {
//	  			fmt.Println("Greetings")
//	  			return nil
//	  		},
//		}
//
//		cmd.Run(context.Background(), os.Args)
//	}
package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func tracef(format string, a ...any) (int, error) {
	if os.Getenv("URFAVE_CLI_TRACE") != "on" {
		return 0, nil
	}

	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}

	_, file, line, _ := runtime.Caller(1)

	return fmt.Fprintf(os.Stderr, "# URFAVE CLI TRACE "+file+":"+fmt.Sprintf("%v", line)+" ---> "+format, a...)
}
