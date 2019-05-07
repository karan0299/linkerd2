package destination

import (
	"context"
	"fmt"

	pb "github.com/linkerd/linkerd2-proxy-api/go/destination"
	"github.com/linkerd/linkerd2/controller/api/destination/watcher"
	discoveryPb "github.com/linkerd/linkerd2/controller/gen/controller/discovery"
	"github.com/linkerd/linkerd2/controller/k8s"
	"github.com/linkerd/linkerd2/pkg/prometheus"
	logging "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type (
	server struct {
		trafficSplit *watcher.TrafficSplitWatcher
		profiles     *watcher.ProfileWatcher

		enableH2Upgrade     bool
		controllerNS        string
		identityTrustDomain string

		log      *logging.Entry
		shutdown <-chan struct{}
	}
)

// NewServer returns a new instance of the destination server.
//
// The destination server serves service discovery and other information to the
// proxy.  This implementation supports the "k8s" destination scheme and expects
// destination paths to be of the form:
// <service>.<namespace>.svc.cluster.local:<port>
//
// If the port is omitted, 80 is used as a default.  If the namespace is
// omitted, "default" is used as a default.append
//
// Addresses for the given destination are fetched from the Kubernetes Endpoints
// API.
func NewServer(
	addr string,
	controllerNS string,
	identityTrustDomain string,
	enableH2Upgrade bool,
	k8sAPI *k8s.API,
	shutdown <-chan struct{},
) *grpc.Server {
	log := logging.WithFields(logging.Fields{
		"addr":      addr,
		"component": "server",
	})
	endpoints := watcher.NewEndpointsWatcher(k8sAPI, log)
	trafficSplit := watcher.NewTrafficSplitWatcher(endpoints, k8sAPI, log)
	profiles := watcher.NewProfileWatcher(k8sAPI, log)

	srv := server{
		trafficSplit,
		profiles,
		enableH2Upgrade,
		controllerNS,
		identityTrustDomain,
		log,
		shutdown,
	}

	s := prometheus.NewGrpcServer()
	// linkerd2-proxy-api/destination.Destination (proxy-facing)
	pb.RegisterDestinationServer(s, &srv)
	// controller/discovery.Discovery (controller-facing)
	discoveryPb.RegisterDiscoveryServer(s, &srv)
	return s
}

func (s *server) Get(dest *pb.GetDestination, stream pb.Destination_GetServer) error {
	client, _ := peer.FromContext(stream.Context())
	log := s.log
	if client != nil {
		log = s.log.WithField("remote", client.Addr)
	}
	log.Debugf("Get %s", dest.GetPath())

	translator, err := newEndpointTranslator(
		s.controllerNS,
		s.identityTrustDomain,
		s.enableH2Upgrade,
		dest.GetPath(),
		stream,
		log,
	)
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.trafficSplit.Subscribe(dest.GetPath(), translator)
	if err != nil {
		log.Errorf("Failed to subscribe to %s: %s", dest.GetPath(), err)
		return err
	}
	defer s.trafficSplit.Unsubscribe(dest.GetPath(), translator)

	select {
	case <-s.shutdown:
	case <-stream.Context().Done():
		log.Debugf("Get %s cancelled", dest.GetPath())
	}

	return nil
}

func (s *server) GetProfile(dest *pb.GetDestination, stream pb.Destination_GetProfileServer) error {
	client, _ := peer.FromContext(stream.Context())
	log := s.log.WithField("remote", client.Addr)
	log.Debugf("GetProfile(%+v)", dest)

	translator := newProfileTranslator(stream, log)

	primary, secondary := newFallbackProfileListener(translator)

	if dest.GetContextToken() != "" {
		s.profiles.Subscribe(dest.GetPath(), dest.GetContextToken(), primary)
		defer s.profiles.Unsubscribe(dest.GetPath(), dest.GetContextToken(), primary)
	}

	s.profiles.Subscribe(dest.GetPath(), "", secondary)
	defer s.profiles.Unsubscribe(dest.GetPath(), "", secondary)

	select {
	case <-s.shutdown:
	case <-stream.Context().Done():
		log.Debugf("GetProfile(%+v) cancelled", dest)
	}

	return nil
}

func (s *server) Endpoints(ctx context.Context, params *discoveryPb.EndpointsParams) (*discoveryPb.EndpointsResponse, error) {
	s.log.Debugf("serving endpoints request")

	// servicePorts := e.getState()

	// rsp := discoveryPb.EndpointsResponse{
	// 	ServicePorts: make(map[string]*discoveryPb.ServicePort),
	// }

	// for serviceID, portMap := range servicePorts {
	// 	discoverySP := discoveryPb.ServicePort{
	// 		PortEndpoints: make(map[uint32]*discoveryPb.PodAddresses),
	// 	}
	// 	for port, sp := range portMap {
	// 		podAddrs := discoveryPb.PodAddresses{
	// 			PodAddresses: []*discoveryPb.PodAddress{},
	// 		}

	// 		for _, ua := range sp.addresses {
	// 			ownerKind, ownerName := s.k8sAPI.GetOwnerKindAndName(ua.pod)
	// 			pod := util.K8sPodToPublicPod(*ua.pod, ownerKind, ownerName)

	// 			podAddrs.PodAddresses = append(
	// 				podAddrs.PodAddresses,
	// 				&discoveryPb.PodAddress{
	// 					Addr: addr.NetToPublic(ua.address),
	// 					Pod:  &pod,
	// 				},
	// 			)
	// 		}

	// 		discoverySP.PortEndpoints[port] = &podAddrs
	// 	}

	// 	s.log.Debugf("ServicePorts[%s]: %+v", serviceID, discoverySP)
	// 	rsp.ServicePorts[serviceID.String()] = &discoverySP
	// }

	return nil, fmt.Errorf("Not implemented;")
}
