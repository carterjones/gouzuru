package gouzuru

import (
	"github.com/carterjones/gouzuru/w32"
)

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

func GetMinMaxAddress() (min, max int32, err error) {
	si, err := w32.GetSystemInfo()
	if err != nil {
		return 0, 0, err
	}

	return int32(si.MinimumApplicationAddress),
		int32(si.MaximumApplicationAddress),
		nil
}

type Process struct {
	Name   string
	Pid    int32
	Handle uintptr
}

func (p Process) IdentifyRegions() (regions []w32.MEMORY_BASIC_INFORMATION, err error) {
	minAddress, maxAddress, err := GetMinMaxAddress()
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
