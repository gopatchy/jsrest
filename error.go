package jsrest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

var (
	ErrBadRequest                   = NewHTTPError(http.StatusBadRequest)
	ErrUnauthorized                 = NewHTTPError(http.StatusUnauthorized)
	ErrPaymentRequired              = NewHTTPError(http.StatusPaymentRequired)
	ErrForbidden                    = NewHTTPError(http.StatusForbidden)
	ErrNotFound                     = NewHTTPError(http.StatusNotFound)
	ErrMethodNotAllowed             = NewHTTPError(http.StatusMethodNotAllowed)
	ErrNotAcceptable                = NewHTTPError(http.StatusNotAcceptable)
	ErrProxyAuthRequired            = NewHTTPError(http.StatusProxyAuthRequired)
	ErrRequestTimeout               = NewHTTPError(http.StatusRequestTimeout)
	ErrConflict                     = NewHTTPError(http.StatusConflict)
	ErrGone                         = NewHTTPError(http.StatusGone)
	ErrLengthRequired               = NewHTTPError(http.StatusLengthRequired)
	ErrPreconditionFailed           = NewHTTPError(http.StatusPreconditionFailed)
	ErrRequestEntityTooLarge        = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrRequestURITooLong            = NewHTTPError(http.StatusRequestURITooLong)
	ErrUnsupportedMediaType         = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrRequestedRangeNotSatisfiable = NewHTTPError(http.StatusRequestedRangeNotSatisfiable)
	ErrExpectationFailed            = NewHTTPError(http.StatusExpectationFailed)
	ErrTeapot                       = NewHTTPError(http.StatusTeapot)
	ErrMisdirectedRequest           = NewHTTPError(http.StatusMisdirectedRequest)
	ErrUnprocessableEntity          = NewHTTPError(http.StatusUnprocessableEntity)
	ErrLocked                       = NewHTTPError(http.StatusLocked)
	ErrFailedDependency             = NewHTTPError(http.StatusFailedDependency)
	ErrTooEarly                     = NewHTTPError(http.StatusTooEarly)
	ErrUpgradeRequired              = NewHTTPError(http.StatusUpgradeRequired)
	ErrPreconditionRequired         = NewHTTPError(http.StatusPreconditionRequired)
	ErrTooManyRequests              = NewHTTPError(http.StatusTooManyRequests)
	ErrRequestHeaderFieldsTooLarge  = NewHTTPError(http.StatusRequestHeaderFieldsTooLarge)
	ErrUnavailableForLegalReasons   = NewHTTPError(http.StatusUnavailableForLegalReasons)

	ErrInternalServerError           = NewHTTPError(http.StatusInternalServerError)
	ErrNotImplemented                = NewHTTPError(http.StatusNotImplemented)
	ErrBadGateway                    = NewHTTPError(http.StatusBadGateway)
	ErrServiceUnavailable            = NewHTTPError(http.StatusServiceUnavailable)
	ErrGatewayTimeout                = NewHTTPError(http.StatusGatewayTimeout)
	ErrHTTPVersionNotSupported       = NewHTTPError(http.StatusHTTPVersionNotSupported)
	ErrVariantAlsoNegotiates         = NewHTTPError(http.StatusVariantAlsoNegotiates)
	ErrInsufficientStorage           = NewHTTPError(http.StatusInsufficientStorage)
	ErrLoopDetected                  = NewHTTPError(http.StatusLoopDetected)
	ErrNotExtended                   = NewHTTPError(http.StatusNotExtended)
	ErrNetworkAuthenticationRequired = NewHTTPError(http.StatusNetworkAuthenticationRequired)
)

type HTTPError struct {
	Code    int
	Message string
}

func NewHTTPError(code int) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: http.StatusText(code),
	}
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("[%d] %s", err.Code, err.Message)
}

type SilentJoinError struct {
	Wraps []error
}

func SilentJoin(errs ...error) *SilentJoinError {
	return &SilentJoinError{
		Wraps: errs,
	}
}

func (err *SilentJoinError) Error() string {
	return err.Wraps[0].Error()
}

func (err *SilentJoinError) Unwrap() []error {
	return err.Wraps
}

func WriteError(w http.ResponseWriter, err error) {
	je := ToJSONError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(je.Code)

	enc := json.NewEncoder(w)
	_ = enc.Encode(je) //nolint:errchkjson
}

func Errorf(he *HTTPError, format string, a ...any) error {
	err := fmt.Errorf(format, a...) //nolint:goerr113

	if hasHTTPError(err) {
		return err
	}

	return SilentJoin(err, he)
}

type JSONError struct {
	Code     int      `json:"-"`
	Messages []string `json:"messages"`
}

func (je *JSONError) Error() string {
	if len(je.Messages) > 0 {
		return je.Messages[0]
	}

	return "no error message"
}

func (je *JSONError) Unwrap() error {
	if len(je.Messages) > 1 {
		return &JSONError{
			Code:     je.Code,
			Messages: je.Messages[1:],
		}
	}

	return nil
}

func ToJSONError(err error) *JSONError {
	je := &JSONError{
		Code: 500,
	}
	je.importError(err)

	return je
}

func ReadError(resp *resty.Response) error {
	jse := &JSONError{}

	err := json.Unmarshal(resp.Body(), jse)
	if err == nil {
		return jse
	}

	return NewHTTPError(resp.StatusCode())
}

type singleUnwrap interface {
	Unwrap() error
}

type multiUnwrap interface {
	Unwrap() []error
}

func hasHTTPError(err error) bool {
	if _, has := err.(*HTTPError); has { //nolint:errorlint
		return true
	}

	if unwrap, ok := err.(singleUnwrap); ok { //nolint:errorlint
		return hasHTTPError(unwrap.Unwrap())
	} else if unwrap, ok := err.(multiUnwrap); ok { //nolint:errorlint
		for _, sub := range unwrap.Unwrap() {
			if hasHTTPError(sub) {
				return true
			}
		}
	}

	return false
}

func (je *JSONError) importError(err error) {
	if he, ok := err.(*HTTPError); ok { //nolint:errorlint
		je.Code = he.Code
	}

	if _, is := err.(*SilentJoinError); !is { //nolint:errorlint
		je.Messages = append(je.Messages, err.Error())
	}

	if unwrap, ok := err.(singleUnwrap); ok { //nolint:errorlint
		je.importError(unwrap.Unwrap())
	} else if unwrap, ok := err.(multiUnwrap); ok { //nolint:errorlint
		for _, sub := range unwrap.Unwrap() {
			je.importError(sub)
		}
	}
}
