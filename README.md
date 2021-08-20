# Thing Directory Service 

Golang implementation of a Thing Directory Service client and server library.

## Objective

Provide a service for registering Thing Descriptions, compatible with the WoT specification for directory services.


## Status

The status of this library is Alpha. 
This initial MVP version is intended for the Wost Hub directory protocol binding, but it is designed to be usable outside of WoST.

Still to do:
* authentication hook
* authorization hook


## Audience

This project is aimed at IoT developers that value the security and interoperability that WoST brings.
WoST Things are more secure than traditional IoT devices as they do not run a server, but instead connect to a Hub to publish their information and receive actions. 


## Dependencies


## Summary

This directory service provides the means to store and query registered Things.

The [WoT Directory Specification](https://w3c.github.io/wot-discovery/#exploration-directory-api) describes the requirements and is used to guide this implementation. The intent is to be compliant where possible. Note that at the time of implementation this specification is still a draft and subject to change. While the specification covers both service discovery and a directory service, this service focuses exclusively on the directory aspect. For discovery see the '[idprov-go](https://github.com/wostzone/idprov-go)' provisioning plugin.

In WoST, the registration of Things is the responsibility of the Directory Service. Things themselves only have to publish their updates on the message bus without consideration for who uses the information. This separation of concerns simplifies centralized access control using the hubauth service. It also simplifies the Thing device as it only has to connect to the message bus.


This package consists of 4 parts:
1. Directory client for golang clients. Clients for other languages will be made available as well, or users can implement their own using the protocol described below.

2. Directory store for storing and querying Thing Description documents. The current implementation uses a file based store with an in-memory cache. Additional storage backends can be added in the future.

3. Directory server to serve directory requests. This implements the server side of the directory protocol as described below. Authentication and authorization is handled using the hubauth service. See hubauth for details on the groups and roles that govern access.

4. Directory protocol binding to the WoST message bus. This service subscribes to the Thing message bus and updates the Thing directory with changes. Things do NOT update the directory themselves, they only need to publish their updates on the message bus.

The initial version of the Thing Directory Store uses a file store with in-memory cache for fast JSON-Path based queries. Updates are periodicially flushed to disk. 


## API

The directory service supports a registration API as outlined in the WoT directory specification. 

### Register a Thing TD

```http
HTTP PUT https://server:port/things/thingID
{
  ...TD...
}
201 (Created)
```
Other responses:
 * 201 (Update)  - the TD already exists and was replaced
 * 400 (Bad Request) - invalid serialization or TD
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization, or anonymous TDs 

```
A note on anonymous TDs
Anonymous TDs are not allowed in WoST. In order for things to provision and receive a certificate, they must have a thing ID.
```

### Get a Thing TD

```http
HTTP GET https://server:port/things/thingID
200 (OK)
{
  ... TD ...
}
```

Other responses:
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization, or anonymous TDs 
 * 404 (Not Found) - no such thing ID

### Update a Thing TD

To replace an existing TD:

```http
HTTP PUT https://server:port/things/thingID
Content-Type: application/td+json 
{
  ...TD...
}
204 (No Content)
```
Other responses:
 * 201 (Created) - Thing didn't exist and was created
 * 400 (Bad Request) - invalid serialization or TD
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization, or anonymous TDs 


To partially update an existing TD:

```http
HTTP PATCH https://server:port/things/thingID
Content-Type: application/merge-patch+json 
{
  ...Partial TD...
}
204 (No Content)
```
Other responses:
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization
 * 404 (Not Found) - TD with the given id not found


### Delete a Thing TD

```http
HTTP DELETE https://server:port/things/thingID
Content-Type: application/td+json 
{
  ...TD...
}
204 (No Content)
```

### Listing of Thing TDs

Example limit nr of results to 10

```http
HTTP GET https://server:port/things?offset=10&limit=10
200 (OK)
Content-Type: application/ld+json
Link: </things?offset=10>; rel="next"
[{TD},...]
```

The optional next link in the response is used to paginate additional results.

Other responses:
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization

### Search For Things With JSONPATH


JSONPATH queries are supported as follows:
> $.td[*].id                   -> list of IDs of things
> $.td[?(@.type=='thetype')]   -> TDs of type 'thetype'
> $.td[0,1]                    -> First two TDs


Example search

```http
HTTP GET https://server:port/things?queryparams=""
200 (OK)
Content-Type: application/json
[{TD},...]
```

Other responses:
 * 400 (Bad Request) - invalid serialization or TD
 * 401 (Unauthorized) - insufficient authentication
 * 403 (Forbidden) - insufficient authorization


Where queryparams identify property fields in the TD.

## Security

This services is a WoST Hub plugin and uses the Hub authentication and authorization facilities.

In addition, the following denial of service mitigation is provided:
1. Rate limiting. Limit the number of requests from the same client.
2. Request duration. Requests that take too long are aborted.
3. Monitoring. Track normal traffic load and alert on sudden traffic changes.

The parameters governing the mitigation can be defined in the service configuration.


## Build and Installation

### System Requirements

This is a stand-alone library intended to be used by WoST. It uses the wostlib-go library but is has otherwise no other dependencies on WoST.

### Manual Installation

See the hub README on plugin installation.


### Build From Source

Build with:
```
make all
```

When successful, the plugin can be found in dist/bin. An example configuration file is provided in config/directory.yaml. 

"make install" copies these to the local Hub directory at ~/bin/wost/{bin,config}


## Usage

This is a library intended to be used with an application that handles authentication and authorization. See also the thingdir-pb protocol binding that is included in the WoST Hub core.

Currently a file based backend is supported. Additional backends can be added in the future.


Starting and stopping the service:

```golang 
  import "github.com/wostzone/thingdir/pkg/dirserver"
  import "github.com/wostzone/wostlib-go/pkg/hubnet"


  instanceID := "mydirserver"
  serviceDiscoveryName := "thingdir"
  storeFolder := "./config"      // use your own folder to store the directory files
  serverAddress := hubnet.GetOutboundIP("").String()
  serverPort := "9999"
  serverCertPath := "certs/serverCert.pem"
  serverKeyPath := "certs/serverKey.pem"
  caCertPath := "certs/caCert.pem"
  authenticationHook := nil // your login/password authentication function
  authorizationHook := nil  // your authorization function

	directoryServer = dirserver.NewDirectoryServer(
		instanceID, 
    storeFolder, 
    serverAddress, serverPort,
		serviceDiscoveryName, 
    serverCertPath, serverKeyPath, 
    caCertPath )
	directoryServer.Start()

  // do stuff 

  // stop when done
  directoryServer.Stop()
```
