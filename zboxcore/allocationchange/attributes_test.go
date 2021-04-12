package allocationchange

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestAttributesChange_ProcessChange(t *testing.T) {
	change := change{Size: rand.Int63(), NumBlocks: 1, Operation: UPDATE_OPERATION}
	connectionID := zboxutil.NewConnectionId()
	allocationID := ""
	path := "/alloc/folder/1.txt"
	attributes := fileref.Attributes{WhoPaysForReads: common.WhoPaysOwner}
	ac := &AttributesChange{
		change:       change,
		ConnectionID: connectionID,
		AllocationID: allocationID,
		Path:         path,
		Attributes:   attributes,
	}

	t.Run("Test_Success", func(t *testing.T) {
		err := ac.ProcessChange(&fileref.Ref{
			Type: fileref.DIRECTORY,
			Name: "/",
			Children: []fileref.RefEntity{
				&fileref.Ref{Type: fileref.DIRECTORY,
					Name: "alloc",
					Children: []fileref.RefEntity{
						&fileref.Ref{
							Type: fileref.DIRECTORY,
							Name: "folder",
							Children: []fileref.RefEntity{
								&fileref.FileRef{
									Ref: fileref.Ref{
										Type: fileref.FILE,
										Name: "1.txt",
										Path: path,
									},
								},
							},
						},
					},
				},
			},
		})
		assert.NoError(t, err, "unexpected ac.ProcessChange() error but got: %v", err)
	})

	t.Run("Test_Error_Invalid_Referrence_Path_Failed", func(t *testing.T) {
		err := ac.ProcessChange(&fileref.Ref{
			Type: fileref.DIRECTORY,
			Name: "/",
			Children: []fileref.RefEntity{
				&fileref.Ref{Type: fileref.DIRECTORY,
					Name:     "alloc",
					Children: []fileref.RefEntity{
						&fileref.Ref{
							Type: fileref.DIRECTORY,
							Name: "folder",
							Children: []fileref.RefEntity{},
						},
					},
				},
			},
		})
		assert.Error(t, err, "expected ac.ProcessChange() error != nil ")
	})

	t.Run("Test_Error_File_Which_Attribute_Update_For_Are_Not_Found_Failed", func(t *testing.T) {
		err := ac.ProcessChange(&fileref.Ref{
			Type: fileref.DIRECTORY,
			Name: "/",
			Children: []fileref.RefEntity{
				&fileref.Ref{Type: fileref.DIRECTORY,
					Name:     "alloc",
					Children: []fileref.RefEntity{},
				},
			},
		})
		assert.Error(t, err, "expected ac.ProcessChange() error != nil ")
	})
}

func TestAttributesChange_GetAffectedPath(t *testing.T) {
	path := "/alloc/folder/1.txt"
	ac := &AttributesChange{
		Path: path,
	}
	assert.Equal(t, path, ac.GetAffectedPath())
}

func TestAttributesChange_GetSize(t *testing.T) {
	ac := &AttributesChange{}
	assert.Equal(t, int64(0), ac.GetSize())
}
