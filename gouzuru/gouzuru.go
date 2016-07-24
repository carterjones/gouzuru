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

func GetMaxAddress() (address int32, err error) {
	// TODO: Use SYSTEM_INFO to get the max application address
	return 0, nil
}

type Process struct {
	Name   string
	Pid    int32 // DWORD
	Handle uintptr
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
