//go:build mage

package main

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/sh"
)

// Build builds the ghamon binary.
func Build() {
	runV("go", "build", "-o", "ghamon", ".")
}

// Test runs the unit tests.
func Test() {
	runV("go", "test", "-v", "./...")
}

// Clean removes build artifacts.
func Clean() {
	runV("rm", "ghamon")
}

func runV(cmd string, args ...string) error {
	fmt.Println(cmd, strings.Join(args, " "))
	return sh.RunV(cmd, args...)
}
