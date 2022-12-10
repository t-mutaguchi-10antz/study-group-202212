package main

import (
	"fmt"

	"golang.org/x/xerrors"
)

func main() {
	err := func() error {
		err := func() error {
			return xerrors.New("original error")
		}()
		if err != nil {
			return xerrors.Errorf("wrapped error: %w", err)
		}

		return nil
	}()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
