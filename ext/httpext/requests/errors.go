package requests

import "errors"

func Is404NoRetryError(err error) bool {
	return errors.Is(err, ErrNotFoundNoRetry)
}
