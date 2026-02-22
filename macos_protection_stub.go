//go:build !darwin

package main

// ProtectWindow is a no-op on non-macOS platforms.
func ProtectWindow() {}
