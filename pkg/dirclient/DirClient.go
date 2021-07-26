// Package dirclient with client side functions to access the directory
package dirclient

import (
	"fmt"

	"github.com/wostzone/wostlib-go/pkg/td"
)

// Wost Directory service ID, used for discovery
const DirectoryServiceDiscoveryType = "wostdir"

// DirClient is a client for the WoST Directory service
// Intended for updating and reading TDs
type DirClient struct {
	address string // address of the directory server, "" if unknown
	port    uint   // port of the directory server, 0 if unknown
}

// Close the connection to the directory server
func (dc *DirClient) Close() {
}

// Open the connection to the directory server
func (dc *DirClient) Open() error {
	return fmt.Errorf("not implemented")
}

// Read the TD with the given ID
func (dc *DirClient) Read(id string) error {
	return fmt.Errorf("not implemented")
}

// Write the TD with the given ID, eg create/update
func (dc *DirClient) Write(id string, td td.ThingTD) error {
	return fmt.Errorf("not implemented")
}

// Query the TD with the given JSONPATH expression
func (dc *DirClient) Query(jsonpath string) ([]td.ThingTD, error) {
	return nil, fmt.Errorf("not implemented")
}

// Create a new instance of the directory client
//  address is the listening address of the client
func NewDirClient(address string, port uint) *DirClient {
	dc := &DirClient{
		address: address,
		port:    port,
	}
	return dc
}
