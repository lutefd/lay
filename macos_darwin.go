//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon -framework ScreenCaptureKit -framework AVFoundation -framework CoreMedia -framework AudioToolbox

#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>
#include <Carbon/Carbon.h>
#import <AVFoundation/AVFoundation.h>
#import <AudioToolbox/AudioToolbox.h>
#import <CoreMedia/CoreMedia.h>
#import <ScreenCaptureKit/ScreenCaptureKit.h>

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

// ──────────────────────────────────────────────────────────────────────────────
// Audio capture — ScreenCaptureKit (system audio) + AVAudioEngine (mic)
// ──────────────────────────────────────────────────────────────────────────────

static AVAudioEngine      *_audioEngine    = nil;
static AVAudioFile        *_micAudioFile   = nil;
static id                  _scStream       = nil;  // SCStream*
static id                  _scDelegate     = nil;  // LaySCStreamDelegate*
static ExtAudioFileRef     _sysExtFile     = NULL;
static dispatch_queue_t    _audioWriteQ    = nil;
static char               *_sysFilePath    = NULL; // malloc'd, owned here
static BOOL                _sysFileReady   = NO;   // file created on first callback
static BOOL                _isRecording    = NO;

@interface LaySCStreamDelegate : NSObject <SCStreamOutput, SCStreamDelegate>
@end

@implementation LaySCStreamDelegate

- (void)stream:(SCStream *)stream
    didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer
    ofType:(SCStreamOutputType)type
{
    if (type != SCStreamOutputTypeAudio) return;

    // On the first callback we learn the actual format from the buffer itself —
    // no assumptions about interleaving or sample rate needed.
    if (!_sysFileReady && _sysFilePath) {
        CMFormatDescriptionRef fmtDesc = CMSampleBufferGetFormatDescription(sampleBuffer);
        const AudioStreamBasicDescription *asbd =
            CMAudioFormatDescriptionGetStreamBasicDescription(fmtDesc);
        if (asbd) {
            NSURL *url = [NSURL fileURLWithPath:
                [NSString stringWithUTF8String:_sysFilePath]];
            // CAF handles any PCM format natively; no conversion needed.
            ExtAudioFileCreateWithURL((__bridge CFURLRef)url, kAudioFileCAFType,
                asbd, NULL, kAudioFileFlags_EraseFile, &_sysExtFile);
            // Client format = file format — passthrough, no conversion.
            ExtAudioFileSetProperty(_sysExtFile,
                kExtAudioFileProperty_ClientDataFormat, sizeof(*asbd), asbd);
            _sysFileReady = YES;
        }
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
    }

    free(abl);
    if (blockBuf) CFRelease(blockBuf);
}

- (void)stream:(SCStream *)stream didStopWithError:(NSError *)error {}

@end

// startCapture starts both mic and system audio capture into outDir.
// Returns NULL on success, or a malloc'd C string with the error message.
static const char *startCapture(const char *outDir) {
    if (_isRecording) return NULL;

    NSString *dir = [NSString stringWithUTF8String:outDir];

    // ── Microphone via AVAudioEngine ─────────────────────────────────────────

    _audioEngine = [[AVAudioEngine alloc] init];
    AVAudioInputNode *inputNode = [_audioEngine inputNode];
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

    // Install tap in the file's processingFormat so AVAudioEngine converts
    // from the hardware format for us before we write.
    AVAudioFormat *tapFmt = _micAudioFile.processingFormat;
    [inputNode installTapOnBus:0 bufferSize:4096 format:tapFmt
                         block:^(AVAudioPCMBuffer *buf, AVAudioTime *when) {
        if (_micAudioFile) {
            [_micAudioFile writeFromBuffer:buf error:nil];
        }
    }];

    NSError *engErr = nil;
    [_audioEngine startAndReturnError:&engErr];
    if (engErr) {
        [inputNode removeTapOnBus:0];
        _audioEngine = nil;
        _micAudioFile = nil;
        return strdup(engErr.localizedDescription.UTF8String);
    }

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
            cfg.sampleRate           = 48000;
            cfg.channelCount         = 2;
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
            [[_audioEngine inputNode] removeTapOnBus:0];
            [_audioEngine stop];
            _audioEngine = nil;
            _micAudioFile = nil;
            return scErr;
        }
    }

    _isRecording = YES;
    return NULL;
}

// stopCapture stops both streams and flushes audio files to disk.
static void stopCapture(void) {
    if (!_isRecording) return;
    _isRecording = NO;

    if (_audioEngine) {
        [[_audioEngine inputNode] removeTapOnBus:0];
        [_audioEngine stop];
        _audioEngine = nil;
    }
    _micAudioFile = nil; // ARC release flushes the CAF

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

    if (_sysExtFile) {
        ExtAudioFileDispose(_sysExtFile);
        _sysExtFile = NULL;
    }
    if (_sysFilePath) { free(_sysFilePath); _sysFilePath = NULL; }
    _sysFileReady = NO;
    _audioWriteQ  = nil;
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

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

// StartCapture begins mic + system audio capture, writing WAV files into dir.
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

// StopCapture stops any active capture and flushes audio to disk.
func StopCapture() {
	C.stopCapture()
}
