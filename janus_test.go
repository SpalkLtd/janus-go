package janus_test

import (
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/SpalkLtd/janus-go"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestGatewayReconnect(t *testing.T) {
	// Reconnect(sessionId uint64, handleId uint64, plugin string) (IHandle, ISession, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = successResponse(12345, transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	Claim req: SuccessMsg
	handle, session, err := gw.Reconnect(12345, 54321, "video-room")
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, uint64(12345), session.GetId())
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())
}

func TestGatewayReconnectNoHandleId(t *testing.T) {
	// Reconnect(sessionId uint64, handleId uint64, plugin string) (IHandle, ISession, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			resp = successResponse(12345, transaction)
		} else if transaction == 2 {
			resp = successResponse(1111, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	Claim req: SuccessMsg
	handle, session, err := gw.Reconnect(12345, 0, "video-room")
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, uint64(12345), session.GetId())
	require.NotNil(t, handle)
	require.Equal(t, uint64(1111), handle.GetId())
}

func TestGatewayReconnectError(t *testing.T) {
	// Reconnect(sessionId uint64, handleId uint64, plugin string) (IHandle, ISession, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = errorResponse("failed to reconnect", transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	Claim req: ERROR
	_, _, err = gw.Reconnect(12345, 54321, "video-room")
	require.Error(t, err)
	require.Equal(t, "failed to reconnect", err.Error())
}

func TestGatewayInfo(t *testing.T) {
	// 	Info() (*InfoMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("GET", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = infoResponse(transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	info req: InfoMsg
	infoMsg, err := gw.Info()
	require.NoError(t, err)
	require.NotNil(t, infoMsg)
	require.Equal(t, "janus-test-server", infoMsg.Name)
}

func TestGatewayInfoError(t *testing.T) {
	// 	Info() (*InfoMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("GET", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = errorResponse("error getting info", transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	info req: ERROR
	_, err = gw.Info()
	require.Error(t, err)
	require.Equal(t, "error getting info", err.Error())
}

func TestGatewayInfoUnexpectedResponse(t *testing.T) {
	// 	Info() (*InfoMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("GET", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = successResponse(12345, transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	info req: unknown response
	_, err = gw.Info()
	require.Error(t, err)
	require.Equal(t, "Unexpected response received to 'info' request", err.Error())
}
func TestGatewayCreate(t *testing.T) {
	// 	Create() (ISession, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = successResponse(12345, transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())
}

func TestGatewayCreateError(t *testing.T) {
	// 	Create() (ISession, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = errorResponse("error creating room", transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ERROR
	_, err = gw.Create()
	require.Error(t, err)
	require.Equal(t, "error creating room", err.Error())
}

func TestGatewayClose(t *testing.T) {
	// 	Close() error
	uri := "http://127.0.0.1:12345"
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	does nothing for http Close
	err = gw.Close()
	require.NoError(t, err)
}

func TestSessionGetId(t *testing.T) {
	// GetId() uint64
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = successResponse(12345, transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId()) //test this
}

func TestSessionEvents(t *testing.T) {
	// Events() <-chan interface{}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = successResponse(12345, transaction)
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)

	go func() {
		select {
		case msg := <-sess.Events():
			_, ok := msg.(*janus.SuccessMsg)
			require.True(t, ok)
			require.Equal(t, 2, transaction)
		case <-time.After(1 * time.Second):
			log.Println("Error getting the event message")
			t.FailNow()
		}
	}()

	sess.Attach("video-room") //cause a message to come through
}

func TestSessionAttach(t *testing.T) {
	// Attach(plugin string) (IHandle, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())
}

func TestSessionAttachError(t *testing.T) {
	// Attach(plugin string) (IHandle, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = errorResponse("error attaching", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	_, err = sess.Attach("video-room")
	require.Error(t, err)
	require.Equal(t, "error attaching", err.Error())
}

func TestSessionDestroy(t *testing.T) {
	// Destroy() (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			//session destroy ack response
			resp = ackResponse(transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)

	//don't need to check the ack as it's just an empty struct
	_, err = sess.Destroy()
	require.NoError(t, err)
}

func TestSessionDestroyError(t *testing.T) {
	// Destroy() (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			//session destroy ack response
			resp = errorResponse("error destroying", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)

	//don't need to check the ack as it's just an empty struct
	_, err = sess.Destroy()
	require.Error(t, err)
	require.Equal(t, "error destroying", err.Error())
}

func TestSessionLongPollStartAndStop(t *testing.T) {
	// StopPoll() LongPollForEvents()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	stopPollCh := make(chan bool)
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		resp = successResponse(12345, transaction)
		transaction++ //should make it 2
		return resp, nil
	})
	httpmock.RegisterResponder("GET", uri+"/12345", func(req *http.Request) (*http.Response, error) {
		resp := ackResponse(transaction)
		transaction++ //should make it 3
		stopPollCh <- true
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	go sess.LongPollForEvents()

	stopPollWg := sync.WaitGroup{}
	stopPollWg.Add(1)
	go func() {
		select {
		case <-stopPollCh:
			sess.StopPoll()
			stopPollWg.Done()
		case <-time.After(1 * time.Second):
			log.Println("Error getting the event message")
			stopPollWg.Done()
		}
	}()

	stopPollWg.Wait()
	require.Equal(t, 3, transaction)
}

func TestSessionKeepAlive(t *testing.T) {
	// KeepAlive() (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// session keep alive response
			resp = ackResponse(transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	keepalive req: ackMsg
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	//dont need to check the ack message as it's empty
	_, err = sess.KeepAlive()
	require.NoError(t, err)
}

func TestSessionKeepAliveError(t *testing.T) {
	// KeepAlive() (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// session keep alive response
			resp = errorResponse("error keeping alive", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	keepalive req: ackMsg
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	//dont need to check the ack message as it's empty
	_, err = sess.KeepAlive()
	require.Error(t, err)
	require.Equal(t, "error keeping alive", err.Error())
}

func TestSessionKeepAliveUnexpectedResponse(t *testing.T) {
	// KeepAlive() (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// session keep alive response
			resp = successResponse(1234, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	keepalive req: successMsg (should be ackMsg)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	//dont need to check the ack message as it's empty
	_, err = sess.KeepAlive()
	require.Error(t, err)
	require.Equal(t, "Unexpected response received to 'keepalive' request", err.Error())
}

func TestHandleGetId(t *testing.T) {
	// GetId() uint64
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())
}

func TestHandleRequest(t *testing.T) {
	// Request(body interface{}) (*SuccessMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (start resp)
			resp = successResponse(54321, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	create req: successMsg
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	msg, err := handle.Request(struct {
		Request string
		Room    int
	}{"create", 1234})
	require.NoError(t, err)
	require.NotNil(t, msg)
}

func TestHandleRequestError(t *testing.T) {
	// Request(body interface{}) (*SuccessMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (start resp)
			resp = errorResponse("error creating", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	create req: ERROR
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	_, err = handle.Request(struct {
		Request string
		Room    int
	}{"create", 1234})
	require.Error(t, err)
	require.Equal(t, "error creating", err.Error())
}

func TestHandleRequestUnexpectedResponse(t *testing.T) {
	// Request(body interface{}) (*SuccessMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (start resp)
			resp = ackResponse(transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	create req: UNEXPECTED (should be success)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	_, err = handle.Request(struct {
		Request string
		Room    int
	}{"create", 1234})
	require.Error(t, err)
	require.Equal(t, "Unexpected response received to 'message' request", err.Error())
}

func TestHandleMessage(t *testing.T) {
	// Message(body, jsep interface{}) (*EventMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (publish resp)
			resp = eventResponse(54321, transaction, "sdp-answer")
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	publish req: eventMsg
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	msg, err := handle.Message(struct {
		Request string
		Room    int
	}{"publish", 1234}, struct {
		Type string
		Sdp  string
	}{"offer", "sdp-offer"})
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "sdp-answer", msg.Jsep["sdp"])
}

func TestHandleMessageError(t *testing.T) {
	// Message(body, jsep interface{}) (*EventMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (publish resp)
			resp = errorResponse("error publishing", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	publish req: ERROR
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	_, err = handle.Message(struct {
		Request string
		Room    int
	}{"publish", 1234}, struct {
		Type string
		Sdp  string
	}{"offer", "sdp-offer"})
	require.Error(t, err)
	require.Equal(t, "error publishing", err.Error())
}

func TestHandleMessageUnexpectedResponse(t *testing.T) {
	// Message(body, jsep interface{}) (*EventMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (publish resp)
			resp = successResponse(54321, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	publish req: Unexpected (should be event response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	_, err = handle.Message(struct {
		Request string
		Room    int
	}{"publish", 1234}, struct {
		Type string
		Sdp  string
	}{"offer", "sdp-offer"})
	require.Error(t, err)
	require.Equal(t, "Unexpected response received to 'message' request", err.Error())
}

func TestHandleTrickle(t *testing.T) {
	// Trickle(candidate interface{}) (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = ackResponse(transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	trickle req: Unexpected (should be event response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.Trickle(janus.Candidate{})
	require.NoError(t, err)
}

func TestHandleTrickleError(t *testing.T) {
	// Trickle(candidate interface{}) (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = errorResponse("error trickling", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	publish req: Unexpected (should be event response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.Trickle(janus.Candidate{})
	require.Error(t, err)
	require.Equal(t, "error trickling", err.Error())
}

func TestHandleTrickleUnexpectedResponse(t *testing.T) {
	// Trickle(candidate interface{}) (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = successResponse(1234, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	publish req: Unexpected (should be event response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.Trickle(janus.Candidate{})
	require.Error(t, err)
	require.Equal(t, "Unexpected response received to 'trickle' request", err.Error())
}

func TestHandleTrickleMany(t *testing.T) {
	// TrickleMany(candidates interface{}) (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = ackResponse(transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	trickle req: Unexpected (should be event response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.TrickleMany([]janus.Candidate{})
	require.NoError(t, err)
}

func TestHandleTrickleManyError(t *testing.T) {
	// TrickleMany(candidates interface{}) (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = errorResponse("error trickling many", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	trickle many req: Unexpected (should be event response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.TrickleMany([]janus.Candidate{})
	require.Error(t, err)
	require.Equal(t, "error trickling many", err.Error())
}

func TestHandleTrickleManyUnexpectedResponse(t *testing.T) {
	// TrickleMany(candidates interface{}) (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = successResponse(1234, transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	trickle many req: Unexpected (should be event response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.TrickleMany([]janus.Candidate{})
	require.Error(t, err)
	require.Equal(t, "Unexpected response received to 'trickle' request", err.Error())
}

func TestHandleDetach(t *testing.T) {
	// Detach() (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = ackResponse(transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	detach req: ack response
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.Detach()
	require.NoError(t, err)
}
func TestHandleDetachError(t *testing.T) {
	// Detach() (*AckMsg, error)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	uri := "http://127.0.0.1:12345"
	transaction := 1
	httpmock.RegisterResponder("POST", uri, func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		if transaction == 1 {
			//session create response
			resp = successResponse(12345, transaction)
		}
		if transaction == 2 {
			// handle attach response
			resp = successResponse(54321, transaction)
		}
		if transaction == 3 {
			//handle request response (trickle resp)
			resp = errorResponse("error detaching", transaction)
		}
		transaction++
		return resp, nil
	})
	gw, err := janus.ConnectHttp(uri)
	require.NoError(t, err)

	//Order of http responses
	//	create req: ISession
	//	attach req: IHandle
	// 	detach req: ERROR (should be ack response)
	sess, err := gw.Create()
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.Equal(t, uint64(12345), sess.GetId())

	handle, err := sess.Attach("video-room")
	require.NoError(t, err)
	require.NotNil(t, handle)
	require.Equal(t, uint64(54321), handle.GetId())

	//dont need to check the ack
	_, err = handle.Detach()
	require.Error(t, err)
	require.Equal(t, "error detaching", err.Error())
}
