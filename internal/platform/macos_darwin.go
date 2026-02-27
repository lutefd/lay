//go:build darwin

package platform

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon -framework ScreenCaptureKit -framework AVFoundation -framework CoreMedia -framework AudioToolbox -framework CoreAudio

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>
#import <objc/runtime.h>
#include <Carbon/Carbon.h>
#import <AVFoundation/AVFoundation.h>
#import <AudioToolbox/AudioToolbox.h>
#import <CoreAudio/CoreAudio.h>
#import <CoreMedia/CoreMedia.h>
#import <ScreenCaptureKit/ScreenCaptureKit.h>

static EventHotKeyRef  _hotKeyRefs[5]    = { NULL };
static EventHandlerRef _hotKeyHandlerRef = NULL;
static id _localKeyMonitor = nil;
static Class _layPanelClass = nil;
static CGFloat _savedAlpha = 1.0;

// makeWindowStealth converts the Wails NSWindow into a non-activating NSPanel
// at runtime so that clicking or toggling lay never triggers a browser blur event.
// NSPanel and NSWindow share the same ivar layout on modern macOS, making the
// object_setClass swap safe. Must be called after the window is fully initialised.
static void makeWindowStealth() {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *w = [NSApp mainWindow];
        if (!w) {
            for (NSWindow *win in [NSApp windows]) { w = win; break; }
        }
        if (!w) return;

        // Build the subclass once; subsequent calls are no-ops.
        if (!_layPanelClass) {
            _layPanelClass = objc_allocateClassPair([NSPanel class], "LayGhostPanel", 0);
            if (!_layPanelClass) {
                _layPanelClass = objc_getClass("LayGhostPanel"); // already registered
            } else {
                // Always accept keyboard input so text fields in lay still work.
                class_addMethod(_layPanelClass, @selector(canBecomeKey),
                    imp_implementationWithBlock(^BOOL(id _){ return YES; }), "c@:");
                // Never claim main-window status.
                class_addMethod(_layPanelClass, @selector(canBecomeMain),
                    imp_implementationWithBlock(^BOOL(id _){ return NO; }), "c@:");
                objc_registerClassPair(_layPanelClass);
            }
        }
        if (!_layPanelClass) return;

        // Re-class the live Wails window to our non-activating panel subclass.
        object_setClass(w, _layPanelClass);

        // NSWindowStyleMaskNonactivatingPanel (1<<7 = 128): clicking or ordering
        // this window to front will NOT activate the lay application, so the
        // browser (or any other app) never receives a focus-loss / blur event.
        NSWindowStyleMask mask = [w styleMask];
        if (!(mask & NSWindowStyleMaskNonactivatingPanel)) {
            [w setStyleMask:mask | NSWindowStyleMaskNonactivatingPanel];
        }

        NSPanel *p = (NSPanel *)w;
        [p setHidesOnDeactivate:NO];  // stay visible when lay is in background

        // Hidden from Mission Control, Exposé, cmd+`, and Spaces cycling.
        [w setCollectionBehavior:
            NSWindowCollectionBehaviorCanJoinAllSpaces  |
            NSWindowCollectionBehaviorStationary        |
            NSWindowCollectionBehaviorIgnoresCycle      |
            NSWindowCollectionBehaviorFullScreenAuxiliary];
    });
}

// toggleWindow soft-hides or soft-shows the window using alpha + ignoresMouseEvents
// instead of orderOut/orderFront. This keeps the WKWebView in the window hierarchy
// so it never resets its HTML focus or first-responder state — which is why text
// fields work immediately after showing again. orderOut would tear that state down.
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

    BOOL hidden = ![w isVisible] || [w alphaValue] < 0.01;

    if (!hidden) {
        // Soft-hide: go transparent and pass mouse events through.
        // The window stays ordered so WKWebView keeps all its internal state.
        _savedAlpha = [w alphaValue] > 0.05 ? [w alphaValue] : 1.0;
        [w setAlphaValue:0.0];
        [w setIgnoresMouseEvents:YES];
    } else {
        // Soft-show: restore opacity and re-enable mouse events.
        // orderFrontRegardless brings it front without activating the lay app.
        [w setIgnoresMouseEvents:NO];
        [w setAlphaValue:_savedAlpha];
        [w orderFrontRegardless];
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

// ──────────────────────────────────────────────────────────────────────────────
// Audio capture — ScreenCaptureKit (system audio) + AVAudioEngine (mic)
// ──────────────────────────────────────────────────────────────────────────────

static AVAudioEngine      *_audioEngine    = nil;
static AVAudioFile        *_micAudioFile   = nil;
static AVAudioFile        *_chunkMicFile   = nil;  // current live mic chunk
static AVAudioFormat      *_micRecordingFmt = nil; // hardware format for chunk creation
static id                  _scStream       = nil;  // SCStream*
static id                  _scDelegate     = nil;  // LaySCStreamDelegate*
static ExtAudioFileRef     _sysExtFile       = NULL; // full system.caf
static ExtAudioFileRef     _chunkSysExtFile  = NULL; // current live sys chunk
static AudioStreamBasicDescription _sysNativeAsbd   = {0}; // format SCStream delivers
static AudioStreamBasicDescription _sysNormFileFmt  = {0}; // normalised file format
static BOOL                _sysNativeAsbdValid = NO;  // true after first callback
static dispatch_queue_t    _audioWriteQ    = nil;
static char               *_sysFilePath    = NULL; // malloc'd, owned here
static BOOL                _sysFileReady   = NO;   // file created on first callback
static BOOL                _isRecording    = NO;
static char               *_captureEvent   = NULL; // malloc'd, consumed by Go
static AudioDeviceID       _savedDefaultOutput = kAudioObjectUnknown;
static AudioDeviceID       _aggOutputDevice    = kAudioObjectUnknown;
static id                  _engineConfigObserver = nil;

API_AVAILABLE(macos(13.0))
@interface LaySCStreamDelegate : NSObject <SCStreamOutput, SCStreamDelegate>
@end

static void setCaptureEvent(NSString *message) {
    if (_captureEvent) {
        free(_captureEvent);
        _captureEvent = NULL;
    }
    const char *utf8 = (message && message.length > 0)
        ? message.UTF8String
        : "system audio capture stopped";
    _captureEvent = strdup(utf8);
}

static char *consumeCaptureEvent(void) {
    if (!_captureEvent) return NULL;
    char *out = _captureEvent;
    _captureEvent = NULL;
    return out;
}

API_AVAILABLE(macos(13.0))
@implementation LaySCStreamDelegate

- (void)stream:(SCStream *)stream
    didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer
    ofType:(SCStreamOutputType)type
{
    if (type != SCStreamOutputTypeAudio) return;

    // On the first callback we learn the actual format from the buffer itself.
    if (!_sysFileReady && _sysFilePath) {
        CMFormatDescriptionRef fmtDesc = CMSampleBufferGetFormatDescription(sampleBuffer);
        const AudioStreamBasicDescription *asbd =
            CMAudioFormatDescriptionGetStreamBasicDescription(fmtDesc);
        if (!asbd) return;  // retry next callback

        // Normalise to packed interleaved float32 so CAF creation always succeeds.
        // The client format (what SCStream delivers) may be non-interleaved; we let
        // ExtAudioFile's AudioConverter handle the layout conversion automatically.
        AudioStreamBasicDescription fileFmt = {0};
        fileFmt.mSampleRate       = asbd->mSampleRate;
        fileFmt.mFormatID         = kAudioFormatLinearPCM;
        fileFmt.mFormatFlags      = kAudioFormatFlagIsFloat | kAudioFormatFlagIsPacked;
        fileFmt.mBitsPerChannel   = 32;
        fileFmt.mChannelsPerFrame = asbd->mChannelsPerFrame;
        fileFmt.mBytesPerFrame    = 4 * asbd->mChannelsPerFrame;
        fileFmt.mFramesPerPacket  = 1;
        fileFmt.mBytesPerPacket   = fileFmt.mBytesPerFrame;

        NSURL *url = [NSURL fileURLWithPath:
            [NSString stringWithUTF8String:_sysFilePath]];
        OSStatus st = ExtAudioFileCreateWithURL((__bridge CFURLRef)url, kAudioFileCAFType,
            &fileFmt, NULL, kAudioFileFlags_EraseFile, &_sysExtFile);
        if (st != noErr) {
            NSLog(@"[lay] ExtAudioFileCreateWithURL failed: %d (path: %@)", (int)st, url.path);
            return;  // retry next callback
        }

        // Tell ExtAudioFile the format SCStream actually delivers so it can convert.
        st = ExtAudioFileSetProperty(_sysExtFile,
            kExtAudioFileProperty_ClientDataFormat, sizeof(*asbd), asbd);
        if (st != noErr) {
            NSLog(@"[lay] ExtAudioFileSetProperty(ClientDataFormat) failed: %d", (int)st);
            ExtAudioFileDispose(_sysExtFile);
            _sysExtFile = NULL;
            return;  // retry next callback
        }

        // Store formats so rotateChunk can open new sys chunk files later.
        _sysNativeAsbd    = *asbd;
        _sysNormFileFmt   = fileFmt;
        _sysNativeAsbdValid = YES;

        // Open the first live sys chunk (chunk-sys-0.caf alongside chunk-0.caf).
        NSString *dir = url.path.stringByDeletingLastPathComponent;
        NSURL *chunk0URL = [NSURL fileURLWithPath:
            [dir stringByAppendingPathComponent:@"chunk-sys-0.caf"]];
        OSStatus cst = ExtAudioFileCreateWithURL((__bridge CFURLRef)chunk0URL,
            kAudioFileCAFType, &fileFmt, NULL, kAudioFileFlags_EraseFile, &_chunkSysExtFile);
        if (cst == noErr) {
            ExtAudioFileSetProperty(_chunkSysExtFile,
                kExtAudioFileProperty_ClientDataFormat, sizeof(*asbd), asbd);
        }

        _sysFileReady = YES;
    }

    if (!_sysExtFile) return;

    CMItemCount numSamples = CMSampleBufferGetNumSamples(sampleBuffer);
    if (numSamples == 0) return;

    size_t ablSize = 0;
    CMSampleBufferGetAudioBufferListWithRetainedBlockBuffer(
        sampleBuffer, &ablSize, NULL, 0, NULL, NULL, 0, NULL);

    AudioBufferList *abl = (AudioBufferList *)malloc(ablSize);
    CMBlockBufferRef blockBuf = NULL;
    OSStatus st = CMSampleBufferGetAudioBufferListWithRetainedBlockBuffer(
        sampleBuffer, &ablSize, abl, ablSize, NULL, NULL,
        kCMSampleBufferFlag_AudioBufferList_Assure16ByteAlignment, &blockBuf);

    if (st == noErr) {
        ExtAudioFileWrite(_sysExtFile, (UInt32)numSamples, abl);
        if (_chunkSysExtFile) {
            ExtAudioFileWrite(_chunkSysExtFile, (UInt32)numSamples, abl);
        }
    }

    free(abl);
    if (blockBuf) CFRelease(blockBuf);
}

- (void)stream:(SCStream *)stream didStopWithError:(NSError *)error {
    NSString *msg = @"System audio capture stopped.";
    if (error.localizedDescription.length > 0) {
        msg = [NSString stringWithFormat:@"System audio capture stopped: %@",
               error.localizedDescription];
    }
    setCaptureEvent(msg);

    if (_sysExtFile) {
        ExtAudioFileDispose(_sysExtFile);
        _sysExtFile = NULL;
    }
    _sysFileReady = NO;
    _scStream   = nil;
    _scDelegate = nil;
}

@end

// rotateChunk finalizes the current live mic chunk and opens a new one.
// Returns 0 on success, -1 on error.
static int rotateChunk(const char *newMicPath) {
    if (!_micRecordingFmt) return -1;
    _chunkMicFile = nil; // ARC flushes old mic chunk to disk
    NSURL *micURL = [NSURL fileURLWithPath:[NSString stringWithUTF8String:newMicPath]];
    NSError *err = nil;
    _chunkMicFile = [[AVAudioFile alloc] initForWriting:micURL
                                              settings:_micRecordingFmt.settings
                                                 error:&err];
    if (err) return -1;

    // Rotate the matching sys chunk: chunk-N.caf → chunk-sys-N.caf in same dir.
    if (_sysNativeAsbdValid) {
        if (_chunkSysExtFile) {
            ExtAudioFileDispose(_chunkSysExtFile); // flushes previous chunk
            _chunkSysExtFile = NULL;
        }
        NSString *mp   = [NSString stringWithUTF8String:newMicPath];
        NSString *name = [mp.lastPathComponent
            stringByReplacingOccurrencesOfString:@"chunk-" withString:@"chunk-sys-"];
        NSURL *sysURL  = [NSURL fileURLWithPath:
            [mp.stringByDeletingLastPathComponent stringByAppendingPathComponent:name]];
        OSStatus st = ExtAudioFileCreateWithURL((__bridge CFURLRef)sysURL,
            kAudioFileCAFType, &_sysNormFileFmt, NULL, kAudioFileFlags_EraseFile, &_chunkSysExtFile);
        if (st == noErr) {
            ExtAudioFileSetProperty(_chunkSysExtFile,
                kExtAudioFileProperty_ClientDataFormat, sizeof(_sysNativeAsbd), &_sysNativeAsbd);
        }
    }
    return 0;
}

// builtInInputDevice returns the AudioDeviceID of the first built-in device
// that has at least one input channel, or kAudioObjectUnknown if none found.
// We use this to avoid handing AVAudioEngine a Bluetooth input device, which
// would trigger the HFP profile switch and break SCStream's audio tap.
static AudioDeviceID builtInInputDevice(void) {
    AudioObjectPropertyAddress addr = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    UInt32 sz = 0;
    if (AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &addr, 0, NULL, &sz) != noErr || sz == 0)
        return kAudioObjectUnknown;

    UInt32 n = sz / sizeof(AudioDeviceID);
    AudioDeviceID *devs = (AudioDeviceID *)malloc(sz);
    if (AudioObjectGetPropertyData(kAudioObjectSystemObject, &addr, 0, NULL, &sz, devs) != noErr) {
        free(devs);
        return kAudioObjectUnknown;
    }

    AudioDeviceID found = kAudioObjectUnknown;
    for (UInt32 i = 0; i < n && found == kAudioObjectUnknown; i++) {
        UInt32 transport = 0, tsz = sizeof(transport);
        AudioObjectPropertyAddress ta = {
            kAudioDevicePropertyTransportType,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };
        if (AudioObjectGetPropertyData(devs[i], &ta, 0, NULL, &tsz, &transport) != noErr) continue;
        if (transport != kAudioDeviceTransportTypeBuiltIn) continue;

        AudioObjectPropertyAddress ia = {
            kAudioDevicePropertyStreamConfiguration,
            kAudioObjectPropertyScopeInput,
            kAudioObjectPropertyElementMain
        };
        UInt32 ablSz = 0;
        if (AudioObjectGetPropertyDataSize(devs[i], &ia, 0, NULL, &ablSz) != noErr || ablSz < sizeof(AudioBufferList)) continue;
        AudioBufferList *abl = (AudioBufferList *)malloc(ablSz);
        AudioObjectGetPropertyData(devs[i], &ia, 0, NULL, &ablSz, abl);
        BOOL hasInput = abl->mNumberBuffers > 0 && abl->mBuffers[0].mNumberChannels > 0;
        free(abl);
        if (hasInput) found = devs[i];
    }
    free(devs);
    return found;
}

// ──────────────────────────────────────────────────────────────────────────────
// Multi-output aggregate device — makes SCStream work when Bluetooth is output
//
// SCStream's audio tap sits on the display's Core Audio render path. When
// Bluetooth (AirPods) is the default output, the render path bypasses the
// display entirely, so SCStream receives no audio callbacks.
//
// The fix: create a "stacked" (multi-output) aggregate device that combines
// the Bluetooth device with the built-in speakers.  Both devices then play the
// same audio simultaneously.  SCStream can tap the built-in path while the
// user continues to hear through their headphones.
//
// Side-effect: the MacBook speakers will also play audio while recording.
// ──────────────────────────────────────────────────────────────────────────────

// findCoreAudioPlugin returns the AudioObjectID for the CoreAudio HAL plugin
// (bundle ID "com.apple.audio.CoreAudio"), which owns the aggregate-device API.
static AudioObjectID findCoreAudioPlugin(void) {
    AudioObjectPropertyAddress addr = {
        kAudioHardwarePropertyPlugInList,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    UInt32 sz = 0;
    if (AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &addr, 0, NULL, &sz) != noErr || sz == 0)
        return kAudioObjectUnknown;

    UInt32 n = sz / sizeof(AudioObjectID);
    AudioObjectID *plugins = (AudioObjectID *)malloc(sz);
    if (AudioObjectGetPropertyData(kAudioObjectSystemObject, &addr, 0, NULL, &sz, plugins) != noErr) {
        free(plugins);
        return kAudioObjectUnknown;
    }

    AudioObjectID found = kAudioObjectUnknown;
    for (UInt32 i = 0; i < n && found == kAudioObjectUnknown; i++) {
        CFStringRef bundleID = NULL;
        UInt32 bsz = sizeof(bundleID);
        AudioObjectPropertyAddress ba = {
            kAudioPlugInPropertyBundleID,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };
        if (AudioObjectGetPropertyData(plugins[i], &ba, 0, NULL, &bsz, &bundleID) != noErr) continue;
        if (bundleID) {
            if (CFStringCompare(bundleID, CFSTR("com.apple.audio.CoreAudio"), 0) == kCFCompareEqualTo)
                found = plugins[i];
            CFRelease(bundleID);
        }
    }
    free(plugins);
    return found;
}

// getDeviceUID returns the UID string for a Core Audio device.
static NSString *getDeviceUID(AudioDeviceID deviceID) {
    CFStringRef uid = NULL;
    UInt32 sz = sizeof(uid);
    AudioObjectPropertyAddress addr = {
        kAudioDevicePropertyDeviceUID,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    if (AudioObjectGetPropertyData(deviceID, &addr, 0, NULL, &sz, &uid) != noErr || !uid)
        return nil;
    NSString *result = [NSString stringWithString:(__bridge NSString *)uid];
    CFRelease(uid);
    return result;
}

// builtInOutputDevice finds the built-in audio device that has output channels.
static AudioDeviceID builtInOutputDevice(void) {
    AudioObjectPropertyAddress addr = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    UInt32 sz = 0;
    if (AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &addr, 0, NULL, &sz) != noErr || sz == 0)
        return kAudioObjectUnknown;

    UInt32 n = sz / sizeof(AudioDeviceID);
    AudioDeviceID *devs = (AudioDeviceID *)malloc(sz);
    if (AudioObjectGetPropertyData(kAudioObjectSystemObject, &addr, 0, NULL, &sz, devs) != noErr) {
        free(devs);
        return kAudioObjectUnknown;
    }

    AudioDeviceID found = kAudioObjectUnknown;
    for (UInt32 i = 0; i < n && found == kAudioObjectUnknown; i++) {
        UInt32 transport = 0, tsz = sizeof(transport);
        AudioObjectPropertyAddress ta = {
            kAudioDevicePropertyTransportType,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };
        if (AudioObjectGetPropertyData(devs[i], &ta, 0, NULL, &tsz, &transport) != noErr) continue;
        if (transport != kAudioDeviceTransportTypeBuiltIn) continue;

        AudioObjectPropertyAddress oa = {
            kAudioDevicePropertyStreamConfiguration,
            kAudioObjectPropertyScopeOutput,
            kAudioObjectPropertyElementMain
        };
        UInt32 ablSz = 0;
        if (AudioObjectGetPropertyDataSize(devs[i], &oa, 0, NULL, &ablSz) != noErr || ablSz < sizeof(AudioBufferList)) continue;
        AudioBufferList *abl = (AudioBufferList *)malloc(ablSz);
        AudioObjectGetPropertyData(devs[i], &oa, 0, NULL, &ablSz, abl);
        BOOL hasOutput = abl->mNumberBuffers > 0 && abl->mBuffers[0].mNumberChannels > 0;
        free(abl);
        if (hasOutput) found = devs[i];
    }
    free(devs);
    return found;
}

// createStackedOutputForCapture detects a Bluetooth default output and, if found,
// creates a stacked (multi-output) aggregate device combining it with the
// built-in speakers, then sets that aggregate as the system default output.
// Returns YES if a workaround was installed; the caller must call
// restoreOutputDevice() when done.
static BOOL createStackedOutputForCapture(void) {
    // Get current default output
    AudioDeviceID currentOut = kAudioObjectUnknown;
    UInt32 sz = sizeof(AudioDeviceID);
    AudioObjectPropertyAddress outAddr = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    AudioObjectGetPropertyData(kAudioObjectSystemObject, &outAddr, 0, NULL, &sz, &currentOut);
    if (currentOut == kAudioObjectUnknown) return NO;

    // Only apply when the default output is Bluetooth
    UInt32 transport = 0, tsz = sizeof(transport);
    AudioObjectPropertyAddress ta = {
        kAudioDevicePropertyTransportType,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    AudioObjectGetPropertyData(currentOut, &ta, 0, NULL, &tsz, &transport);
    if (transport != kAudioDeviceTransportTypeBluetooth &&
        transport != kAudioDeviceTransportTypeBluetoothLE) {
        return NO;
    }

    AudioDeviceID builtIn = builtInOutputDevice();
    if (builtIn == kAudioObjectUnknown) {
        NSLog(@"[lay] Bluetooth output but no built-in speakers found; system audio capture may be empty");
        return NO;
    }

    NSString *btUID      = getDeviceUID(currentOut);
    NSString *builtInUID = getDeviceUID(builtIn);
    if (!btUID || !builtInUID) return NO;

    AudioObjectID coreAudioPlugin = findCoreAudioPlugin();
    if (coreAudioPlugin == kAudioObjectUnknown) {
        NSLog(@"[lay] CoreAudio plugin not found; cannot create multi-output device");
        return NO;
    }

    // Build a stacked aggregate: Bluetooth (primary/master) + built-in.
    // "Stacked" means the same mix is sent to every sub-device simultaneously.
    NSDictionary *desc = @{
        @kAudioAggregateDeviceNameKey:           @"Lay Capture",
        @kAudioAggregateDeviceUIDKey:            @"com.lay.captureoutput.v1",
        @kAudioAggregateDeviceSubDeviceListKey:  @[
            @{ @kAudioSubDeviceUIDKey: btUID      },
            @{ @kAudioSubDeviceUIDKey: builtInUID }
        ],
        @kAudioAggregateDeviceMasterSubDeviceKey: btUID,
        @kAudioAggregateDeviceIsStackedKey:       @YES,
        @kAudioAggregateDeviceIsPrivateKey:       @YES,  // hide from Audio MIDI Setup
    };
    CFDictionaryRef descCF = (__bridge CFDictionaryRef)desc;

    AudioObjectPropertyAddress createAddr = {
        kAudioPlugInCreateAggregateDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    AudioDeviceID newDevice = kAudioObjectUnknown;
    UInt32 newDevSz = sizeof(newDevice);
    OSStatus st = AudioObjectGetPropertyData(
        coreAudioPlugin, &createAddr,
        sizeof(descCF), &descCF,
        &newDevSz, &newDevice
    );
    if (st != noErr || newDevice == kAudioObjectUnknown) {
        NSLog(@"[lay] aggregate device creation failed: %d", (int)st);
        return NO;
    }

    // Set as default output
    st = AudioObjectSetPropertyData(kAudioObjectSystemObject, &outAddr,
        0, NULL, sizeof(newDevice), &newDevice);
    if (st != noErr) {
        NSLog(@"[lay] failed to set aggregate as default output: %d", (int)st);
        // Destroy the device we just created
        AudioObjectPropertyAddress destroyAddr = {
            kAudioPlugInDestroyAggregateDevice,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };
        AudioObjectSetPropertyData(coreAudioPlugin, &destroyAddr,
            0, NULL, sizeof(newDevice), &newDevice);
        return NO;
    }

    _savedDefaultOutput = currentOut;
    _aggOutputDevice    = newDevice;
    return YES;
}

// restoreOutputDevice tears down the aggregate device and restores the original
// default output.  Safe to call even if createStackedOutputForCapture() was not
// called or returned NO.
static void restoreOutputDevice(void) {
    if (_aggOutputDevice == kAudioObjectUnknown) return;

    if (_savedDefaultOutput != kAudioObjectUnknown) {
        AudioObjectPropertyAddress outAddr = {
            kAudioHardwarePropertyDefaultOutputDevice,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };
        AudioObjectSetPropertyData(kAudioObjectSystemObject, &outAddr,
            0, NULL, sizeof(_savedDefaultOutput), &_savedDefaultOutput);
    }

    AudioObjectID plugin = findCoreAudioPlugin();
    if (plugin != kAudioObjectUnknown) {
        AudioObjectPropertyAddress destroyAddr = {
            kAudioPlugInDestroyAggregateDevice,
            kAudioObjectPropertyScopeGlobal,
            kAudioObjectPropertyElementMain
        };
        AudioObjectSetPropertyData(plugin, &destroyAddr,
            0, NULL, sizeof(_aggOutputDevice), &_aggOutputDevice);
    }

    _savedDefaultOutput = kAudioObjectUnknown;
    _aggOutputDevice    = kAudioObjectUnknown;
}

// restartMicCapture re-pins AVAudioEngine to the built-in mic and restarts it
// after an AVAudioEngineConfigurationChangeNotification.  macOS automatically
// stops the engine and removes all taps before posting that notification, so
// we only need to re-install the tap and restart.  Called when the user plugs
// in or unplugs headphones while a recording is in progress.
static void restartMicCapture(void) {
    if (!_audioEngine || !_isRecording || !_micAudioFile) return;

    AVAudioInputNode *inputNode = [_audioEngine inputNode];

    // Re-pin to the built-in mic so switching to headphones doesn't activate
    // the HFP profile, which would degrade sample rate for the whole session.
    AudioDeviceID builtIn = builtInInputDevice();
    if (builtIn != kAudioObjectUnknown) {
        AudioUnitSetProperty(inputNode.audioUnit,
            kAudioOutputUnitProperty_CurrentDevice,
            kAudioUnitScope_Global, 0,
            &builtIn, sizeof(builtIn));
    }

    _micRecordingFmt = [inputNode outputFormatForBus:0];

    // Re-install the tap — system has already removed it.
    [inputNode installTapOnBus:0 bufferSize:4096 format:_micAudioFile.processingFormat
                         block:^(AVAudioPCMBuffer *buf, AVAudioTime *when) {
        if (_micAudioFile) [_micAudioFile writeFromBuffer:buf error:nil];
        if (_chunkMicFile) [_chunkMicFile writeFromBuffer:buf error:nil];
    }];

    NSError *engErr = nil;
    [_audioEngine startAndReturnError:&engErr];
    if (engErr) {
        NSLog(@"[lay] mic restart after device change failed: %@", engErr);
    }
}

// startCapture starts both mic and system audio capture into outDir.
// Returns NULL on success, or a malloc'd C string with the error message.
static const char *startCapture(const char *outDir) {
    if (_isRecording) return NULL;

    NSString *dir = [NSString stringWithUTF8String:outDir];

    // ── Configure output device before starting AVAudioEngine ────────────────
    //
    // If Bluetooth is the default output, createStackedOutputForCapture()
    // creates a stacked aggregate device and sets it as the system default.
    // That triggers an AVAudioEngineConfigurationChange notification which
    // silently stops any already-running AVAudioEngine instance — killing
    // the mic tap while SCStream keeps going. Doing this first ensures the
    // hardware is stable before the engine starts so no mid-recording
    // disruption occurs. This is a no-op when Bluetooth is not in use.
    if (@available(macOS 13.0, *)) {
        createStackedOutputForCapture();
    }

    // ── Microphone via AVAudioEngine ─────────────────────────────────────────

    _audioEngine = [[AVAudioEngine alloc] init];
    AVAudioInputNode *inputNode = [_audioEngine inputNode];

    // Force the I/O unit to use the built-in microphone.
    //
    // When a Bluetooth device (AirPods) is the default input, AVAudioEngine
    // activates the HFP profile on start. HFP reconfigures the entire Core
    // Audio HAL (drops to 8–16 kHz, mono) BEFORE SCStream is set up, which
    // causes SCStream to deliver zero audio callbacks even though it reports
    // no error. Pinning to the built-in mic keeps the HAL in normal A2DP
    // mode so SCStream can tap system audio as expected.
    AudioDeviceID builtIn = builtInInputDevice();
    if (builtIn != kAudioObjectUnknown) {
        AudioUnitSetProperty(inputNode.audioUnit,
            kAudioOutputUnitProperty_CurrentDevice,
            kAudioUnitScope_Global, 0,
            &builtIn, sizeof(builtIn));
    }

    // Re-query format after the device override so the file and tap use the
    // correct (built-in mic) format rather than whatever was cached earlier.
    AVAudioFormat *micFmt = [inputNode outputFormatForBus:0];

    // Use native mic format for the CAF file so the processingFormat matches
    // the tap buffer format exactly — no silent format-mismatch writes.
    NSURL *micURL = [NSURL fileURLWithPath:
        [dir stringByAppendingPathComponent:@"mic.caf"]];
    NSError *micErr = nil;
    _micAudioFile = [[AVAudioFile alloc] initForWriting:micURL
                                               settings:micFmt.settings
                                                  error:&micErr];
    if (micErr) {
        _audioEngine = nil;
        return strdup(micErr.localizedDescription.UTF8String);
    }

    // Store format for live chunk file creation and open the first chunk.
    _micRecordingFmt = micFmt;
    NSURL *chunk0URL = [NSURL fileURLWithPath:
        [dir stringByAppendingPathComponent:@"chunk-0.caf"]];
    _chunkMicFile = [[AVAudioFile alloc] initForWriting:chunk0URL
                                              settings:micFmt.settings
                                                 error:nil];

    // Install tap in the file's processingFormat so AVAudioEngine converts
    // from the hardware format for us before we write.
    AVAudioFormat *tapFmt = _micAudioFile.processingFormat;
    [inputNode installTapOnBus:0 bufferSize:4096 format:tapFmt
                         block:^(AVAudioPCMBuffer *buf, AVAudioTime *when) {
        if (_micAudioFile) [_micAudioFile writeFromBuffer:buf error:nil];
        if (_chunkMicFile) [_chunkMicFile writeFromBuffer:buf error:nil];
    }];

    NSError *engErr = nil;
    [_audioEngine startAndReturnError:&engErr];
    if (engErr) {
        [inputNode removeTapOnBus:0];
        _audioEngine = nil;
        _micAudioFile = nil;
        return strdup(engErr.localizedDescription.UTF8String);
    }

    // Watch for hardware changes (e.g. plugging in / unplugging headphones).
    // AVAudioEngine stops itself and removes all taps before posting this
    // notification; restartMicCapture() re-installs the tap and restarts so
    // recording continues uninterrupted.
    _engineConfigObserver = [[NSNotificationCenter defaultCenter]
        addObserverForName:AVAudioEngineConfigurationChangeNotification
                    object:_audioEngine
                     queue:[NSOperationQueue mainQueue]
                usingBlock:^(NSNotification *note) {
        restartMicCapture();
    }];

    // ── System audio via ScreenCaptureKit (macOS 13+) ────────────────────────

    if (@available(macOS 13.0, *)) {
        _audioWriteQ = dispatch_queue_create("com.lay.audiowrite", DISPATCH_QUEUE_SERIAL);

        // Store the output path; the file is created on the first sample callback
        // so we use the actual ASBD from the stream instead of guessing.
        NSString *sysPath = [dir stringByAppendingPathComponent:@"system.caf"];
        if (_sysFilePath) { free(_sysFilePath); }
        _sysFilePath  = strdup(sysPath.UTF8String);
        _sysFileReady = NO;

        dispatch_semaphore_t sem = dispatch_semaphore_create(0);
        __block const char *scErr = NULL;

        [SCShareableContent getShareableContentWithCompletionHandler:^(
            SCShareableContent *content, NSError *error)
        {
            if (error || content.displays.count == 0) {
                scErr = error
                    ? strdup(error.localizedDescription.UTF8String)
                    : strdup("no displays found for system audio capture");
                dispatch_semaphore_signal(sem);
                return;
            }

            SCDisplay *display = content.displays[0];
            SCContentFilter *filter =
                [[SCContentFilter alloc] initWithDisplay:display excludingWindows:@[]];

            SCStreamConfiguration *cfg = [[SCStreamConfiguration alloc] init];
            cfg.capturesAudio        = YES;
            // Do NOT force sampleRate/channelCount: when Bluetooth (AirPods) is the
            // output device the HAL switches to HFP at 8 or 16 kHz. Demanding 48000
            // causes the stream to start without errors but deliver no callbacks.
            // Leaving these unset lets SCStream match whatever the HAL is currently at.
            if (@available(macOS 14.2, *)) {
                cfg.excludesCurrentProcessAudio = YES;
            }
            cfg.width                = 2;
            cfg.height               = 2;
            cfg.showsCursor          = NO;
            cfg.minimumFrameInterval = CMTimeMake(1, 1);

            _scDelegate = [[LaySCStreamDelegate alloc] init];
            _scStream   = [[SCStream alloc] initWithFilter:filter
                                             configuration:cfg
                                                  delegate:_scDelegate];

            NSError *addErr = nil;
            [_scStream addStreamOutput:_scDelegate
                                  type:SCStreamOutputTypeAudio
                    sampleHandlerQueue:_audioWriteQ
                                 error:&addErr];
            if (addErr) {
                scErr = strdup(addErr.localizedDescription.UTF8String);
                _scStream = nil; _scDelegate = nil;
                dispatch_semaphore_signal(sem);
                return;
            }

            [_scStream startCaptureWithCompletionHandler:^(NSError *startErr) {
                if (startErr) {
                    scErr = strdup(startErr.localizedDescription.UTF8String);
                    _scStream = nil; _scDelegate = nil;
                }
                dispatch_semaphore_signal(sem);
            }];
        }];

        dispatch_semaphore_wait(sem, dispatch_time(DISPATCH_TIME_NOW, 5 * NSEC_PER_SEC));

        if (scErr) {
            // System audio capture failed (e.g. Bluetooth output device, permissions).
            // Log and continue with mic-only — don't tear down the mic recording.
            NSLog(@"[lay] system audio capture unavailable (mic-only): %s", scErr);
            free((void *)scErr);
            _scStream = nil;
            _scDelegate = nil;
        }
    }

    _isRecording = YES;
    return NULL;
}

// stopCapture stops both streams and flushes audio files to disk.
static void stopCapture(void) {
    if (!_isRecording) return;
    _isRecording = NO;

    if (_engineConfigObserver) {
        [[NSNotificationCenter defaultCenter] removeObserver:_engineConfigObserver];
        _engineConfigObserver = nil;
    }

    if (_audioEngine) {
        [[_audioEngine inputNode] removeTapOnBus:0];
        [_audioEngine stop];
        _audioEngine = nil;
    }
    _micAudioFile = nil;    // ARC release flushes the full recording
    _chunkMicFile = nil;    // ARC release flushes the final live chunk
    _micRecordingFmt = nil;

    if (@available(macOS 13.0, *)) {
        if (_scStream) {
            dispatch_semaphore_t sem = dispatch_semaphore_create(0);
            [_scStream stopCaptureWithCompletionHandler:^(NSError *err) {
                dispatch_semaphore_signal(sem);
            }];
            dispatch_semaphore_wait(sem,
                dispatch_time(DISPATCH_TIME_NOW, 3 * NSEC_PER_SEC));
            _scStream = nil;
            _scDelegate = nil;
        }
    }

    if (_chunkSysExtFile) {
        ExtAudioFileDispose(_chunkSysExtFile); // flush final live sys chunk
        _chunkSysExtFile = NULL;
    }
    if (_sysExtFile) {
        ExtAudioFileDispose(_sysExtFile);
        _sysExtFile = NULL;
    }
    if (_sysFilePath) { free(_sysFilePath); _sysFilePath = NULL; }
    _sysFileReady       = NO;
    _sysNativeAsbdValid = NO;
    _audioWriteQ        = nil;
    if (_captureEvent) { free(_captureEvent); _captureEvent = NULL; }

    restoreOutputDevice();
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

func ProtectWindow() {
	C.protectAllWindows()
}

func MakeWindowStealth() {
	C.makeWindowStealth()
}

func SetAccessoryPolicy() {
	C.setAccessoryPolicy()
}

func RegisterGlobalHotkey() {
	C.registerGlobalHotkey()
}

func RegisterLocalKeyMonitor() {
	C.registerLocalKeyMonitor()
}

func UnregisterGlobalHotkey() {
	C.unregisterGlobalHotkey()
}

func UnregisterLocalKeyMonitor() {
	C.unregisterLocalKeyMonitor()
}

func StartCapture(dir string) error {
	cs := C.CString(dir)
	defer C.free(unsafe.Pointer(cs))
	errStr := C.startCapture(cs)
	if errStr != nil {
		msg := C.GoString(errStr)
		C.free(unsafe.Pointer(errStr))
		return errors.New(msg)
	}
	return nil
}

func StopCapture() {
	C.stopCapture()
}

func RotateChunk(newMicPath string) error {
	cs := C.CString(newMicPath)
	defer C.free(unsafe.Pointer(cs))
	if C.rotateChunk(cs) != 0 {
		return errors.New("chunk rotation failed")
	}
	return nil
}

func ConsumeCaptureEvent() (string, bool) {
	cs := C.consumeCaptureEvent()
	if cs == nil {
		return "", false
	}
	defer C.free(unsafe.Pointer(cs))
	return C.GoString(cs), true
}
