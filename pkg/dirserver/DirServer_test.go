package internal_test

import (
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/directory/internal"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

const testDirectoryPort = 9990
const testDirectoryServiceID = "directory"
const testDiscoveryType = "_test._wost_directory._tcp"

// These are set in TestMain
var serverCertFolder string
var certStoreFolder string
var clientCertFolder string

var directoryTestAddr string
var homeFolder string
var directoryServer *internal.DirectoryServer

// easy cleanup for existing device certificate
func removeDeviceCerts() {
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(clientCertFolder, "*.pem")).Output()
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(certStoreFolder, "*.pem")).Output()
}

// func removeServerCerts() {
// 	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(serverCertFolder, "*.pem")).Output()
// }

// TestMain runs a idProv server, gets the directory for futher calls
// Used for all test cases in this package
// NOTE: Don't run tests in parallel as each test creates and deletes certificates
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryServer ------")
	address := hubconfig.GetOutboundIP("").String()
	hostnames := []string{address}

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	serverCertFolder = path.Join(homeFolder, "certs")
	certStoreFolder = path.Join(homeFolder, "certstore")

	// Start test with new certificates
	logrus.Infof("Creating certificate bundle for names: %s", hostnames)
	certsetup.CreateCertificateBundle(hostnames, serverCertFolder)

	directoryServer = internal.NewDirectoryServer(testDirectoryServiceID, address,
		testDirectoryPort, serverCertFolder, testDiscoveryType)
	directoryServer.Start()
	// testDirectoryAddr := directoryServer.Address()
	res := m.Run()
	directoryServer.Stop()
	time.Sleep(time.Second)
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	// start without existing client cert
	// deviceID1 := "device1"
	dirClient := directory.NewDirectoryClient(testDirectoryAddr, testDirectoryPort)

	// Client start only succeeds if server is running
	err := dirClient.Start()
	assert.NoError(t, err)

	dirClient.Stop()
}
