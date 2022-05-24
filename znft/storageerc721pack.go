package znft

import (
	"context"

	storageerc721pack "github.com/0chain/gosdk/znft/contracts/dstorageerc721pack/binding"
)

type IStorageECR721Pack interface {
	IStorageECR721
}

type StorageECR721Pack struct {
	StorageECR721
	session *storageerc721pack.BindingsSession
	ctx     context.Context
}
