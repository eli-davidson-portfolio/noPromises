package validation

import "errors"

var (
	ErrEmptyConfig       = errors.New("empty configuration")
	ErrMissingID         = errors.New("missing flow ID")
	ErrInvalidNodes      = errors.New("invalid nodes configuration")
	ErrInvalidNodeConfig = errors.New("invalid node configuration")
	ErrMissingNodeType   = errors.New("missing node type")
	ErrInvalidNodeType   = errors.New("invalid node type")
)
