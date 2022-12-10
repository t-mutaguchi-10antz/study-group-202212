package main

import (
	"errors"
	"fmt"
)

func main() {
	err := func() error {
		err := func() error {
			return errors.New("original error")
		}()
		if err != nil {
			return fmt.Errorf("wrapped error: %w", err)
		}

		return nil
	}()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
