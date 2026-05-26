package sdk

import (
	"testing"

	"github.com/0chain/gosdk/zboxcore/fileref"
)

// WithPrefetchedFileMeta must set the field that lets getFileRef skip the
// redundant getFileMetaFromBlobbers round-trip.
func TestWithPrefetchedFileMeta_SetsField(t *testing.T) {
	fr := &fileref.FileRef{}
	dr := &DownloadRequest{}
	if dr.prefetchedFileMeta != nil {
		t.Fatal("field should start nil")
	}
	WithPrefetchedFileMeta(fr)(dr)
	if dr.prefetchedFileMeta != fr {
		t.Fatal("WithPrefetchedFileMeta did not set prefetchedFileMeta")
	}
}

// When a file ref is prefetched, getFileRef returns it directly with NO blobber
// round-trip (the whole point of the fix). Reaching the network would block/panic
// here since no blobbers are configured — returning fast proves the skip.
func TestGetFileRef_PrefetchedSkipsFetch(t *testing.T) {
	fr := &fileref.FileRef{}
	fr.Type = fileref.FILE
	fr.Name = "obj"
	fr.ActualFileSize = 100

	dr := &DownloadRequest{prefetchedFileMeta: fr}
	got, err := dr.getFileRef()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != fr {
		t.Fatal("getFileRef did not return the prefetched ref")
	}
}

// A prefetched directory ref must be rejected, same as the fetched path.
func TestGetFileRef_PrefetchedDirectoryErrors(t *testing.T) {
	fr := &fileref.FileRef{}
	fr.Type = fileref.DIRECTORY

	dr := &DownloadRequest{prefetchedFileMeta: fr}
	if _, err := dr.getFileRef(); err == nil {
		t.Fatal("expected error downloading a directory ref")
	}
}
