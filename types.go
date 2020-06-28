// Msg Types
//
// All messages received from the gateway are first decoded to the BaseMsg
// type. The BaseMsg type extracts the following JSON from the message:
//		{
//			"janus": <Type>,
//			"transaction": <ID>,
//			"session_id": <Session>,
//			"sender": <Handle>
//		}
// The Type field is inspected to determine which concrete type
// to decode the message to, while the other fields (ID/Session/Handle) are
// inspected to determine where the message should be delivered. Messages
// with an ID field defined are considered responses to previous requests, and
// will be passed directly to requester. Messages without an ID field are
// considered unsolicited events from the gateway and are expected to have
// both Session and Handle fields defined. They will be passed to the Events
// channel of the related Handle and can be read from there.

package janus

var msgtypes = map[string]func() interface{}{
	"error":       func() interface{} { return &ErrorMsg{} },
	"success":     func() interface{} { return &SuccessMsg{} },
	"detached":    func() interface{} { return &DetachedMsg{} },
	"server_info": func() interface{} { return &InfoMsg{} },
	"ack":         func() interface{} { return &AckMsg{} },
	"event":       func() interface{} { return &EventMsg{} },
	"webrtcup":    func() interface{} { return &WebRTCUpMsg{} },
	"media":       func() interface{} { return &MediaMsg{} },
	"hangup":      func() interface{} { return &HangupMsg{} },
	"slowlink":    func() interface{} { return &SlowLinkMsg{} },
	"timeout":     func() interface{} { return &TimeoutMsg{} },
	"trickle":     func() interface{} { return &TrickleMsg{} },
}

type BaseMsg struct {
	Type    string `json:"janus"`
	ID      string `json:"transaction"`
	Session uint64 `json:"session_id"`
	Handle  uint64 `json:"sender"`
}

type ErrorMsg struct {
	Err ErrorData `json:"error"`
}

type ErrorData struct {
	Code   int
	Reason string
}

func (err *ErrorMsg) Error() string {
	return err.Err.Reason
}

type SuccessMsg struct {
	Data       SuccessData
	PluginData PluginData
	Session    uint64 `json:"session_id"`
	Handle     uint64 `json:"sender"`
}

type SuccessData struct {
	ID uint64
}

type DetachedMsg struct{}

type InfoMsg struct {
	Name          string
	Version       int
	VersionString string `json:"version_string"`
	Author        string
	DataChannels  bool   `json:"data_channels"`
	IPv6          bool   `json:"ipv6"`
	LocalIP       string `json:"local-ip"`
	IceTCP        bool   `json:"ice-tcp"`
	Transports    map[string]PluginInfo
	Plugins       map[string]PluginInfo
}

type PluginInfo struct {
	Name          string
	Author        string
	Description   string
	Version       int
	VersionString string `json:"version_string"`
}

type AckMsg struct{}

type EventMsg struct {
	Plugindata PluginData
	Jsep       map[string]interface{}
	Session    uint64 `json:"session_id"`
	Handle     uint64 `json:"sender"`
}

type PluginData struct {
	Plugin string
	Data   map[string]interface{}
}

type WebRTCUpMsg struct {
	Session uint64 `json:"session_id"`
	Handle  uint64 `json:"sender"`
}

type TimeoutMsg struct {
	Session uint64 `json:"session_id"`
}

type SlowLinkMsg struct {
	Uplink bool
	Lost   int64
}

type MediaMsg struct {
	Type      string
	Receiving bool
}

type HangupMsg struct {
	Reason  string
	Session uint64 `json:"session_id"`
	Handle  uint64 `json:"sender"`
}

type TrickleMsg struct {
	Session   uint64    `json:"session_id"`
	Handle    uint64    `json:"sender"`
	Candidate Candidate `json:"candidate"`
}

type Candidate struct {
	SdpMid        string `json:"sdpMid"`
	SdpMLineIndex int    `json:"sdpMLineIndex"`
	Candidate     string `json:"candidate"`
	Completed     bool   `json:"completed"`
}

//MessageRequest ...
type MessageRequest struct {
	Request string `json:"request"`
}

//CreateRoomMessage creates the janus video room
type CreateRoomMessage struct {
	MessageRequest
	Room int `json:"room"`
}

//PublishVideoRoomMessage for janus
type PublishVideoRoomMessage struct {
	MessageRequest
	Audio   bool   `json:"audio"`
	Video   bool   `json:"video"`
	Data    bool   `json:"data"`
	Display string `json:"display"`
}

//JoinRoomMessage for janus
type JoinRoomMessage struct {
	MessageRequest
	Display string `json:"display"`
	PType   string `json:"ptype"`
	Room    int    `json:"room"`
}

//AttachToPublishersRoomMessage for janus
type AttachToPublishersRoomMessage struct {
	JoinRoomMessage
	Feed int `json:"feed"` //unique ID of the publisher to subscribe to
}

//ListParticipantsInRoomMessage ...
type ListParticipantsInRoomMessage struct {
	MessageRequest
	Room int `json:"room"`
}
