package main

import (
	"fmt"

	"golang.org/x/sync/errgroup"
)

type cmd interface {
	run() error
}

type concurrent map[string]cmd

func (c concurrent) run() error {
	errGr := errgroup.Group{}

	for name, c := range c {
		name := name
		c := c
		errGr.Go(func() error {
			if err := c.run(); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			return nil
		})
	}

	return errGr.Wait()
}

type sequential map[string]cmd

func (s sequential) run() error {
	for name, c := range s {
		if err := c.run(); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

type cmdFunc func() error

func (f cmdFunc) run() error {
	return f()
}
