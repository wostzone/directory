package dirserver

import (
	"net/http"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/thingdir-go/pkg/dirclient"
	"github.com/wostzone/thingdir-go/pkg/dirstore/dirfilestore"
	"github.com/wostzone/wostlib-go/pkg/tlsserver"
)

const DirectoryPluginID = "directory"

// const RouteUpdateTD = "/things/{thingID}"
// const RouteGetTD = "/things/{thingID}"
// const RouteDeleteTD = "/things/{thingID}"
// const RoutePatchTD = "/things/{thingID}"
// const RouteListTD = "/things"
// const RouteQueryTD = "/things"

// DirectoryServer for web of things
type DirectoryServer struct {
	// config
	address        string // listening address
	caCertPath     string // path to CA certificate PEM file
	instanceID     string // ID of this service
	port           uint   // listening port
	serverCertPath string // path to server certificate PEM file
	serverKeyPath  string // path to server private key PEM file
	authenticator  func(http.ResponseWriter, *http.Request) error

	// the service name. Use dirclient.DirectoryServiceName for default or "" to disable DNS discovery
	discoveryName string

	// runtime status
	running     bool
	tlsServer   *tlsserver.TLSServer
	discoServer *zeroconf.Server
	store       *dirfilestore.DirFileStore
}

// Return the address that the server listens on
// This is automatically determined from the default network interface
func (srv *DirectoryServer) Address() string {
	return srv.address
}

// Start the server.
func (srv *DirectoryServer) Start() error {
	var err error

	if !srv.running {
		// srv.listenAddress = listenAddress
		srv.running = true

		logrus.Warningf("Starting directory server on %s:%d", srv.address, srv.port)
		// srv.caCertPEM, err = certsetup.LoadPEM(srv.caCertFolder, certsetup.CaCertFile)
		// if err != nil {
		// 	return err
		// }
		// srv.caKeyPEM, err = certsetup.LoadPEM(srv.caCertFolder, certsetup.CaKeyFile)
		// if err != nil {
		// 	return err
		// }

		// srv.address = hubconfig.GetOutboundIP("").String()
		srv.tlsServer = tlsserver.NewTLSServer(
			srv.address, srv.port,
			srv.serverCertPath, srv.serverKeyPath, srv.caCertPath, srv.authenticator)
		err = srv.tlsServer.Start()
		if err != nil {
			return err
		}
		// setup the handlers for the paths. The GET/PUT/... operations are resolved by the handler
		srv.tlsServer.AddHandler(dirclient.RouteThings, srv.ServeThings)
		srv.tlsServer.AddHandler(dirclient.RouteThingID, srv.ServeThingByID)

		if srv.discoveryName != "" {
			srv.discoServer, _ = ServeDirDiscovery(srv.instanceID, srv.discoveryName, srv.address, srv.port)
		}
		// Make sure the server is listening before continuing
		// Not pretty but it handles it
		time.Sleep(time.Second)
	}
	return nil
}

// Stop the IdProv server
func (srv *DirectoryServer) Stop() {
	if srv.running {
		srv.running = false
		logrus.Warningf("Stopping directory server on %s:%d", srv.address, srv.port)
		if srv.discoServer != nil {
			srv.discoServer.Shutdown()
			srv.discoServer = nil
		}
	}
}

// NewDirectoryServer creates a new instance of the IoT Device Provisioning Server.
//  - instanceID is the unique ID for this service used in discovery and communication
//  - storePath is the location of the directory storage file. This must be writable.
//  - address the server listening address. Typically the same address as the services
//  - port server listening port
//  - caCertFolder location of CA Cert and server certificates and keys
//  - discoveryName for use in dns-sd. Use "" to disable discover, or the dirclient.DirectoryServiceName for default
//  - authenticator authenticates the user for the request
func NewDirectoryServer(
	instanceID string,
	storePath string,
	address string,
	port uint,
	discoveryName string,
	serverCertPath string,
	serverKeyPath string,
	caCertPath string,
	authenticator func(http.ResponseWriter, *http.Request) error,
) *DirectoryServer {

	if instanceID == "" || port == 0 {
		logrus.Panic("NewDirectoryServer: Invalid arguments for instanceID or port")
		panic("Exit due to invalid args")
	}
	srv := DirectoryServer{
		address:        address,
		serverCertPath: serverCertPath,
		serverKeyPath:  serverKeyPath,
		caCertPath:     caCertPath,
		discoveryName:  discoveryName,
		instanceID:     instanceID,
		port:           port,
		store:          dirfilestore.NewDirFileStore(storePath),
		authenticator:  authenticator,
	}
	return &srv
}
