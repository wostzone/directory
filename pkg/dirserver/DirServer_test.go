package dirserver_test

import (
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/thingdir/pkg/dirclient"
	"github.com/wostzone/thingdir/pkg/dirserver"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubnet"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

const testDirectoryPort = 9990
const testDirectoryServiceID = "directory"
const testDiscoveryType = "_test._wost_directory._tcp"

// These are set in TestMain
var serverCertFolder string
var serverAddress string

var homeFolder string
var caCertPath string
var directoryServer *dirserver.DirectoryServer
var pluginCertPath string
var pluginKeyPath string

// TD's for testing. Expect 2 sensors in this list
var tdDefs = []struct {
	id         string
	deviceType vocab.DeviceType
	name       string
}{
	{"thing1", vocab.DeviceTypeBeacon, "a beacon"},
	{"thing2", vocab.DeviceTypeSensor, "hallway sensor"},
	{"thing3", vocab.DeviceTypeSensor, "garage sensor"},
	{"thing4", vocab.DeviceTypeNetSwitch, "main switch"},
}

// Add a bunch of TDs
func AddTds(client *dirclient.DirClient) {
	for _, tdDef := range tdDefs {
		td1 := td.CreateTD(tdDef.id, tdDef.deviceType)
		td.AddTDProperty(td1, "name", td.CreateProperty(tdDef.name, "", vocab.PropertyTypeAttr))
		client.UpdateTD(tdDef.id, td1)
	}
}

// TestMain runs a directory server for use by the test cases in this package
// This uses the directory client in testing
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryServer ------")
	serverAddress = hubnet.GetOutboundIP("").String()
	hostnames := []string{serverAddress}

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	serverCertFolder = path.Join(homeFolder, "certs")
	// certStoreFolder = path.Join(homeFolder, "certstore")
	storePath := path.Join(homeFolder, "config")

	// make sure the certificates are there
	certsetup.CreateCertificateBundle(hostnames, serverCertFolder)
	serverCertPath := path.Join(serverCertFolder, certsetup.HubCertFile)
	serverKeyPath := path.Join(serverCertFolder, certsetup.HubKeyFile)
	caCertPath = path.Join(serverCertFolder, certsetup.CaCertFile)
	pluginCertPath = path.Join(serverCertFolder, certsetup.PluginCertFile)
	pluginKeyPath = path.Join(serverCertFolder, certsetup.PluginKeyFile)

	directoryServer = dirserver.NewDirectoryServer(
		testDirectoryServiceID, storePath, serverAddress, testDirectoryPort,
		testDiscoveryType, serverCertPath, serverKeyPath, caCertPath)
	directoryServer.Start()

	res := m.Run()

	directoryServer.Stop()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	dirClient := dirclient.NewDirClient(serverAddress, testDirectoryPort, caCertPath)

	a := directoryServer.Address()
	assert.Equal(t, serverAddress, a)

	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	assert.NoError(t, err)

	dirClient.Close()
	directoryServer.Stop()
}

func TestUpdate(t *testing.T) {
	thingID1 := "thing1"
	deviceType1 := vocab.DeviceTypeSensor

	dirClient := dirclient.NewDirClient(serverAddress, testDirectoryPort, caCertPath)
	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	require.NoError(t, err)

	// Create
	td1 := td.CreateTD(thingID1, deviceType1)
	err = dirClient.UpdateTD(thingID1, td1)
	assert.NoError(t, err)

	// get result
	td2, err := dirClient.GetTD(thingID1)
	assert.NoError(t, err)
	assert.Equal(t, td1["id"], td2["id"])

	dirClient.Close()
}

func TestPatch(t *testing.T) {

	dirClient := dirclient.NewDirClient(serverAddress, testDirectoryPort, caCertPath)

	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	require.NoError(t, err)

	AddTds(dirClient)

	// Change the device type to sensor using patch
	thingID1 := tdDefs[0].id
	td1 := td.CreateTD(thingID1, vocab.DeviceTypeSensor)
	td.AddTDProperty(td1, "name", td.CreateProperty("name1", "just a name", vocab.PropertyTypeAttr))

	err = dirClient.PatchTD(thingID1, td1)
	assert.NoError(t, err)
	props1 := td1["properties"].(map[string]interface{})
	nameProp1 := props1["name"].(map[string]interface{})
	nameProp1val := nameProp1["title"]

	// check result
	td2, err := dirClient.GetTD(thingID1)
	assert.NoError(t, err)
	assert.Equal(t, td1["id"], td2["id"])
	assert.Equal(t, string(vocab.DeviceTypeSensor), td2["@type"])
	props2 := td2["properties"].(map[string]interface{})
	nameProp2 := props2["name"].(map[string]interface{})
	nameProp2val := nameProp2["title"]
	assert.NotEmpty(t, nameProp2val)
	assert.Equal(t, nameProp1val, nameProp2val)
	dirClient.Close()
}

func TestQueryAndList(t *testing.T) {
	const query = `$[?(@['@type']=='sensor')]`

	dirClient := dirclient.NewDirClient(serverAddress, testDirectoryPort, caCertPath)
	dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	AddTds(dirClient)

	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	assert.NoError(t, err)

	// expect 2 sensors
	td2, err := dirClient.QueryTDs(query, 0, 99999)
	require.NoError(t, err)
	assert.Equal(t, 2, len(td2))

	// test list
	td3, err := dirClient.ListTDs(0, 99999)
	require.NoError(t, err)
	assert.Equal(t, len(tdDefs), len(td3))

	dirClient.Close()
}

func TestDelete(t *testing.T) {
	dirClient := dirclient.NewDirClient(serverAddress, testDirectoryPort, caCertPath)
	dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	AddTds(dirClient)

	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	assert.NoError(t, err)

	// expect 4 items
	tds, err := dirClient.ListTDs(0, 0)
	require.NoError(t, err)
	assert.Equal(t, len(tdDefs), len(tds))

	// remove 1 sensor
	err = dirClient.Delete(tdDefs[1].id)
	assert.NoError(t, err)
	tds, err = dirClient.ListTDs(0, 0)
	require.NoError(t, err)
	assert.Equal(t, len(tdDefs)-1, len(tds))

	// deleting a non existing ID is not an error
	err = dirClient.Delete("notavalidID")
	require.NoError(t, err)

	dirClient.Close()
}

func TestBadRequest(t *testing.T) {
	const query = `$[?(badquery@['@type']=='sensor')]`

	dirClient := dirclient.NewDirClient(serverAddress, testDirectoryPort, caCertPath)
	dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)

	_, err := dirClient.QueryTDs(query, 0, 99999)
	assert.Error(t, err)

	// test list
	_, err = dirClient.ListTDs(-1, 0)
	require.Error(t, err)

	_, err = dirClient.GetTD("notavalidID")
	require.Error(t, err)

}
