package main

import (
	"fmt"
	"github.com/carterjones/gouzuru/w32"
	"os"
)

type Process struct {
	name   string
	pid    int32 // DWORD
	handle uintptr
}

func GetProcessNameFromPid(pid int32) (name string, err error) {
	accessLevel := int32(w32.PROCESS_QUERY_INFORMATION |
		w32.PROCESS_QUERY_LIMITED_INFORMATION)
	hwnd, err := w32.OpenProcess(pid, accessLevel)
	if err != nil {
		return "", err
	}

	procName, err := w32.GetProcessImageFileName(hwnd)
	if err != nil {
		return "", err
	}

	return procName, nil
}

func GetMaxAddress() (address int32, err error) {
	// TODO: Use SYSTEM_INFO to get the max application address
	return 0, nil
}

func (p Process) IdentifyRegions() (regions []w32.MEMORY_BASIC_INFORMATION, err error) {
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
	pids, err := w32.GetProcessIds()
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
	hwnd, err := w32.OpenProcess(targetPid, int32(w32.PROCESS_ALL_ACCESS))
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
