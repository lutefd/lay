//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void protectAllWindows() {
    for (NSWindow *window in [NSApp windows]) {
        [window setSharingType:NSWindowSharingNone];
    }
}
*/
import "C"

// ProtectWindow sets NSWindowSharingNone on all app windows so the overlay
// is invisible to screen capture software (Google Meet, Zoom, etc.).
func ProtectWindow() {
	C.protectAllWindows()
}
