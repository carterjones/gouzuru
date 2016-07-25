package main

import (
	"flag"
	"fmt"
	"github.com/carterjones/gouzuru/gouzuru"
	"github.com/carterjones/gouzuru/w32"
	"strings"
)

func handleError(err error) bool {
	if err != nil {
		fmt.Println("[-] error:", err)
		return true
	}
	return false
}

func main() {
	var targetProcName = flag.String("p",
		"<target-process.exe>",
		"name of the target process (including .exe)")
	flag.Parse()

	// Get the process list.
	pids, err := w32.EnumProcesses()
	if handleError(err) {
		return
	}

	// Find the target PID.
	targetPid := int32(0)
	for _, p := range pids {
		name, err := gouzuru.GetProcessNameFromPid(p)
		if err != nil {
			// Ignore the SYSTEM process.
			if p == 0 {
				continue
			}

			// Ignore access denied errors.
			if strings.Contains(err.Error(), "Access is denied.") {
				continue
			}

			fmt.Printf("[-] error for PID: %v: %v\n", p, err)
		} else if name == *targetProcName {
			targetPid = p
		}
	}
	if targetPid == 0 {
		fmt.Printf("Unable to open %v. You might need more permissions or the "+
			"target process might not exist.\n", *targetProcName)
		return
	}

	// Open the target process.
	hwnd, err := w32.OpenProcess(targetPid, int32(w32.PROCESS_ALL_ACCESS))
	if handleError(err) {
		return
	}

	// Make a process object.
	proc := gouzuru.Process{
		Name:   *targetProcName,
		Pid:    targetPid,
		Handle: hwnd,
	}
	fmt.Printf("Successfully opened %v. PID: %v. Handle: %v.\n",
		proc.Name, proc.Pid, proc.Handle)

	// Get information about the page ranges of the process.
	regions, err := proc.IdentifyRegions()
	if handleError(err) {
		return
	}

	// Iterate over each region.
	for _, r := range regions {
		// Read the entire region into a buffer.
		if r.IsReadable() {
			data, err := w32.ReadProcessMemory(proc.Handle, r.BaseAddress, r.RegionSize)
			if handleError(err) {
				return
			}

			// Print the first two bytes.
			fmt.Printf("first 2 bytes: %c%c\n", data[0], data[1])
		}
	}
}
