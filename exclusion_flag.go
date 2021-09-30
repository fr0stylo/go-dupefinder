package main

import "fmt"

type exclusions struct {
	e map[string]bool
}

func (c *exclusions) String() string {
	return fmt.Sprint(c.e)
}

func (c *exclusions) Set(item string) error {
	if c.e == nil {
		c.e = map[string]bool{}
	}

	if _, ok := c.e[item]; !ok {
		c.e[item] = true
	}
	return nil
}
