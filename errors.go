package memap

import "github.com/memap-project/memap-core/ns"

// Re-exported sentinel errors so callers do not need to import the internal ns package.
var (
	ErrNamespaceAlreadyExists = ns.ErrNamespaceAlreadyExists
	ErrNamespaceNotFound      = ns.ErrNamespaceNotFound
	ErrNamespaceLimitExceeded = ns.ErrNamespaceLimitExceeded
	ErrInvalidNamespace       = ns.ErrInvalidNamespace
	ErrKeyNotFound            = ns.ErrKeyNotFound
	ErrKeyAlreadyExists       = ns.ErrKeyAlreadyExists
	ErrKeyTooLarge            = ns.ErrKeyTooLarge
	ErrValueTooLarge          = ns.ErrValueTooLarge
	ErrBufferEmpty            = ns.ErrBufferEmpty
	ErrIndexOutOfBounds       = ns.ErrIndexOutOfBounds
	ErrLimitExceeded          = ns.ErrLimitExceeded
	ErrFieldNotFound          = ns.ErrFieldNotFound
	ErrInvalidCapacity        = ns.ErrInvalidCapacity
	ErrInvalidTTL             = ns.ErrInvalidTTL
	ErrInvalidArg             = ns.ErrInvalidArg
)
