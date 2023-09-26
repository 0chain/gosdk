//go:build !windows
// +build !windows

package sdk

import "syscall"

var (
	SysProcAttr = &syscall.SysProcAttr{}
)