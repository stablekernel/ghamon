//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/sh"
)

// Build builds the ghamon binary.
func Build() {
	sh.RunV("go", "build", "-o", "ghamon", ".")
}

// Test runs the unit tests.
func Test() {
	sh.RunV("go", "test", "-v", "./...")
}

// Clean removes build artifacts.
func Clean() {
	os.Remove("ghamon")
}
