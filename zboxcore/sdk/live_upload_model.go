package sdk

import "github.com/0chain/gosdk/zboxcore/fileref"

// LiveMeta metadata of live stream input
type LiveMeta struct {
	// Mimetype mime type of source file
	MimeType string

	//RemoteName remote file name
	RemoteName string
	// RemotePath remote path
	RemotePath string
	// Attributes file attributes in blockchain
	Attributes fileref.Attributes
}
