// Package dirfilestore
// This is just a simple in-memory store that is loaded from file and written regularly after updates.
//
// The jsonpath query feature is provided by a library that works with the in-memory object store.
// A good overview of implementations can be found here:
// > https://cburgmer.github.io/json-path-comparison/
//
// Two good options for jsonpath queries:
//  > github.com/ohler55/ojg/jp
//  > github.com/PaesslerAG/jsonpath

package dirfilestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
)

// DirFileStore is a crude little file based Directory store
// Intended as a testing MVP for the directory service
// Implements the IDirStore interface
type DirFileStore struct {
	docs      map[string]interface{} // documents by ID
	changed   bool
	storePath string
	running   bool
	mutex     sync.Mutex
}

func createStoreFile(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// new file
		fp, err2 := os.Create(filePath)
		if err2 == nil {
			fp.Chmod(0600)
			fp.Write([]byte("{}"))
			fp.Close()
		}
		err = err2
	}
	if err != nil {
		logrus.Errorf("createStoreFile: %s", err)
	}
	return err
}

// createStoreFolder creates the folder if it only exists
// The parent folder must exist otherwise this fails
func createStoreFolder(storeFolder string) error {
	_, err := os.Stat(storeFolder)
	if os.IsNotExist(err) {
		err = os.Mkdir(storeFolder, os.ModeDir)
	}
	if err != nil {
		logrus.Errorf("createStoreFolder. Error %s", err)
	}
	return err
}

// readStoreFile loads the store JSON content into a map
func readStoreFile(storePath string) (docs map[string]interface{}, err error) {
	var rawData []byte
	rawData, err = os.ReadFile(storePath)

	if err == nil {
		err = json.Unmarshal(rawData, &docs)
	}

	if err != nil {
		logrus.Infof("DirFileStore.readStoreFile: failed read store '%s', error %s", storePath, err)
	}
	return docs, err
}

// saveStoreFile writes the store to file
func saveStoreFile(storePath string, docs map[string]interface{}) error {
	rawData, err := json.Marshal(docs)
	if err == nil {
		// only allow this user access
		err = os.WriteFile(storePath, rawData, 0600)
	}
	if err != nil {
		logrus.Errorf("DirFileStore.save: Error while saving store to %s: %s", storePath, err)
	}
	return err
}

// Close the store
func (store *DirFileStore) Close() {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.changed {
		saveStoreFile(store.storePath, store.docs)
	}
	store.running = false
}

// Create a document
// Returns an error if it already exists
func (store *DirFileStore) Create(id string, document interface{}) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.docs[id] = document
	store.changed = true
	return nil
}

// Delete a document
// Succeeds if the document doesn't exist
func (store *DirFileStore) Delete(id string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	delete(store.docs, id)
	store.changed = true
	return nil
}

// Open the store
// Returns error if it can't be opened or already open
func (store *DirFileStore) Open() error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	// create the folder if needed
	storeFolder := path.Dir(store.storePath)
	err := createStoreFolder(storeFolder)
	if err == nil {
		err = createStoreFile(store.storePath)
	}
	if err == nil {
		store.docs, err = readStoreFile(store.storePath)
	}
	store.running = true
	return err
}

// Query for documents using JSONPATH
// Eg `$[? @.properties.deviceType=="sensor"]`
func (store *DirFileStore) Query(jsonPath string) (interface{}, error) {
	//  "github.com/PaesslerAG/jsonpath" - just works, amazing!
	// Unfortunately no filter with bracket notation $[? @.["title"]=="my title"]
	// res, err := jsonpath.Get(jsonPath, store.docs)
	// github.com/ohler55/ojg/jp - seems to work with in-mem maps, no @token in bracket notation
	jpExpr, err := jp.ParseString(jsonPath)
	if err != nil {
		return nil, err
	}
	res := jpExpr.Get(store.docs)

	// The following parsers don't work on an in-memory store :/
	//  They would need to parse the whole tree after each update.
	// github.com/vmware-labs/yaml-jsonpath - requires a yaml node

	return res, err
}

// Read a document by its ID
// Returns an error if it doesn't exist
func (store *DirFileStore) Read(id string) (interface{}, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	doc, ok := store.docs[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return doc, nil
}

// Update a document
// Returns an error if it doesn't exist
func (store *DirFileStore) Update(id string, doc interface{}) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.docs[id] = doc
	return nil
}

// Create a new directory file store instance
//  filePath path to JSON store file
func NewDirFileStore(jsonFilePath string) *DirFileStore {
	store := DirFileStore{
		docs:      make(map[string]interface{}),
		storePath: jsonFilePath,
	}
	return &store
}
