package engine

/*
#cgo CFLAGS: -I../../
#cgo LDFLAGS: -L../../ -lapple_ai_bridge -Wl,-rpath,../../ -Wl,-rpath,. -Wl,-rpath,@loader_path/../../
#include <stdlib.h>

extern void* CreateAppleAISession(const char* instructions);
extern void DestroyAppleAISession(void* sessionPtr);
extern char* AppleAISessionRespond(void* sessionPtr, const char* prompt);
extern void FreeAppleAIString(char* str);

typedef void (*TokenCallback)(const char* token, void* user_data);
extern char* AppleAISessionRespondStream(void* sessionPtr, const char* prompt, TokenCallback callback, void* userData);

// Declare the exported Go callback
void goTokenCallback(char* token, void* userData);
*/
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// Registry for active stream callbacks.
var (
	streamsMu     sync.Mutex
	nextStreamID  int
	activeStreams = make(map[int]func(string))
)

func registerStream(callback func(string)) int {
	streamsMu.Lock()
	defer streamsMu.Unlock()
	nextStreamID++
	activeStreams[nextStreamID] = callback
	return nextStreamID
}

func unregisterStream(id int) {
	streamsMu.Lock()
	defer streamsMu.Unlock()
	delete(activeStreams, id)
}

//export goTokenCallback
func goTokenCallback(token *C.char, userData unsafe.Pointer) {
	id := int(uintptr(userData))
	
	streamsMu.Lock()
	callback, exists := activeStreams[id]
	streamsMu.Unlock()
	
	if exists && token != nil {
		callback(C.GoString(token))
	}
}

// Session wraps the native Cgo Apple AI session.
type Session struct {
	ptr unsafe.Pointer
}

// NewSession instantiates a new Apple AI session with optional system instructions.
func NewSession(instructions string) (*Session, error) {
	var cInst *C.char
	if instructions != "" {
		cInst = C.CString(instructions)
		defer C.free(unsafe.Pointer(cInst))
	}

	ptr := C.CreateAppleAISession(cInst)
	if ptr == nil {
		return nil, errors.New("failed to initialize Apple AI session")
	}

	return &Session{ptr: ptr}, nil
}

// Close destroys the underlying native session and releases its resources.
func (s *Session) Close() {
	if s.ptr != nil {
		C.DestroyAppleAISession(s.ptr)
		s.ptr = nil
	}
}

// Respond executes a synchronous inference prompt and returns the full text response.
func (s *Session) Respond(prompt string) (string, error) {
	if s.ptr == nil {
		return "", errors.New("session is closed")
	}

	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	cResult := C.AppleAISessionRespond(s.ptr, cPrompt)
	if cResult == nil {
		return "", errors.New("nil response from engine")
	}
	defer C.FreeAppleAIString(cResult)

	res := C.GoString(cResult)
	return res, nil
}

// RespondStream executes a streaming inference prompt, calling the onToken callback for each new token.
func (s *Session) RespondStream(prompt string, onToken func(string)) error {
	if s.ptr == nil {
		return errors.New("session is closed")
	}

	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	streamID := registerStream(onToken)
	defer unregisterStream(streamID)

	userData := unsafe.Pointer(uintptr(streamID))

	cError := C.AppleAISessionRespondStream(
		s.ptr,
		cPrompt,
		C.TokenCallback(C.goTokenCallback),
		userData,
	)

	if cError != nil {
		defer C.FreeAppleAIString(cError)
		return fmt.Errorf("stream error: %s", C.GoString(cError))
	}

	return nil
}
