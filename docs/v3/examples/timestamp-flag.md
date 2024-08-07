---
tags:
  - v3
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
	"context"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.TimestampFlag{
				Name: "meeting", 
				Config: cli.TimestampConfig{
					AvailableLayouts: []string{"2006-01-02T15:04:05"},
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("%s", cmd.Timestamp("meeting").String())
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
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
cmd := &cli.Command{
	Flags: []cli.Flag{
		&cli.TimestampFlag{
			Name: "meeting",
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				AvailableLayouts: []string{"2006-01-02T15:04:05"},
			},
		},
	},
}
```

(time.Local contains the system's local time zone.)

Side note: quotes may be necessary around the date depending on your layout (if
you have spaces for instance)
