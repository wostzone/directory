# Thing Directory Service 

Golang implementation of a Thing Directory Service plugin for the WoST Hub.

## Objective

Provide a WoST Hub plugin for registering Thing Descriptions and let consumers query these.


## Status

The status of this plugin is Under Development. 
This initial version is intended for local Hubs to track available things.


## Audience

This project is aimed at IoT developers that value the security and interoperability that WoST brings.
WoST Things are more secure than traditional IoT devices as they do not run a server, but instead connect to a Hub to publish their information and receive actions. 


## Dependencies


## Summary

This directory service provides the means to store and query registered Things.

The [WoT Directory Specification](https://w3c.github.io/wot-discovery/#exploration-directory-api) describes the requirements and is used to guide this implementation. The intent is to be compliant where possible. Note that at the time of implementation this specification is still a draft and subject to change. While the specification covers both service discovery and a directory service, this service focuses exclusively on the directory aspect. For discovery see the '[idprov-go](https://github.com/wostzone/idprov-go)' plugin.

In WoST, the registration of Things is the responsibility of the Hub Directory Protocol Binding. This protocol binding listens on the message bus for updates to Thing Descriptions and registers these into the Thing Directory. Things themselves only have to publish their updates on the message bus.
 
Multiple hubs can use the same Thing Directory to register store Thing TD information.

> TD data flow: Thing -> Hub message bus -> WostDir protocol binding -> WostDir Thing Directory Store

The initial version of the Thing Directory Store uses a simple file store with in-memory cache for fast JSON-Path based queries. Updates are periodicially flushed to disk. 

Future versions might support different choices of backends and support for more complex query facilities.



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

This plugin runs as part of the WoST hub. It has no additional requirements other than a working hub.

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

Like all WoST plugins, it can be run from the commandline or via the WoST hub as long as it can find the hub.yaml configuration, needed to connect to the message bus.
The plugin loads the hub configuration and the plugin configuration to determine how to connect to the hub and configure plugin specific settings.

Once started consumers can use the service API to query discovered Things and subscribe to events.

### Authentication

Authentication and access control follows the standard Hub authentication and authorization method:

Two authentication methods can be used:
 1. A valid certificate 
 2. A valid username/password login as for the mosquitto message bus

 For authorization, ACL restricts access to Things that are in the same group as the user.
 
 Full administration rights is available to certificate holders with the admin OU.


