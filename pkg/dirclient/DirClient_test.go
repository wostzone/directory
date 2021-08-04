package dirclient_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/thingdir-go/pkg/dirclient"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubnet"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/tlsserver"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

var testDirectoryAddr string

const testDirectoryPort = 9990

var serverCertFolder string
var caCertPath string
var pluginCertPath string
var pluginKeyPath string

/*
 */
func startTestServer() *tlsserver.TLSServer {
	serverCertPath := path.Join(serverCertFolder, certsetup.HubCertFile)
	serverKeyPath := path.Join(serverCertFolder, certsetup.HubKeyFile)
	server := tlsserver.NewTLSServer(testDirectoryAddr, testDirectoryPort, serverCertPath, serverKeyPath, caCertPath)

	server.Start()
	return server
}

func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryClient ------")
	testDirectoryAddr = hubnet.GetOutboundIP("").String()
	hostnames := []string{testDirectoryAddr}

	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	serverCertFolder = path.Join(homeFolder, "certs")
	// certStoreFolder = path.Join(homeFolder, "certstore")
	// storePath := path.Join(homeFolder, "config")

	// Start test with new certificates
	logrus.Infof("Creating certificate bundle for names: %s", hostnames)
	certsetup.CreateCertificateBundle(hostnames, serverCertFolder)
	pluginCertPath = path.Join(serverCertFolder, certsetup.PluginCertFile)
	pluginKeyPath = path.Join(serverCertFolder, certsetup.PluginKeyFile)
	caCertPath = path.Join(serverCertFolder, certsetup.CaCertFile)

	res := m.Run()
	os.Exit(res)
}

func TestOpenClose(t *testing.T) {
	// launch a server to receive requests
	server := startTestServer()
	//
	dirClient := dirclient.NewDirClient(testDirectoryAddr, testDirectoryPort, caCertPath)

	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	assert.NoError(t, err)

	dirClient.Close()
	server.Stop()
}

func TestUpdateTD(t *testing.T) {
	var receivedTD td.ThingTD
	var body []byte
	var err2 error
	const id1 = "thing1"

	server := startTestServer()
	server.AddHandler(dirclient.RouteThingID, func(response http.ResponseWriter, request *http.Request) {
		logrus.Infof("TestUpdateTD: %s %s", request.Method, request.RequestURI)

		if request.Method == "POST" {
			body, err2 = ioutil.ReadAll(request.Body)
			if err2 == nil {
				err2 = json.Unmarshal(body, &receivedTD)
			}
		} else if request.Method == "GET" {
			parts := strings.Split(request.URL.Path, "/")
			id := parts[len(parts)-1]
			assert.Equal(t, id1, id)
			//return the previously sent td
			msg, _ := json.Marshal(receivedTD)
			response.Write(msg)
		}
	})

	dirClient := dirclient.NewDirClient(testDirectoryAddr, testDirectoryPort, caCertPath)
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	require.NoError(t, err)

	// write a TD document
	td := td.CreateTD(id1, vocab.DeviceTypeSensor)
	err = dirClient.UpdateTD(id1, td)
	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.Equal(t, id1, receivedTD["id"])

	//
	receivedTD2, err := dirClient.GetTD(id1)
	assert.NoError(t, err)
	assert.Equal(t, id1, receivedTD2["id"])

	dirClient.Close()
	server.Stop()

}

func TestQueryAndList(t *testing.T) {
	const query = "$.hello.world"
	server := startTestServer()
	// if no thingID is specified then this is a request for a list or a query
	server.AddHandler(dirclient.RouteThings, func(response http.ResponseWriter, request *http.Request) {
		logrus.Infof("TestQuery: %s %s", request.Method, request.RequestURI)

		if request.Method == "GET" {
			q := request.URL.Query().Get(dirclient.ParamQuery)
			thd := td.CreateTD("thing1", vocab.DeviceTypeSensor)
			prop := td.CreateProperty("query", "", vocab.PropertyTypeAttr)
			td.SetPropertyDataTypeString(prop, 0, 0)
			td.SetPropertyValue(prop, q)
			td.AddTDProperty(thd, dirclient.ParamQuery, prop)
			tdList := []td.ThingTD{thd}
			data, _ := json.Marshal(tdList)
			response.Write(data)
		} else {
			server.WriteBadRequest(response, "Only GET is supported")
		}
	})

	dirClient := dirclient.NewDirClient(testDirectoryAddr, testDirectoryPort, caCertPath)
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	require.NoError(t, err)

	// test query
	td2, err := dirClient.QueryTDs(query, 0, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, td2)
	val, _ := td.GetPropertyValue(td2[0], dirclient.ParamQuery)
	assert.Equal(t, query, val)

	// test list
	td3, err := dirClient.ListTDs(0, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, td3)

	dirClient.Close()
	server.Stop()
}

func TestDelete(t *testing.T) {
	const thingID1 = "thing1"
	var idToDelete string

	server := startTestServer()
	server.AddHandler(dirclient.RouteThingID, func(response http.ResponseWriter, request *http.Request) {
		if request.Method == "DELETE" {
			parts := strings.Split(request.URL.Path, "/")
			idToDelete = parts[len(parts)-1]
		} else {
			server.WriteBadRequest(response, "wrong method: "+request.Method)
		}
	})

	dirClient := dirclient.NewDirClient(testDirectoryAddr, testDirectoryPort, caCertPath)
	err := dirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	require.NoError(t, err)

	err = dirClient.Delete(thingID1)
	require.NoError(t, err)
	assert.Equal(t, thingID1, idToDelete)

	dirClient.Close()
	server.Stop()
}
