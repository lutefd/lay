//go:build !darwin

package main

func ProtectWindow()             {}
func SetAccessoryPolicy()        {}
func RegisterGlobalHotkey()      {}
func RegisterLocalKeyMonitor()   {}
func UnregisterGlobalHotkey()    {}
func UnregisterLocalKeyMonitor() {}
func StartCapture(_ string) error { return nil }
func StopCapture()                {}
