package znft

import (
	"context"

	storageerc721random "github.com/0chain/gosdk/znft/contracts/dstorageerc721random/binding"
)

type IStorageECR721Random interface {
	IStorageECR721
}

type StorageECR721Random struct {
	StorageECR721
	session *storageerc721random.BindingsSession
	ctx     context.Context
}
