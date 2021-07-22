package errors

/* Is - tells whether actual error is targer error
where, actual error can be either Error/withError
if actual error is wrapped error then if any internal error
matches the target error then function results in true
*/
func Is(actual error, target *Error) bool {
	actualError := isError(actual)
	if actualError != nil {
		if actualError.Code == "" && target.Code == "" {
			return actualError.Msg == target.Msg
		} else {
			return actualError.Code == target.Code
		}

	} else {
		actualWithError := isWithError(actual)
		if actualWithError != nil {
			if actualWithError.current != nil {
				return Is(actualWithError.current, target)
			} else {
				return Is(actualWithError.previous, target)
			}
		} else {
			return false
		}
	}
}

// isError - parses the error into Error
func isError(err error) *Error {
	t, ok := err.(*Error)
	if ok {
		return t
	}
	return nil
}

// isError - parses the error into withError
func isWithError(err error) *withError {
	t, ok := err.(*withError)
	if ok {
		return t
	}
	return nil
}
