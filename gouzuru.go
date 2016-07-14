package main

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

// The following definitions assume the OS is greater than or equal to Windows 7.
var (
	modkernel32             = windows.NewLazySystemDLL("kernel32.dll")
	enumProcesses           = modkernel32.NewProc("K32EnumProcesses")
	openProcess             = modkernel32.NewProc("OpenProcess")
	getProcessImageFileName = modkernel32.NewProc("K32GetProcessImageFileNameA")
)

const (
	MAX_PATH                          int = 256
	PROCESS_QUERY_INFORMATION         int = 0x0400
	PROCESS_QUERY_LIMITED_INFORMATION int = 0x1000
)

type Process struct {
	name string
	pid  int // DWORD
}

func GetProcessIds() (procList []int32, err error) {
	// BOOL WINAPI EnumProcesses(
	//   _Out_ DWORD *pProcessIds,
	//   _In_  DWORD cb,
	//   _Out_ DWORD *pBytesReturned
	// );
	var numArgs uintptr = 4
	const numProcIds uint32 = 65535
	var processIds [numProcIds]int32
	cb := numProcIds
	numBytesReturned := 0

	ret, _, err := syscall.Syscall(enumProcesses.Addr(),
		numArgs,
		uintptr(unsafe.Pointer(&processIds[0])),
		uintptr(numProcIds),
		uintptr(unsafe.Pointer(&numBytesReturned)))
	if ret == 0 {
		return nil, err
	}

	const sizeofDword = 4
	numProcs := uint32(numBytesReturned / sizeofDword)

	if numProcs == cb {
		// There may be more processes.
		// Try re-enumerating with a larger array.
		return nil, errors.New("too many process IDs were returned")
	}

	return processIds[:numProcs], nil
}

func OpenProcess(pid int32, accessLevel int32) (hwnd uintptr, err error) {
	// HANDLE WINAPI OpenProcess(
	//   _In_ DWORD dwDesiredAccess,
	//   _In_ BOOL  bInheritHandle,
	//   _In_ DWORD dwProcessId
	// );
	var numArgs uintptr = 3
	ret, _, err := syscall.Syscall(openProcess.Addr(),
		numArgs,
		uintptr(accessLevel),
		uintptr(0),
		uintptr(pid))
	if ret == 0 {
		return uintptr(0), err
	}

	return ret, nil
}

func GetProcessNameFromPid(pid int32) (name string, err error) {
	accessLevel := int32(PROCESS_QUERY_INFORMATION |
		PROCESS_QUERY_LIMITED_INFORMATION)
	hwnd, err := OpenProcess(pid, accessLevel)
	if err != nil {
		return "", err
	}

	// DWORD WINAPI GetProcessImageFileName(
	//   _In_  HANDLE hProcess,
	//   _Out_ LPTSTR lpImageFileName,
	//   _In_  DWORD  nSize
	// );
	var numArgs uintptr = 3
	var imageFileName [MAX_PATH]byte
	ret, _, err := syscall.Syscall(getProcessImageFileName.Addr(),
		numArgs,
		hwnd,
		uintptr(unsafe.Pointer(&imageFileName[0])),
		uintptr(MAX_PATH))
	if ret == 0 {
		return "", err
	}

	return string(imageFileName[:ret]), nil
}

func main() {
	// Get the process list.
	pids, err := GetProcessIds()
	if err != nil {
		fmt.Println("[-] error:", err)
		return
	}

	// Correlate the process IDs with process names.
	for _, p := range pids {
		name, err := GetProcessNameFromPid(p)
		if err != nil {
			fmt.Printf("[-] error for PID: %v: %v\n", p, err)
		} else {
			fmt.Println("PID: ", p, "Name: ", name)
		}
	}

	// Open one of the processes.
	// TODO: hwnd, err := OpenProcess(w32.PROCESS_ALL_ACCESS, false, pid)
	// HANDLE WINAPI OpenProcess(
	//   _In_ DWORD dwDesiredAccess,
	//   _In_ BOOL  bInheritHandle,
	//   _In_ DWORD dwProcessId
	// );

	// Read some memory.
	// TODO: data, err := ReadProcessMemory(hwnd, address, 1)
	// BOOL WINAPI ReadProcessMemory(
	//   _In_  HANDLE  hProcess,
	//   _In_  LPCVOID lpBaseAddress,
	//   _Out_ LPVOID  lpBuffer,
	//   _In_  SIZE_T  nSize,
	//   _Out_ SIZE_T  *lpNumberOfBytesRead
	// );
}
