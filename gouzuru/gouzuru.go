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

func GetMinMaxAddress() (min, max uintptr, err error) {
	si, err := w32.GetSystemInfo()
	if err != nil {
		return 0, 0, err
	}

	return si.MinimumApplicationAddress,
		si.MaximumApplicationAddress,
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

	// Use VirtualQueryEx to get info about all the regions in the process.
	for addr := minAddress; addr < maxAddress; {
		ret, err := w32.VirtualQueryEx(p.Handle, addr)
		if err != nil {
			return regions, err
		}

		// Save the region.
		regions = append(regions, ret)

		// Move to the next region.
		addr += uintptr(ret.RegionSize)
	}

	return regions, nil
}
