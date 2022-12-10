package main

import (
	"errors"
	"fmt"

	pkg_errors "github.com/pkg/errors"
)

func main() {
	err := func() error {
		err := func() error {
			return errors.New("original error")
		}()
		if err != nil {
			return pkg_errors.Wrap(err, "wrapped error")
		}

		return nil
	}()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
