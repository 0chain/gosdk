//go:build !windows
// +build !windows

package sdk

import "syscall"

var (
	sysProcAttr = &syscall.SysProcAttr{}
)
