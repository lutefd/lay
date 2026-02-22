//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

static id _globalMonitor = nil;
static id _localMonitor  = nil;

// toggleWindow checks the real window state and minimises or restores accordingly.
// Must be called on the main thread (called inside dispatch_async blocks below).
static void toggleWindow() {
    // Find our content window. mainWindow is nil when the app is hidden,
    // so fall back to iterating all windows.
    NSWindow *w = [NSApp mainWindow];
    if (!w) {
        for (NSWindow *win in [NSApp windows]) {
            w = win;
            break;
        }
    }
    if (!w) return;

    if ([w isMiniaturized]) {
        // Restore from dock
        [w deminiaturize:nil];
        [NSApp activateIgnoringOtherApps:YES];
    } else if (![NSApp isHidden] && [w isVisible]) {
        // Visible → minimise to dock
        [w miniaturize:nil];
    } else {
        // Hidden (via NSApp hide) → unhide and bring forward
        [NSApp unhide:nil];
        [NSApp activateIgnoringOtherApps:YES];
        [w makeKeyAndOrderFront:nil];
    }
}

static void protectAllWindows() {
    dispatch_async(dispatch_get_main_queue(), ^{
        for (NSWindow *w in [NSApp windows]) {
            [w setSharingType:NSWindowSharingNone];
        }
    });
}

static void registerGlobalHotkey() {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSUInteger mask = NSEventModifierFlagCommand | NSEventModifierFlagShift;

        // Global monitor: fires when a different app is frontmost.
        _globalMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskKeyDown
                                                                handler:^(NSEvent *event) {
            if ((event.modifierFlags & mask) == mask && event.keyCode == 37) {
                dispatch_async(dispatch_get_main_queue(), ^{ toggleWindow(); });
            }
        }];

        // Local monitor: fires when this app is frontmost.
        // Returns nil to consume the event so it doesn't propagate to the webview.
        _localMonitor = [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskKeyDown
                                                              handler:^NSEvent *(NSEvent *event) {
            if ((event.modifierFlags & mask) == mask && event.keyCode == 37) {
                dispatch_async(dispatch_get_main_queue(), ^{ toggleWindow(); });
                return nil;
            }
            return event;
        }];
    });
}

static void unregisterGlobalHotkey() {
    dispatch_sync(dispatch_get_main_queue(), ^{
        if (_globalMonitor) { [NSEvent removeMonitor:_globalMonitor]; _globalMonitor = nil; }
        if (_localMonitor)  { [NSEvent removeMonitor:_localMonitor];  _localMonitor  = nil; }
    });
}
*/
import "C"

// ProtectWindow makes the overlay invisible to screen capture.
func ProtectWindow() {
	C.protectAllWindows()
}

// RegisterGlobalHotkey installs ⌘+Shift+L monitors (global + local).
func RegisterGlobalHotkey() {
	C.registerGlobalHotkey()
}

// UnregisterGlobalHotkey removes both event monitors.
func UnregisterGlobalHotkey() {
	C.unregisterGlobalHotkey()
}
