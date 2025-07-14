package errs

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(status int, msg string) error {
	return HTTPError{StatusCode: status, Message: msg}
}
