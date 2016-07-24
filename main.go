package main

import (
	"fmt"
	"github.com/carterjones/gouzuru/gouzuru"
	"github.com/carterjones/gouzuru/w32"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		_, exeName := filepath.Split(os.Args[0])
		fmt.Println("Usage:", exeName, "<process-name>")
		return
	}

	// Set the target process to the first argument.
	targetProcName := os.Args[1]

	// Get the process list.
	pids, err := w32.EnumProcesses()
	if err != nil {
		fmt.Println("[-] error:", err)
		return
	}

	// Find the target PID.
	targetPid := int32(0)
	for _, p := range pids {
		name, err := gouzuru.GetProcessNameFromPid(p)
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

	proc := gouzuru.Process{
		Name: targetProcName,
		Pid: targetPid,
		Handle: hwnd,
	}
	fmt.Printf("Successfully opened %v. PID: %v. Handle: %v.\n",
		proc.Name, proc.Pid, proc.Handle)

	// Get information about the page ranges of the process.
	regions, err := proc.IdentifyRegions()
	if err != nil {
		fmt.Println("[-] error:", err)
		return
	}

	for _, r := range(regions) {
		fmt.Println("region: %v, size: %v", r.BaseAddress, r.RegionSize)
	}

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
