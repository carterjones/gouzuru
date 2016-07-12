package main

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	// The following assumes OS is greater than or equal to Windows 7.
	enumProcesses = modkernel32.NewProc("K32EnumProcesses")
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

func main() {
	// Get the process list.
	pids, err := GetProcessIds()
	if err != nil {
		fmt.Println("[-] error:")
		fmt.Println(err)
		return
	}

	// List the process IDs.
	for _, p := range pids {
		fmt.Println("PID: ", p)
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
