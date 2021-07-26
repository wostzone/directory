package dirstore

// Interface to the directory JSON object store
// Simple CRUD interface with JSONPATH support
type IDirStore interface {
	// Close the store
	Close()

	// Create a document
	// Returns an error if it already exists
	Create(id string, document interface{}) error

	// Delete a document
	// Succeeds if the document doesn't exist
	Delete(id string) error

	// Open the store
	// Returns error if it can't be opened or already open
	Open() error

	// Query for documents using JSONPATH
	// Returns list of documents by their ID, or error if jsonPath is invalid
	Query(jsonPath string) ([]interface{}, error)

	// Read a document by its ID
	// Returns an error if it doesn't exist
	Read(id string) (interface{}, error)

	// Update a document
	// Returns an error if it doesn't exist
	Update(id string, doc interface{}) error
}
