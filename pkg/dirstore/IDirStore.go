package dirstore

// Interface to the directory JSON object store
// Simple CRUD interface with JSONPATH support
type IDirStore interface {
	// Close the store
	Close()
	// Get a document by its ID
	// Returns an error if it doesn't exist
	Get(id string) (interface{}, error)

	// Get a list of documents
	//  offset to start
	//  maximum nr of documents to return
	List(offset int, limit int) []interface{}

	// Open the store
	// Returns error if it can't be opened or already open
	Open() error

	// Patch part of a document
	// Returns an error if it doesn't exist
	Patch(id string, doc map[string]interface{}) error

	// Query for documents using JSONPATH
	//  offset to return the results
	//  maximum nr of documents to return
	// Returns list of documents by their ID, or error if jsonPath is invalid
	Query(jsonPath string, offset int, limit int) ([]interface{}, error)

	// Remove a document
	// Succeeds if the document doesn't exist
	Remove(id string) error

	// Replace a document
	// The document does not have to exist
	Replace(id string, document map[string]interface{}) error
}
