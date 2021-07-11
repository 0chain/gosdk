package commitmeta

import "github.com/0chain/gosdk/zboxcore/fileref"

type CommitMetaResponse struct {
	TxnID    string
	MetaData *fileref.ConsolidatedFileMeta
}

type CommitMetaData struct {
	CrudType string
	MetaData *fileref.ConsolidatedFileMeta
}
