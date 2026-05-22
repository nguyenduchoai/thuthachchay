package steps

import "errors"

var (
	ErrNonceReplay   = errors.New("client_nonce đã được dùng")
	ErrInvalidSource = errors.New("source phải là zmp hoặc strava")
)
