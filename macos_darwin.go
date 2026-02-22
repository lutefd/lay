//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

// hotkeyFired is implemented in Go.
extern void hotkeyFired();

static id _hotkeyMonitor = nil;

// protectAllWindows must run on the main thread.
static void protectAllWindows() {
    dispatch_async(dispatch_get_main_queue(), ^{
        for (NSWindow *w in [NSApp windows]) {
            [w setSharingType:NSWindowSharingNone];
        }
    });
}

// registerGlobalHotkey must run on the main thread.
static void registerGlobalHotkey() {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSUInteger mask = NSEventModifierFlagCommand | NSEventModifierFlagShift;
        _hotkeyMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskKeyDown
                                                                handler:^(NSEvent *event) {
            if ((event.modifierFlags & mask) == mask && event.keyCode == 37) {
                hotkeyFired();
            }
        }];
    });
}

// unregisterGlobalHotkey must run on the main thread.
static void unregisterGlobalHotkey() {
    dispatch_sync(dispatch_get_main_queue(), ^{
        if (_hotkeyMonitor) {
            [NSEvent removeMonitor:_hotkeyMonitor];
            _hotkeyMonitor = nil;
        }
    });
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
// Safe to call from any goroutine — dispatches to main thread internally.
func ProtectWindow() {
	C.protectAllWindows()
}

// RegisterGlobalHotkey sets up ⌘+Shift+L global hotkey.
// Safe to call from any goroutine — dispatches to main thread internally.
func RegisterGlobalHotkey() {
	C.registerGlobalHotkey()
}

// UnregisterGlobalHotkey removes the global event monitor.
// Safe to call from any goroutine — dispatches to main thread internally.
func UnregisterGlobalHotkey() {
	C.unregisterGlobalHotkey()
}
