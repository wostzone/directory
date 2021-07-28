package dirserver

import (
	"net"
	"os"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/idprov-go/pkg/idprov"
	"github.com/wostzone/wostlib-go/pkg/discovery"
)

// // Get a list of active network interfaces excluding the loopback interface
// //  address to only return the interface that serves the given IP address
// func GetInterfaces(address string) ([]net.Interface, error) {
// 	result := make([]net.Interface, 0)
// 	ip := net.ParseIP(address)

// 	ifaces, err := net.Interfaces()
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, iface := range ifaces {
// 		addrs, err := iface.Addrs()
// 		// ignore interfaces without address
// 		if err == nil {
// 			for _, a := range addrs {
// 				switch v := a.(type) {
// 				case *net.IPAddr:
// 					result = append(result, iface)
// 					logrus.Infof("GetInterfaces: Found: Interface%s", v.String())

// 				case *net.IPNet:
// 					ifNet := a.(*net.IPNet)
// 					hasIP := ifNet.Contains(ip)

// 					// ignore loopback interface
// 					if hasIP && !a.(*net.IPNet).IP.IsLoopback() {
// 						result = append(result, iface)
// 						logrus.Infof("GetInterfaces: Found network %v : %s [%v/%v]\n", iface.Name, v, v.IP, v.Mask)
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return result, nil
// }

// ServeDiscovery serves the idprov DNS-SD record for IDProv server
//  instanceID is the unique ID of the service instance, usually the pluginID
//  discoveryType is the DNS-SD idprov type. Use "" for default (_idprov._tcp)
//  address idprov listening IP address
//  port idprov listing port
//  directoryPath is the path to the idprov directory endpoint, example: "/idprov/directory"
// Returns the discovery service instance. Use Shutdown() when done.
func ServeDiscovery(instanceID string, discoveryType string,
	address string, port uint, directoryPath string) (*zeroconf.Server, error) {
	var ips []string

	logrus.Infof("ServeDiscovery serviceID=%s, service: %s:%d%s",
		instanceID, address, port, directoryPath)
	if discoveryType == "" {
		discoveryType = idprov.IdprovServiceDiscoveryType
	}

	domain := "local." // fixme, from URL
	pathText := "directory=" + directoryPath
	hostname, _ := os.Hostname()

	// if the given address isn't a valid IP address. try to resolve it instead
	ips = []string{address}
	if net.ParseIP(address) == nil {
		// was a hostname provided instead IP?
		hostname = address
		actualIP, err := net.LookupIP(address)
		if err != nil {
			// can't continue without a valid address
			logrus.Errorf("ServeDiscovery: Provided address '%s' is not an IP and cannot be resolved: %s", address, err)
			return nil, err
		}
		ips = []string{actualIP[0].String()}
	}

	ifaces, err := discovery.GetInterfaces(ips[0])
	if err != nil {
		logrus.Warningf("ServeDiscovery: Address %s does not appear on any interface. Continuing anyways", ips[0])
	}
	text := []string{pathText}

	server, err := zeroconf.RegisterProxy(
		instanceID, discoveryType, domain, int(port), hostname, ips, text, ifaces)
	if err != nil {
		logrus.Errorf("ServeDiscovery: Failed to start the zeroconf server: %s", err)
	}
	return server, err
}
