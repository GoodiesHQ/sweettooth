package system

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type elevation struct {
	tokenIsElevated uint32
}

func IsAdmin() bool {
	// Get the current process token
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	// Check if the token has elevated privileges
	var el elevation
	var retLen uint32
	err = windows.GetTokenInformation(token, windows.TokenElevation, (*byte)(unsafe.Pointer(&el)), uint32(unsafe.Sizeof(el)), &retLen)
	if err != nil {
		return false
	}

	return el.tokenIsElevated != 0
}
