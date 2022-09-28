package sdk

type CommitMetaData struct {
	CrudType string
	MetaData *ConsolidatedFileMeta
}

type CommitMetaResponse struct {
	TxnID    string
	MetaData *ConsolidatedFileMeta
}
