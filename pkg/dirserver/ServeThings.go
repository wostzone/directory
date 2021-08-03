package dirserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wostzone/thingdir/pkg/dirclient"
)

// ServeThings lists or queries available TDs
// If a queryparam is provided then run a query, otherwise get the list
func (srv *DirectoryServer) ServeThings(response http.ResponseWriter, request *http.Request) {
	var offset = 0
	var tdList []interface{}
	limit, err := srv.tlsServer.GetQueryInt(request, dirclient.ParamLimit, dirclient.DefaultLimit)
	if limit > dirclient.MaxLimit {
		limit = dirclient.MaxLimit
	}
	if err == nil {
		offset, err = srv.tlsServer.GetQueryInt(request, dirclient.ParamOffset, 0)
	}
	if err != nil || offset < 0 {
		srv.tlsServer.WriteBadRequest(response, "ServeThings: offset or limit incorrect")
		return
	}
	jsonPath := srv.tlsServer.GetQueryString(request, dirclient.ParamQuery, "")
	if jsonPath == "" {
		tdList = srv.store.List(offset, limit)
	} else {
		tdList, err = srv.store.Query(jsonPath, offset, limit)
		if err != nil {
			msg := fmt.Sprintf("ServeThings: query error: %s", err)
			srv.tlsServer.WriteBadRequest(response, msg)
			return
		}
	}

	msg, err := json.Marshal(tdList)
	if err != nil {
		msg := fmt.Sprintf("ServeThings: Marshal error %s", err)
		srv.tlsServer.WriteInternalError(response, msg)
		return
	}
	response.Write(msg)
}
