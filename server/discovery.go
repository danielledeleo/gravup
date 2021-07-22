package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/suborbital/grav/grav"
	ghttp "github.com/suborbital/grav/transport/http"
	"github.com/suborbital/vektor/vk"
	"github.com/suborbital/vektor/vlog"
)

const StaticDiscoveryPath = "/discovery/uuid"

type UUID struct{ UUID string }

type Peer struct {
	Address string
	Port    string
	falloff int64
}

type StaticDiscovery struct {
	opts          *grav.DiscoveryOpts
	log           *vlog.Logger
	discoveryFunc grav.DiscoveryFunc
	peers         []Peer
}

func NewStaticDiscovery(peers []Peer, vektor *vk.Server, gravhttp *ghttp.Transport) *StaticDiscovery {
	return &StaticDiscovery{
		peers: peers}
}

func (d *StaticDiscovery) Start(opts *grav.DiscoveryOpts, discoveryFunc grav.DiscoveryFunc) error {

	d.opts = opts
	d.log = opts.Logger
	d.discoveryFunc = discoveryFunc

	d.log.Info("[discovery-static] starting static discovery", opts.TransportPort, opts.TransportURI)

	// get each peer's UUID
	for _, p := range d.peers {
		go func(peer Peer) {
			peer.falloff = 1
			for {

				uri := fmt.Sprintf("%s:%s%s", peer.Address, peer.Port, StaticDiscoveryPath)

				resp, err := http.Get(uri)
				if err != nil {
					d.log.Error(err)
					<-time.After(time.Second * time.Duration(peer.falloff))
					peer.falloff *= 2
					continue
				}

				uuid := UUID{}
				err = json.NewDecoder(resp.Body).Decode(&uuid)

				if err != nil {
					d.log.Error(err)
				}
				d.log.Info("found peer", uuid.UUID)

				d.discoveryFunc(fmt.Sprintf("%s:%s%s", peer.Address, peer.Port, "/meta/message"), uuid.UUID)
				return
			}
		}(p)
	}

	return nil
}
