---
tags:
  - v2
search:
  boost: 2
---

Using the timestamp flag is simple. Please refer to
[`time.Parse`](https://golang.org/pkg/time/#example_Parse) to get possible
formats.

<!-- {
  "args": ["&#45;&#45;meeting", "2019-08-12T15:04:05"],
  "output": "2019\\-08\\-12 15\\:04\\:05 \\+0000 UTC"
} -->
```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.TimestampFlag{Name: "meeting", Layout: "2006-01-02T15:04:05"},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("%s", cCtx.Timestamp("meeting").String())
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
```

In this example the flag could be used like this:

```sh-session
$ myapp --meeting 2019-08-12T15:04:05
```

When the layout doesn't contain timezones, timestamp will render with UTC. To
change behavior, a default timezone can be provided with flag definition:

```go
package main

import (
	"log"
	"time"
	"os"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.TimestampFlag{Name: "meeting", Layout: "2006-01-02T15:04:05", Timezone: time.Local},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
```

(time.Local contains the system's local time zone.)

Side note: quotes may be necessary around the date depending on your layout (if
you have spaces for instance)
