package util

func WithRecover(fn func()) (success bool) {
	defer func() {
		if recover() != nil {
			//recover panic from `send on closed channel`
			success = false
		}
	}()

	fn()

	return true
}
