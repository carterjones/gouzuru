package main

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
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
	MAX_PATH int = 256
	// Process rights.
	PROCESS_CREATE_PROCESS            int = 0x0080
	PROCESS_CREATE_THREAD             int = 0x0002
	PROCESS_DUP_HANDLE                int = 0x0040
	PROCESS_QUERY_INFORMATION         int = 0x0400
	PROCESS_QUERY_LIMITED_INFORMATION int = 0x1000
	PROCESS_SET_INFORMATION           int = 0x0200
	PROCESS_SET_QUOTA                 int = 0x0100
	PROCESS_SUSPEND_RESUME            int = 0x0800
	PROCESS_TERMINATE                 int = 0x0001
	PROCESS_VM_OPERATION              int = 0x0008
	PROCESS_VM_READ                   int = 0x0010
	PROCESS_VM_WRITE                  int = 0x0020
	SYNCHRONIZE                       int = 0x00100000
	PROCESS_ALL_ACCESS                int = PROCESS_CREATE_PROCESS | PROCESS_CREATE_THREAD | PROCESS_DUP_HANDLE | PROCESS_QUERY_INFORMATION | PROCESS_QUERY_LIMITED_INFORMATION | PROCESS_SET_INFORMATION | PROCESS_SET_QUOTA | PROCESS_SUSPEND_RESUME | PROCESS_TERMINATE | PROCESS_VM_OPERATION | PROCESS_VM_READ | PROCESS_VM_WRITE | SYNCHRONIZE
)

type Process struct {
	name   string
	pid    int32 // DWORD
	handle uintptr
}

type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect int32
	RegionSize        int32
	State             int32
	Protect           int32
	Type_             int32
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

	imageName := string(imageFileName[:ret])
	_, procName := filepath.Split(imageName)

	return procName, nil
}

func GetMaxAddress() (address int32, err error) {
	// TODO: Use SYSTEM_INFO to get the max application address
	return 0, nil
}

func (p Process) IdentifyRegions() (regions []MEMORY_BASIC_INFORMATION, err error) {
	minAddress := int32(0)
	maxAddress, err := GetMaxAddress()
	if err != nil {
		return nil, err
	}

	// TODO: use VirtualQueryEx to get info about all the regions in the process
	for addr := minAddress; addr < maxAddress; {
		// TODO: make this increase by the current region's size
		addr += 1
	}

	return nil, nil
}

func main() {
	// Set the target process to the first argument.
	targetProcName := os.Args[1]

	// Get the process list.
	pids, err := GetProcessIds()
	if err != nil {
		fmt.Println("[-] error:", err)
		return
	}

	// Find the target PID.
	targetPid := int32(0)
	for _, p := range pids {
		name, err := GetProcessNameFromPid(p)
		if err != nil {
			fmt.Printf("[-] error for PID: %v: %v\n", p, err)
		} else if name == targetProcName {
			targetPid = p
		}
	}
	if targetPid == 0 {
		fmt.Printf("Unable to open %v. You might need more permissions or the "+
			"target process might not exist.\n", targetProcName)
		return
	}

	// Open the target process.
	hwnd, err := OpenProcess(targetPid, int32(PROCESS_ALL_ACCESS))
	if err != nil {
		fmt.Println("[-] error:", err)
		return
	}

	proc := Process{name: targetProcName, pid: targetPid, handle: hwnd}
	fmt.Printf("Successfully opened %v. PID: %v. Handle: %v.\n",
		proc.name, proc.pid, proc.handle)

	// Get information about the page ranges of the process.
	proc.IdentifyRegions()

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
