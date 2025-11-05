package biz

import (
	"strings"
)

type myerror struct {
	s string
}

func (me myerror) Error() string {
	return me.s
}

// NewMyError constructs a value-type myerror. Using a value type makes
// chaining WithMessage() calls convenient because the static type
// exposes the WithMessage method (we frequently use ErrInvaildArgument.WithMessage(...)).
func NewMyError(s string) myerror {
	return myerror{
		s: s,
	}
}

// WithMessage appends the provided message to the error string and returns a new myerror.
func (me myerror) WithMessage(str string) myerror {
	var builder strings.Builder

	builder.WriteString(me.s)
	builder.WriteString(": ")
	builder.WriteString(str)
	return myerror{
		s: builder.String(),
	}

}

var (
	// Note: we intentionally use the concrete myerror type for these
	// variables so callers can call WithMessage directly, e.g.
	// ErrInvaildArgument.WithMessage("origin cannot be empty").
	ErrInvaildArgument myerror = NewMyError("invaild argument")
	ErrNotFound        myerror = NewMyError("not found")
	ErrInternalError   myerror = NewMyError("internal error")
)
