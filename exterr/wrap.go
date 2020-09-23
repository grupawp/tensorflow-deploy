package exterr

import (
	"github.com/pkg/errors"
)

func WrapWithErr(oldErr, newErr error) error {

	if oldErr == nil || newErr == nil {
		return nil
	}

	if _, ok := oldErr.(*Error); !ok {
		oldErr = wrapWithFrame(oldErr, 3)
	}

	if newError, ok := newErr.(*Error); ok {
		newError.Cause = oldErr
		newError.setFrame()
		return newError
	}

	return errors.Wrap(oldErr, newErr.Error())
}

func wrapWithFrame(oldErr error, skip int) error {
	if oldErr == nil {
		return nil
	}

	newErr := Error{}
	newErr.setFrameWithSkip(skip)
	newErr.Cause = oldErr

	if oldError, ok := oldErr.(*Error); ok {
		newErr.Err = oldError.Err
	} else {
		newErr.Err = oldErr.Error()
	}
	return &newErr
}

func WrapWithFrame(oldErr error) error {
	return wrapWithFrame(oldErr, 3)
}

func (e *Error) Unwrap() error {
	return e.Cause
}
