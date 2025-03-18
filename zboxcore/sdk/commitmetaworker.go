package sdk

import "github.com/0chain/gosdk_common/zboxcore/commonsdk"

type CommitMetaData struct {
	CrudType string
	MetaData *commonsdk.ConsolidatedFileMeta
}

type CommitMetaResponse struct {
	TxnID    string
	MetaData *commonsdk.ConsolidatedFileMeta
}
