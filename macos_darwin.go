//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>
#include <Carbon/Carbon.h>

static EventHotKeyRef  _hotKeyRef        = NULL;
static EventHandlerRef _hotKeyHandlerRef = NULL;

// toggleWindow checks the real window state and minimises or restores accordingly.
// Must be called on the main thread.
static void toggleWindow() {
    NSWindow *w = [NSApp mainWindow];
    if (!w) {
        for (NSWindow *win in [NSApp windows]) {
            w = win;
            break;
        }
    }
    if (!w) return;

    if ([w isMiniaturized]) {
        [w deminiaturize:nil];
        [NSApp activateIgnoringOtherApps:YES];
    } else if (![NSApp isHidden] && [w isVisible]) {
        [w miniaturize:nil];
    } else {
        [NSApp unhide:nil];
        [NSApp activateIgnoringOtherApps:YES];
        [w makeKeyAndOrderFront:nil];
    }
}

// hotkeyPressed is the Carbon event handler — fires system-wide regardless of
// which app is frontmost, without requiring Accessibility permissions.
static OSStatus hotkeyPressed(EventHandlerCallRef next, EventRef event, void *data) {
    dispatch_async(dispatch_get_main_queue(), ^{ toggleWindow(); });
    return noErr;
}

static void protectAllWindows() {
    dispatch_async(dispatch_get_main_queue(), ^{
        for (NSWindow *w in [NSApp windows]) {
            [w setSharingType:NSWindowSharingNone];
        }
    });
}

// setAccessoryPolicy hides the app from the macOS menu bar and Dock while
// still allowing it to show floating windows.
static void setAccessoryPolicy() {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    });
}

static void registerGlobalHotkey() {
    dispatch_async(dispatch_get_main_queue(), ^{
        // Install handler for kEventHotKeyPressed on the application target.
        EventTypeSpec spec = { kEventClassKeyboard, kEventHotKeyPressed };
        InstallApplicationEventHandler(hotkeyPressed, 1, &spec, NULL, &_hotKeyHandlerRef);

        // Register ⌘+Shift+L (kVK_ANSI_L == 37).
        EventHotKeyID hkID;
        hkID.signature = 'LAYO';
        hkID.id = 1;
        RegisterEventHotKey(kVK_ANSI_L, cmdKey | shiftKey, hkID,
                            GetApplicationEventTarget(), 0, &_hotKeyRef);
    });
}

static void unregisterGlobalHotkey() {
    dispatch_sync(dispatch_get_main_queue(), ^{
        if (_hotKeyRef)        { UnregisterEventHotKey(_hotKeyRef);     _hotKeyRef        = NULL; }
        if (_hotKeyHandlerRef) { RemoveEventHandler(_hotKeyHandlerRef);  _hotKeyHandlerRef = NULL; }
    });
}
*/
import "C"

// ProtectWindow makes the overlay invisible to screen capture.
func ProtectWindow() {
	C.protectAllWindows()
}

// SetAccessoryPolicy hides lay from the macOS menu bar and Dock.
func SetAccessoryPolicy() {
	C.setAccessoryPolicy()
}

// RegisterGlobalHotkey installs the ⌘+Shift+L Carbon hotkey (system-wide).
func RegisterGlobalHotkey() {
	C.registerGlobalHotkey()
}

// UnregisterGlobalHotkey removes the Carbon hotkey and its event handler.
func UnregisterGlobalHotkey() {
	C.unregisterGlobalHotkey()
}
