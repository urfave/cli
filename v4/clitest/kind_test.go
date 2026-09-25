package clitest_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/urfave/cli/v4"
	"github.com/urfave/cli/v4/clitest"
)

// An application can define its own kind, give it a message, and have it
// matched, translated and mapped to an exit code like the built-in ones.
var errQuota = cli.NewKind("quota_exceeded", cli.Failure)

func TestApplicationKind(t *testing.T) {
	cat := &cli.Catalog{Tag: "en", Messages: map[string]cli.Text{
		"error.quota_exceeded": {Other: "quota exceeded for {value}"},
	}}
	err := fmt.Errorf("upload: %w", &cli.Error{Kind: errQuota, Value: "alice"})
	cmd := &cli.Command{Name: "app", Localizer: cat, Action: func(context.Context, *cli.Command) error {
		return err
	}}

	if !errors.Is(err, &cli.Error{Kind: errQuota}) {
		t.Error("errors.Is did not match the application kind")
	}
	r := clitest.Run(t.Context(), cmd)
	if r.Stderr != "app: quota exceeded for alice\n" {
		t.Errorf("stderr = %q", r.Stderr)
	}
	if r.Code != 1 {
		t.Errorf("exit code = %d, want 1", r.Code)
	}
}
