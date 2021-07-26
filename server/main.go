package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/suborbital/grav/grav"
	ghttp "github.com/suborbital/grav/transport/http"
	"github.com/suborbital/reactr/rt"
	"github.com/suborbital/vektor/vk"
	"github.com/suborbital/vektor/vlog"
)

const gravGetRemoteMessageReq = "grav.getremote.request"
const gravGetRemoteMessageRes = "grav.getremote.response"
const msgTypePing = "grav.pingremote"

type getResponse struct {
	Host   string `json:"host"`
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

func main() {
	logger := vlog.Default(vlog.Level(vlog.LogLevelDebug))
	gravhttp := ghttp.New()

	port := "8080"

	peerlist, found := os.LookupEnv("PEER_FILE")
	if !found {
		logger.ErrorString("PEER_FILE environment variable not set")
		return
	}

	peers := MustParsePeerList(peerlist)

	server := vk.New(vk.UseAppName("gravup"), vk.UseHTTPPort(8080))
	discovery := NewStaticDiscovery(peers, server, gravhttp)
	reactr := rt.New()

	g := grav.New(
		grav.UseLogger(logger),
		grav.UseEndpoint(port, ""),
		grav.UseTransport(gravhttp),
		grav.UseDiscovery(discovery),
	)

	// advertise self
	server.GET(StaticDiscoveryPath, func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		return vk.Respond(200, UUID{UUID: g.NodeUUID}), nil
	})

	reactr.HandleMsg(g.Connect(), msgTypePing, &ping{})
	pod := g.Connect()

	// TODO: collect responses for specific client
	server.GET("/get", func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		host := r.URL.Query().Get("host")
		logger.Info("trying", host)

		pod.Send(grav.NewMsg(gravGetRemoteMessageReq, []byte(host)))
		pod.Send(grav.NewMsg(msgTypePing, []byte(host)))

		fmt.Println("==== local result:", doRemoteRequest(host))

		pod.WaitOn(func(m grav.Message) error {
			if m.Type() != gravGetRemoteMessageRes {
				return grav.ErrMsgNotWanted
			}
			fmt.Println("==== remote result:", string(m.Data()))

			return nil
		})

		// do nothing for now
		return vk.Respond(204, nil), nil
	})

	server.POST("/meta/message", gravhttp.HandlerFunc())

	pod.OnType(gravGetRemoteMessageReq, func(m grav.Message) error {
		result, err := json.Marshal(doRemoteRequest(string(m.Data())))
		if err != nil {
			return err
		}
		pod.Send(grav.NewMsg(gravGetRemoteMessageRes, result))
		return nil
	})

	server.Start()
}

func doRemoteRequest(host string) getResponse {
	getresp, err := http.Get("http://" + host)
	if err != nil {
		return getResponse{
			Host:   host,
			Status: -1,
			Error:  err.Error(),
		}
	}

	defer getresp.Body.Close()

	return getResponse{
		Host:   host,
		Status: getresp.StatusCode,
		Error:  "",
	}
}
