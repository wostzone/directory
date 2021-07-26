package dirfilestore_test

import (
	"testing"

	"github.com/wostzone/wostdir/pkg/dirstore"
	"github.com/wostzone/wostdir/pkg/dirstore/dirfilestore"
)

func makeFileStore() *dirfilestore.DirFileStore {
	filename := "/tmp/test-dirfilestore.json"

	store := dirfilestore.NewDirFileStore(filename)
	return store
}

// Generic directory store testcases
func TestFileStoreStartStop(t *testing.T) {
	fileStore := makeFileStore()
	dirstore.DirStoreStartStop(t, fileStore)
}

func TestFileStoreWrite(t *testing.T) {
	fileStore := makeFileStore()
	dirstore.DirStoreCrud(t, fileStore)
}
