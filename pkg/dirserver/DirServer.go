package dirserver

import (
	"time"

	"github.com/gorilla/mux"
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostdir/pkg/dirstore/dirfilestore"
	"github.com/wostzone/wostlib-go/pkg/tlsserver"
)

const DirectoryPluginID = "directory"

const RouteUpdateTD = "/things/{thingID}"
const RouteGetTD = "/things/{thingID}"
const RouteDeleteTD = "/things/{thingID}"
const RoutePatchTD = "/things/{thingID}"
const RouteListTD = "/things"
const RouteQueryTD = "/things"

// DirectoryServer for web of things
type DirectoryServer struct {
	// config
	address      string // listening address
	caCertFolder string //
	certStore    string // folder where client certificates are stored
	instanceID   string // ID of this service
	port         uint   // listening port

	// the service type. Use  dirclient.DirectoryServiceDiscoveryType for default or "" to disable DNS discovery
	discoveryType string

	// runtime status
	running     bool
	tlsServer   *tlsserver.TLSServer
	router      *mux.Router
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
		srv.tlsServer = tlsserver.NewTLSServer(srv.address, srv.port, srv.caCertFolder)
		err = srv.tlsServer.Start()
		if err != nil {
			return err
		}
		// setup the handlers for the paths
		// srv.tlsServer.AddHandler(RouteUpdateTD, srv.ServeUpdateTD)
		// srv.tlsServer.AddHandler(RouteGetTD, srv.ServeGetTD)
		// srv.tlsServer.AddHandler(RouteDeleteTD, srv.ServeDeleteTD)
		// srv.tlsServer.AddHandler(RoutePatchTD, srv.ServePatchTD)
		// srv.tlsServer.AddHandler(RouteListTD, srv.ServeListTD)
		// srv.tlsServer.AddHandler(RouteQueryTD, srv.ServeQueryTD)

		if srv.discoveryType != "" {
			directoryPath := RouteGetTD
			srv.discoServer, _ = ServeDiscovery(srv.instanceID, srv.discoveryType, srv.address, srv.port, directoryPath)
		}
		// Make sure the server is listening before continuing
		// Not pretty but it handles it
		time.Sleep(time.Second)
	}
	return nil
}

//Stop the IdProv server
func (srv *DirectoryServer) Stop() {
	if srv.running {
		srv.running = false
		if srv.discoServer != nil {
			srv.discoServer.Shutdown()
			srv.discoServer = nil
		}
	}
}

// Create a new instance of the IoT Device Provisioning Server
//  storePath is the location of the directory storage file
//  instanceID is the unique ID for this service used in discovery and communication
//  address the server listening address. Must use the same address as the services
//  port server listening port
//  caCertFolder location of CA Cert and server certificates and keys
//  discoveryType is the DNS-SD type. Use "" to disable discover, or idprov.IdprovServiceDiscoveryType for default

func NewDirectoryServer(
	instanceID string,
	storePath string,
	address string,
	port uint,
	caCertFolder string,
	discoveryType string) *DirectoryServer {

	if instanceID == "" || port == 0 {
		logrus.Panic("NewDirectoryServer: Invalid arguments for instanceID or port")
		panic("Exit due to invalid args")
	}
	srv := DirectoryServer{
		address:       address,
		caCertFolder:  caCertFolder,
		discoveryType: discoveryType,
		instanceID:    instanceID,
		port:          port,
		store:         dirfilestore.NewDirFileStore(storePath),
	}
	return &srv
}
