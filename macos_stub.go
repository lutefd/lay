//go:build !darwin

package main

func ProtectWindow()              {}
func MakeWindowStealth()          {}
func SetAccessoryPolicy()         {}
func RegisterGlobalHotkey()       {}
func RegisterLocalKeyMonitor()    {}
func UnregisterGlobalHotkey()     {}
func UnregisterLocalKeyMonitor()  {}
func StartCapture(_ string) error { return nil }
func StopCapture()                {}
func RotateChunk(_ string) error  { return nil }
func ConsumeCaptureEvent() (string, bool) {
	return "", false
}
