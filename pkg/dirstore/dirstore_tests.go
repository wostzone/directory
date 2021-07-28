// Package dirstore with test cases for store implementations
package dirstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Generic directory store testcases, invoked by specific implementation (eg dirfilestore)
func DirStoreStartStop(t *testing.T, store IDirStore) {
	err := store.Open()
	assert.NoError(t, err)
	store.Close()
}

func DirStoreCrud(t *testing.T, store IDirStore) {
	thingID := "thing1"
	thingTD1 := "this is a td"
	thingTD2 := "this is a td updated"
	err := store.Open()
	assert.NoError(t, err)
	// Create
	err = store.Create(thingID, thingTD1)
	assert.NoError(t, err)
	// Read
	td2, err := store.Read(thingID)
	assert.NoError(t, err)
	assert.Equal(t, thingTD1, td2)
	// Update
	err = store.Update(thingID, thingTD2)
	assert.NoError(t, err)
	td2, err = store.Read(thingID)
	assert.NoError(t, err)
	assert.Equal(t, thingTD2, td2)

	// Delete
	err = store.Delete(thingID)
	assert.NoError(t, err)
	_, err = store.Read(thingID)
	assert.Error(t, err)

	store.Close()
}
