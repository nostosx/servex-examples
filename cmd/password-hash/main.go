// Command password-hash reads one password from stdin and prints a PHC hash.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/nostosx/servex-contrib/auth/password"
	"github.com/nostosx/servex-examples/infrastructure/logging"
)

func main() {
	if err := run(); err != nil {
		logging.Report(err)
		os.Exit(1)
	}
}
func run() error {
	input, err := io.ReadAll(io.LimitReader(os.Stdin, 1025))
	if err != nil {
		return err
	}
	defer clear(input)
	if len(input) == 0 {
		return errors.New("password required on stdin; trailing newlines are significant")
	}
	hasher, err := password.New(password.DefaultParameters())
	if err != nil {
		return err
	}
	hash, err := hasher.Hash(context.Background(), input)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, hash)
	return err
}
