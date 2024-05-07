package repo

import "errors"

var (
	ErrWithdrawUnavailable = errors.New("not enough funds")
)
