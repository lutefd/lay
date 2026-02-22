//go:build !darwin

package main

var hotkeyChannel = make(chan struct{}, 1)

func ProtectWindow()          {}
func RegisterGlobalHotkey()   {}
func UnregisterGlobalHotkey() {}
