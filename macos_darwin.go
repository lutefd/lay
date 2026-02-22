//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// hotkeyFired is implemented in Go.
extern void hotkeyFired();

static id _hotkeyMonitor = nil;

// ProtectWindow sets NSWindowSharingNone on all windows so the overlay is
// invisible to screen capture (Google Meet, Zoom, etc.).
static void protectAllWindows() {
    for (NSWindow *w in [NSApp windows]) {
        [w setSharingType:NSWindowSharingNone];
    }
}

// registerGlobalHotkey installs a global NSEvent monitor for ⌘+Shift+L.
static void registerGlobalHotkey() {
    NSUInteger mask = NSEventModifierFlagCommand | NSEventModifierFlagShift;
    _hotkeyMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskKeyDown
                                                            handler:^(NSEvent *event) {
        if ((event.modifierFlags & mask) == mask && event.keyCode == 37) {
            hotkeyFired();
        }
    }];
}

// unregisterGlobalHotkey removes the event monitor.
static void unregisterGlobalHotkey() {
    if (_hotkeyMonitor) {
        [NSEvent removeMonitor:_hotkeyMonitor];
        _hotkeyMonitor = nil;
    }
}
*/
import "C"

var hotkeyChannel = make(chan struct{}, 1)

//export hotkeyFired
func hotkeyFired() {
	select {
	case hotkeyChannel <- struct{}{}:
	default:
	}
}

// ProtectWindow makes the overlay invisible to screen capture.
func ProtectWindow() {
	C.protectAllWindows()
}

// RegisterGlobalHotkey sets up ⌘+Shift+L global hotkey.
func RegisterGlobalHotkey() {
	C.registerGlobalHotkey()
}

// UnregisterGlobalHotkey removes the global event monitor.
func UnregisterGlobalHotkey() {
	C.unregisterGlobalHotkey()
}
