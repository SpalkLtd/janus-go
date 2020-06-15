package janus_test

import (
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
)

func ackResponse(transaction int) *http.Response {
	return httpmock.NewStringResponse(200, fmt.Sprintf(`{
		"janus" : "ack",
		"transaction" : "%d"
	}`, transaction))
}

func listParticipantsResponse(id, transaction int, displayName string) *http.Response {
	return httpmock.NewStringResponse(200, fmt.Sprintf(`{
		"janus" : "success",
		"transaction" : "%d",
		"data" : {
			"id": %d
			},
		"plugindata": {
			"plugin": "video-room",
			"data": {
				"participants" : [             
					{    
						"id" : 1111,
						"display" : "%s",
						"publisher" : true,
						"talking" : true
					}
				]
				}
		}
	}`, transaction, id, displayName))
}

func successResponse(id, transaction int) *http.Response {
	return httpmock.NewStringResponse(200, fmt.Sprintf(`{
		"janus" : "success",
		"transaction" : "%d",
		"data" : {
				"id" : %d
		}
	}`, transaction, id))
}

func eventResponse(id, transaction int, sdp string) *http.Response {
	return httpmock.NewStringResponse(200, fmt.Sprintf(`{
		"janus" : "event",
		"transaction" : "%d",
		"data" : {
				"id" : %d
		},
		"jsep": {
			"sdp": "%s"
		}
	}`, transaction, id, sdp))

}

func errorResponse(err string, transaction int) *http.Response {
	return httpmock.NewStringResponse(200, fmt.Sprintf(`{
		"janus" : "error",
		"transaction" : "%d",
		"error" : {
			"code": 0,
			"reason": "%s"
		}
	}`, transaction, err))
}

func infoResponse(transaction int) *http.Response {
	return httpmock.NewStringResponse(200, fmt.Sprintf(`{
		"janus" : "server_info",
		"transaction" : "%d",
		"name" : "janus-test-server",
		"version": 1234,
		"version_string": "1234",
		"author": "test_spalk",
		"data_channels": false,
		"ipv6": false,
		"local-ip": "127.0.0.1",
		"ice-tcp": false,
		"transports": {},
		"plugins": {}
	}`, transaction))
}
