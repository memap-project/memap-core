package ns

import (
	"errors"

	"github.com/memap-project/memap-core/shard"
)

// Namespace and storage operation errors.
var (
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrNamespaceLimitExceeded = errors.New("namespace limit exceeded")
	ErrKeyNotFound            = errors.New("key not found")
	ErrKeyAlreadyExists       = errors.New("key already exists")
	ErrKeyTooLarge            = errors.New("key exceeds maximum allowed size")
	ErrValueTooLarge          = errors.New("value exceeds maximum allowed size")
	ErrBufferEmpty            = errors.New("buffer is empty")
	ErrIndexOutOfBounds       = errors.New("index out of bounds")
	ErrLimitExceeded          = errors.New("limit exceeded")
	ErrFieldNotFound          = errors.New("field not found")
	ErrInvalidCapacity        = errors.New("invalid capacity")
	ErrInvalidTTL             = errors.New("invalid ttl")
	ErrInvalidArg             = errors.New("invalid argument")
)

// statusToError converts a shard.Status code to a corresponding error.
func statusToError(status shard.Status) error {
	switch status {
	case shard.StatusSuccess:
		return nil
	case shard.StatusNotFound, shard.StatusExpired:
		return ErrKeyNotFound
	case shard.StatusAlreadyExists:
		return ErrKeyAlreadyExists
	case shard.StatusBufferEmpty:
		return ErrBufferEmpty
	case shard.StatusIndexOutOfBounds:
		return ErrIndexOutOfBounds
	case shard.StatusLimitExceeded:
		return ErrLimitExceeded
	case shard.StatusFieldNotFound:
		return ErrFieldNotFound
	case shard.StatusInvalidCapacity:
		return ErrInvalidCapacity
	case shard.StatusInvalidTTL:
		return ErrInvalidTTL
	case shard.StatusInvalidArg:
		return ErrInvalidArg
	default:
		return ErrKeyNotFound
	}
}
