package sdk

import "syscall"

func init() {
	SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
