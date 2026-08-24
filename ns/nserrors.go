package ns

import (
	"errors"

	"github.com/memap-project/memap-core/shard"
)

var (
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrKeyNotFound            = errors.New("key not found")
	ErrKeyAlreadyExists       = errors.New("key already exists")
	ErrBufferEmpty            = errors.New("buffer is empty")
	ErrIndexOutOfBounds       = errors.New("index out of bounds")
	ErrLimitExceeded          = errors.New("limit exceeded")
	ErrFieldNotFound          = errors.New("field not found")
)

func statusToError(status shard.Status) error {
	switch status {
	case shard.StatusSuccess:
		return nil
	case shard.StatusNotFound, shard.StatusExpired:
		return ErrKeyNotFound
	case shard.StatusBufferEmpty:
		return ErrBufferEmpty
	case shard.StatusIndexOutOfBounds:
		return ErrIndexOutOfBounds
	case shard.StatusLimitExceeded:
		return ErrLimitExceeded
	case shard.StatusFieldNotFound:
		return ErrFieldNotFound
	default:
		return ErrKeyNotFound
	}
}
