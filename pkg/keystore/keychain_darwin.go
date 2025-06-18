package keystore

/*
#cgo LDFLAGS: -framework Security -framework CoreFoundation
#include <Security/Security.h>
#include <Security/SecItem.h>
#include <CoreFoundation/CoreFoundation.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// setGenericPassword stores a password in the macOS Keychain
func setGenericPassword(label, service, account string, password []byte) error {
	allocator := C.kCFAllocatorDefault
	query := C.CFDictionaryCreateMutable(allocator, 0, nil, nil)

	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecClass),
		unsafe.Pointer(C.kSecClassGenericPassword))

	ca := C.CString(account)
	defer C.free(unsafe.Pointer(ca))
	cfAccount := C.CFStringCreateWithCString(allocator, ca, C.kCFStringEncodingUTF8)
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecAttrAccount),
		unsafe.Pointer(cfAccount))

	cs := C.CString(service)
	defer C.free(unsafe.Pointer(cs))
	cfService := C.CFStringCreateWithCString(allocator, cs, C.kCFStringEncodingUTF8)
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecAttrService),
		unsafe.Pointer(cfService))

	cl := C.CString(label)
	defer C.free(unsafe.Pointer(cl))
	cfLabel := C.CFStringCreateWithCString(allocator, cl, C.kCFStringEncodingUTF8)
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecAttrLabel),
		unsafe.Pointer(cfLabel))

	cfPassword := C.CFDataCreate(allocator, (*C.UInt8)(unsafe.Pointer(&password[0])), C.CFIndex(len(password)))
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecValueData),
		unsafe.Pointer(cfPassword))

	status := C.SecItemAdd(C.CFDictionaryRef(query), nil)
	if status == C.errSecDuplicateItem {
		// Update instead
		update := C.CFDictionaryCreateMutable(allocator, 0, nil, nil)
		C.CFDictionaryAddValue(update,
			unsafe.Pointer(C.kSecValueData),
			unsafe.Pointer(cfPassword))
		status = C.SecItemUpdate(C.CFDictionaryRef(query), C.CFDictionaryRef(update))
	}

	if status != C.errSecSuccess {
		return errors.New("failed to set password")
	}
	return nil
}

// getGenericPassword retrieves a password from the macOS Keychain
func getGenericPassword(service, account string) (username string, password []byte, err error) {
	allocator := C.kCFAllocatorDefault
	query := C.CFDictionaryCreateMutable(allocator, 0, nil, nil)

	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecClass),
		unsafe.Pointer(C.kSecClassGenericPassword))

	ca := C.CString(account)
	defer C.free(unsafe.Pointer(ca))
	cfAccount := C.CFStringCreateWithCString(allocator, ca, C.kCFStringEncodingUTF8)
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecAttrAccount),
		unsafe.Pointer(cfAccount))

	cs := C.CString(service)
	defer C.free(unsafe.Pointer(cs))
	cfService := C.CFStringCreateWithCString(allocator, cs, C.kCFStringEncodingUTF8)
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecAttrService),
		unsafe.Pointer(cfService))

	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecMatchLimit),
		unsafe.Pointer(C.kSecMatchLimitOne))
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecReturnAttributes),
		unsafe.Pointer(C.kCFBooleanTrue))
	C.CFDictionaryAddValue(query,
		unsafe.Pointer(C.kSecReturnData),
		unsafe.Pointer(C.kCFBooleanTrue))

	var item C.CFTypeRef
	status := C.SecItemCopyMatching(C.CFDictionaryRef(query), &item)

	if status == C.errSecItemNotFound {
		err = errors.New("no password found")
		return
	} else if status != C.errSecSuccess {
		err = errors.New("unhandled error")
		return
	}

	dict := C.CFDictionaryRef(item)
	valueData := C.CFDictionaryGetValue(dict, unsafe.Pointer(C.kSecValueData))
	dataLen := C.CFDataGetLength((C.CFDataRef)(valueData))
	dataPtr := C.CFDataGetBytePtr((C.CFDataRef)(valueData))
	password = C.GoBytes(unsafe.Pointer(dataPtr), C.int(dataLen))

	accountVal := C.CFDictionaryGetValue(dict, unsafe.Pointer(C.kSecAttrAccount))
	var accountCStr [1024]byte
	C.CFStringGetCString((C.CFStringRef)(accountVal), (*C.char)(unsafe.Pointer(&accountCStr[0])), 1024, C.kCFStringEncodingUTF8)
	username = C.GoString((*C.char)(unsafe.Pointer(&accountCStr[0])))

	return
} 