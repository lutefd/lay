//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>
#include <Carbon/Carbon.h>

static EventHotKeyRef  _hotKeyRefs[5]    = { NULL };
static EventHandlerRef _hotKeyHandlerRef = NULL;
static id _localKeyMonitor = nil;

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

static void setWindowOpacity(CGFloat value) {
    NSWindow *w = [NSApp mainWindow];
    if (!w) {
        for (NSWindow *win in [NSApp windows]) {
            w = win;
            break;
        }
    }
    if (!w) return;

    [w setAlphaValue:value];
}

static void moveWindowEdge(NSString *direction) {
    NSWindow *w = [NSApp mainWindow];
    if (!w) {
        for (NSWindow *win in [NSApp windows]) {
            w = win;
            break;
        }
    }
    if (!w) return;

    NSScreen *screen = [w screen];
    if (!screen) {
        screen = [NSScreen mainScreen];
    }
    if (!screen) return;

    const CGFloat margin = 12.0;
    NSRect visible = [screen visibleFrame];
    NSRect frame = [w frame];

    CGFloat x = frame.origin.x;
    CGFloat y = frame.origin.y;

    if ([direction isEqualToString:@"left"]) {
        x = NSMinX(visible) + margin;
    } else if ([direction isEqualToString:@"right"]) {
        x = NSMaxX(visible) - frame.size.width - margin;
    } else if ([direction isEqualToString:@"up"]) {
        y = NSMaxY(visible) - frame.size.height - margin;
    } else if ([direction isEqualToString:@"down"]) {
        y = NSMinY(visible) + margin;
    }

    if (x < NSMinX(visible)) x = NSMinX(visible);
    if (y < NSMinY(visible)) y = NSMinY(visible);

    [w setFrameOrigin:NSMakePoint(x, y)];
}

// hotkeyPressed is the Carbon event handler — fires system-wide regardless of
// which app is frontmost, without requiring Accessibility permissions.
static OSStatus hotkeyPressed(EventHandlerCallRef next, EventRef event, void *data) {
    EventHotKeyID hkID;
    if (GetEventParameter(event, kEventParamDirectObject, typeEventHotKeyID, NULL,
                          sizeof(hkID), NULL, &hkID) != noErr) {
        return noErr;
    }

    dispatch_async(dispatch_get_main_queue(), ^{
        switch (hkID.id) {
            case 1: toggleWindow(); break;
            case 2: setWindowOpacity(1.0); break;
            case 3: setWindowOpacity(0.75); break;
            case 4: setWindowOpacity(0.5); break;
            case 5: setWindowOpacity(0.25); break;
            default: break;
        }
    });
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

        EventHotKeyID hkID;
        hkID.signature = 'LAYO';

        // Toggle window: ⌘+Shift+L (kVK_ANSI_L == 37).
        hkID.id = 1;
        RegisterEventHotKey(kVK_ANSI_L, cmdKey | shiftKey, hkID,
                            GetApplicationEventTarget(), 0, &_hotKeyRefs[0]);

        // Opacity: ⌘+⌥+1/2/3/4.
        hkID.id = 2;
        RegisterEventHotKey(kVK_ANSI_1, cmdKey | optionKey, hkID,
                            GetApplicationEventTarget(), 0, &_hotKeyRefs[1]);
        hkID.id = 3;
        RegisterEventHotKey(kVK_ANSI_2, cmdKey | optionKey, hkID,
                            GetApplicationEventTarget(), 0, &_hotKeyRefs[2]);
        hkID.id = 4;
        RegisterEventHotKey(kVK_ANSI_3, cmdKey | optionKey, hkID,
                            GetApplicationEventTarget(), 0, &_hotKeyRefs[3]);
        hkID.id = 5;
        RegisterEventHotKey(kVK_ANSI_4, cmdKey | optionKey, hkID,
                            GetApplicationEventTarget(), 0, &_hotKeyRefs[4]);
    });
}

static void registerLocalKeyMonitor() {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (_localKeyMonitor) return;
        _localKeyMonitor = [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskKeyDown handler:^NSEvent * (NSEvent *event) {
            if (!([event modifierFlags] & NSEventModifierFlagCommand)) {
                return event;
            }
            if (!([event modifierFlags] & NSEventModifierFlagShift)) {
                return event;
            }
            if ([event modifierFlags] & (NSEventModifierFlagOption | NSEventModifierFlagControl)) {
                return event;
            }

            switch ([event keyCode]) {
                case 123: moveWindowEdge(@"left"); return nil;  // Left arrow
                case 124: moveWindowEdge(@"right"); return nil; // Right arrow
                case 125: moveWindowEdge(@"down"); return nil;  // Down arrow
                case 126: moveWindowEdge(@"up"); return nil;    // Up arrow
                default: return event;
            }
        }];
    });
}

static void unregisterLocalKeyMonitor() {
    dispatch_sync(dispatch_get_main_queue(), ^{
        if (_localKeyMonitor) {
            [NSEvent removeMonitor:_localKeyMonitor];
            _localKeyMonitor = nil;
        }
    });
}

static void unregisterGlobalHotkey() {
    dispatch_sync(dispatch_get_main_queue(), ^{
        for (int i = 0; i < 5; i++) {
            if (_hotKeyRefs[i]) { UnregisterEventHotKey(_hotKeyRefs[i]); _hotKeyRefs[i] = NULL; }
        }
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

// RegisterLocalKeyMonitor installs focused-only ⌘+Arrow shortcuts for window positioning.
func RegisterLocalKeyMonitor() {
	C.registerLocalKeyMonitor()
}

// UnregisterGlobalHotkey removes the Carbon hotkey and its event handler.
func UnregisterGlobalHotkey() {
	C.unregisterGlobalHotkey()
}

// UnregisterLocalKeyMonitor removes the focused-only key monitor.
func UnregisterLocalKeyMonitor() {
	C.unregisterLocalKeyMonitor()
}
