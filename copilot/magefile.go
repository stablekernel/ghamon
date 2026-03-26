//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Build compiles the ghamon binary (runs Vet first).
func Build() error {
	mg.Deps(Vet)
	return sh.Run("go", "build", "-o", "ghamon", ".")
}

// Test runs all unit tests with verbose output.
func Test() error {
	return sh.Run("go", "test", "-v", "-count=1", "./...")
}

// Vet runs go vet across all packages.
func Vet() error {
	return sh.Run("go", "vet", "./...")
}

// Clean removes the compiled binary.
func Clean() error {
	return sh.Rm("ghamon")
}
