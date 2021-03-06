// +build !windows,!plan9

package sys

import (
	"os"

	"golang.org/x/sys/unix"
)

// Ioctl wraps the ioctl syscall.
func Ioctl(fd int, req int, arg uintptr) error {
	_, _, e := unix.Syscall(
		unix.SYS_IOCTL, uintptr(fd), uintptr(req), arg)
	if e != 0 {
		return os.NewSyscallError("ioctl", e)
	}
	return nil
}
