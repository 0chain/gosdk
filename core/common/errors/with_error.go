package errors

type withError struct {
	previous error
	current  error
}

func (w *withError) Error() string {
	var retError string
	if w.current != nil {
		retError = w.current.Error()
	}
	if w.previous != nil {
		retError += "\n" + w.previous.Error()
	}
	return retError
}

func (w *withError) pprint() string {
	retError := ""
	if w.current != nil {
		switch c := w.current.(type) {
		case *Error:
			retError += c.pprint()
		case *withError:
			retError += c.pprint()
		default:
			retError += c.Error()
		}
	}
	if w.previous != nil {
		// retError += "\n" + w.previous.Error()
		switch p := w.previous.(type) {
		case *Error:
			retError += ": " + p.pprint()
		case *withError:
			retError += ": " + p.pprint()
		default:
			retError += ": " + p.Error()
		}
	}
	return retError
}
