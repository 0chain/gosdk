package sdk

import (
	"path/filepath"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

type Operationer interface {
	Process() (allocationchange.AllocationChange, error)
	Verify() error
	Build() error
}

type CopyOperation struct {
	alloc             *Allocation
	srcPath, destPath string
}

func (co *CopyOperation) Process() {
	copyRequest := CopyRequest{}
	for _, b := range co.alloc.Blobbers {
		go copyRequest.copyBlobberObject(b)

	}
}

type RenameOperation struct {
	alloc    *Allocation
	SrcPath  string
	DestName string
}

func (mo *RenameOperation) Verify() error {

	srcName := filepath.Base(mo.SrcPath)
	if srcName == mo.DestName {
		return errors.New("", "")
	}
	mo.alloc.CanRename()
	return nil
}

type UploadOperation struct {
	// fields
}

func (uo *UploadOperation) ProcessUpload() {
	u := ChunkedUpload{}
	u.processUpload()
}

type MultiOperation struct {
	ConnectionID string
	Operations   []Operationer
	changes      []allocationchange.AllocationChange
}

// 1. copy
// 2. move
// 3. delete
// 4. upload
// 5. upload
// 6. upload
// 7. upload
// 8. upload
func (mo *MultiOperation) Process() {
	wg := sync.WaitGroup{}
	for _, op := range mo.Operations {
		go func() {
			change, err := op.Process()
			if err != nil {
				// cancel all operations or based on user skip this operation and proceed
				_ = err
			}

			// User can request blobber to delete temporary files as well.
			mo.changes = append(mo.changes, change)
		}()
	}
	wg.Wait()

	// rootRef := "get root ref from blobber" // GetReferencePath

	var rootRef *fileref.Ref
	//// /a/b/c.txt
	// {"/a":"uuid1", "/a/b":"uuid2", "/a/b/c.txt":"uuid3"}
	var commitIdMeta map[string]string // be careful as it would create two uuid for same path

	for _, change := range mo.changes {
		_, err := change.ProcessChange(rootRef, commitIdMeta)
		if err != nil {
			return
		}

	}

	// Writemarker lock
	// commit
	// writemarker unlock

}
