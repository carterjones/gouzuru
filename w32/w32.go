package w32

import (
	"errors"
	"golang.org/x/sys/windows"
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
	getSystemInfo           = modkernel32.NewProc("GetSystemInfo")
	virtualQueryEx          = modkernel32.NewProc("VirtualQueryEx")
	readProcessMemory       = modkernel32.NewProc("ReadProcessMemory")
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
	// Memory states.
	MEM_COMMIT  int = 0x1000
	MEM_FREE    int = 0x10000
	MEM_RESERVE int = 0x2000
	// Page protection contstants.
	PAGE_EXECUTE           int = 0x10
	PAGE_EXECUTE_READ      int = 0x20
	PAGE_EXECUTE_READWRITE int = 0x40
	PAGE_EXECUTE_WRITECOPY int = 0x80
	PAGE_NOACCESS          int = 0x01
	PAGE_READONLY          int = 0x02
	PAGE_READWRITE         int = 0x04
	PAGE_WRITECOPY         int = 0x08
	PAGE_TARGETS_INVALID   int = 0x40000000
	PAGE_TARGETS_NO_UPDATE int = 0x40000000
	PAGE_GUARD             int = 0x100
	PAGE_NOCACHE           int = 0x200
	PAGE_WRITECOMBINE      int = 0x400
)

type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect int32
	__alignment1      int32
	RegionSize        uintptr
	State             int32
	Protect           int32
	Type              int32
	__alignment2      int32
}

func (mbi MEMORY_BASIC_INFORMATION) IsReadable() bool {
	// Perform some type conversions.
	state := int(mbi.State)
	protect := int(mbi.Protect)

	// Verify that the memory is used by the target process.
	stateOk := state == MEM_COMMIT

	// Verify that protection flags don't prohibit access.
	protectOk := (protect&PAGE_NOACCESS) == 0 && (protect&PAGE_GUARD) == 0

	// Verify that at least one read flag is enabled.
	protectOk = protectOk && ((protect&PAGE_EXECUTE_READ) != 0 ||
		(protect&PAGE_EXECUTE_READWRITE) != 0 ||
		(protect&PAGE_READONLY) != 0 ||
		(protect&PAGE_READWRITE) != 0)

	// Return the results.
	return stateOk && protectOk
}

type SYSTEM_INFO struct {
	ProcessorArchitecture     int16
	reserved                  int16
	PageSize                  int32
	MinimumApplicationAddress uintptr
	MaximumApplicationAddress uintptr
	ActiveProcessorMask       uintptr
	NumberOfProcessors        int32
	ProcessorType             int32
	AllocationGranularity     int32
	ProcessorLevel            int16
	ProcessorRevision         int16
}

func (si SYSTEM_INFO) OemId() int32 {
	oemId := int32(si.ProcessorArchitecture) << 16
	oemId += int32(si.reserved)
	return oemId
}

func EnumProcesses() (procList []int32, err error) {
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

func GetProcessImageFileName(hwnd uintptr) (procName string, err error) {
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
	_, procName = filepath.Split(imageName)

	return procName, nil
}

func GetSystemInfo() (si SYSTEM_INFO, err error) {
	// void WINAPI GetSystemInfo(
	//   _Out_ LPSYSTEM_INFO lpSystemInfo
	// );
	// This function does not return a value.
	var numArgs uintptr = 1
	_, _, err = syscall.Syscall(getSystemInfo.Addr(),
		numArgs,
		uintptr(unsafe.Pointer(&si)),
		0,
		0)
	if err != nil {
		if err.Error() != "The operation completed successfully." {
			return SYSTEM_INFO{}, err
		}
	}

	return si, nil
}

func VirtualQueryEx(hwnd, baseAddress uintptr) (mbi MEMORY_BASIC_INFORMATION, err error) {
	// SIZE_T WINAPI VirtualQueryEx(
	//   _In_     HANDLE                    hProcess,
	//   _In_opt_ LPCVOID                   lpAddress,
	//   _Out_    PMEMORY_BASIC_INFORMATION lpBuffer,
	//   _In_     SIZE_T                    dwLength
	// );
	// The return value is the actual number of bytes returned in the information buffer.
	var numArgs uintptr = 4
	ret, _, err := syscall.Syscall6(virtualQueryEx.Addr(),
		numArgs,
		hwnd,
		baseAddress,
		uintptr(unsafe.Pointer(&mbi)),
		unsafe.Sizeof(mbi),
		0,
		0)
	if ret == 0 {
		return MEMORY_BASIC_INFORMATION{}, err
	}

	return mbi, nil
}

func ReadProcessMemory(hwnd, addr, size uintptr) (data []byte, err error) {
	// BOOL WINAPI ReadProcessMemory(
	//   _In_  HANDLE  hProcess,
	//   _In_  LPCVOID lpBaseAddress,
	//   _Out_ LPVOID  lpBuffer,
	//   _In_  SIZE_T  nSize,
	//   _Out_ SIZE_T  *lpNumberOfBytesRead
	// );
	var numArgs uintptr = 5
	data = make([]byte, size)
	var nbr uintptr = 0
	ret, _, err := syscall.Syscall6(readProcessMemory.Addr(),
		numArgs,
		hwnd,
		addr,
		uintptr(unsafe.Pointer(&data[0])),
		size,
		uintptr(unsafe.Pointer(&nbr)),
		0)
	if ret == 0 {
		return nil, err
	}

	return data, nil
}
