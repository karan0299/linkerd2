package public

import (
	"context"
	"fmt"
	"net/http"

	healthcheckPb "github.com/linkerd/linkerd2/controller/gen/common/healthcheck"
	discoveryPb "github.com/linkerd/linkerd2/controller/gen/controller/discovery"
	tapPb "github.com/linkerd/linkerd2/controller/gen/controller/tap"
	pb "github.com/linkerd/linkerd2/controller/gen/public"
	"github.com/linkerd/linkerd2/controller/k8s"
	"github.com/linkerd/linkerd2/pkg/prometheus"
	"github.com/linkerd/linkerd2/pkg/protohttp"
	promApi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

var (
	statSummaryPath   = fullURLPathFor("StatSummary")
	topRoutesPath     = fullURLPathFor("TopRoutes")
	versionPath       = fullURLPathFor("Version")
	listPodsPath      = fullURLPathFor("ListPods")
	listServicesPath  = fullURLPathFor("ListServices")
	tapByResourcePath = fullURLPathFor("TapByResource")
	selfCheckPath     = fullURLPathFor("SelfCheck")
	endpointsPath     = fullURLPathFor("Endpoints")
	configPath        = fullURLPathFor("Config")
)

type handler struct {
	grpcServer APIServer
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"req.Method": req.Method, "req.URL": req.URL, "req.Form": req.Form,
	}).Debugf("Serving %s %s", req.Method, req.URL.Path)
	// Validate request method
	if req.Method != http.MethodPost {
		protohttp.WriteErrorToHTTPResponse(w, fmt.Errorf("POST required"))
		return
	}

	// Serve request
	switch req.URL.Path {
	case statSummaryPath:
		h.handleStatSummary(w, req)
	case topRoutesPath:
		h.handleTopRoutes(w, req)
	case versionPath:
		h.handleVersion(w, req)
	case listPodsPath:
		h.handleListPods(w, req)
	case listServicesPath:
		h.handleListServices(w, req)
	case tapByResourcePath:
		h.handleTapByResource(w, req)
	case selfCheckPath:
		h.handleSelfCheck(w, req)
	case endpointsPath:
		h.handleEndpoints(w, req)
	case configPath:
		h.handleConfig(w, req)
	default:
		http.NotFound(w, req)
	}

}

func (h *handler) handleStatSummary(w http.ResponseWriter, req *http.Request) {
	var protoRequest pb.StatSummaryRequest

	err := protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	rsp, err := h.grpcServer.StatSummary(req.Context(), &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

func (h *handler) handleTopRoutes(w http.ResponseWriter, req *http.Request) {
	var protoRequest pb.TopRoutesRequest

	err := protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	rsp, err := h.grpcServer.TopRoutes(req.Context(), &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

func (h *handler) handleVersion(w http.ResponseWriter, req *http.Request) {
	var protoRequest pb.Empty
	err := protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	rsp, err := h.grpcServer.Version(req.Context(), &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

func (h *handler) handleSelfCheck(w http.ResponseWriter, req *http.Request) {
	var protoRequest healthcheckPb.SelfCheckRequest
	err := protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	rsp, err := h.grpcServer.SelfCheck(req.Context(), &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

func (h *handler) handleListPods(w http.ResponseWriter, req *http.Request) {
	var protoRequest pb.ListPodsRequest
	err := protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	rsp, err := h.grpcServer.ListPods(req.Context(), &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

func (h *handler) handleListServices(w http.ResponseWriter, req *http.Request) {
	var protoRequest pb.ListServicesRequest

	err := protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	rsp, err := h.grpcServer.ListServices(req.Context(), &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

func (h *handler) handleTapByResource(w http.ResponseWriter, req *http.Request) {
	flushableWriter, err := protohttp.NewStreamingWriter(w)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	var protoRequest pb.TapByResourceRequest
	err = protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	server := tapServer{w: flushableWriter, req: req}
	err = h.grpcServer.TapByResource(&protoRequest, server)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

func (h *handler) handleConfig(w http.ResponseWriter, req *http.Request) {
	var protoRequest pb.Empty
	err := protohttp.HTTPRequestToProto(req, &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	rsp, err := h.grpcServer.Config(req.Context(), &protoRequest)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}

	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

type tapServer struct {
	w   protohttp.FlushableResponseWriter
	req *http.Request
}

func (s tapServer) Send(msg *pb.TapEvent) error {
	err := protohttp.WriteProtoToHTTPResponse(s.w, msg)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(s.w, err)
		return err
	}

	s.w.Flush()
	return nil
}

// satisfy the pb.Api_TapServer interface
func (s tapServer) SetHeader(metadata.MD) error  { return nil }
func (s tapServer) SendHeader(metadata.MD) error { return nil }
func (s tapServer) SetTrailer(metadata.MD)       {}
func (s tapServer) Context() context.Context     { return s.req.Context() }
func (s tapServer) SendMsg(interface{}) error    { return nil }
func (s tapServer) RecvMsg(interface{}) error    { return nil }

func fullURLPathFor(method string) string {
	return apiRoot + apiPrefix + method
}

func (h *handler) handleEndpoints(w http.ResponseWriter, req *http.Request) {
	rsp, err := h.grpcServer.Endpoints(req.Context(), &discoveryPb.EndpointsParams{})
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
	err = protohttp.WriteProtoToHTTPResponse(w, rsp)
	if err != nil {
		protohttp.WriteErrorToHTTPResponse(w, err)
		return
	}
}

// NewServer creates a Public API HTTP server.
func NewServer(
	addr string,
	prometheusClient promApi.Client,
	tapClient tapPb.TapClient,
	discoveryClient discoveryPb.DiscoveryClient,
	k8sAPI *k8s.API,
	controllerNamespace string,
	ignoredNamespaces []string,
) *http.Server {
	baseHandler := &handler{
		grpcServer: newGrpcServer(
			promv1.NewAPI(prometheusClient),
			tapClient,
			discoveryClient,
			k8sAPI,
			controllerNamespace,
			ignoredNamespaces,
		),
	}

	instrumentedHandler := prometheus.WithTelemetry(baseHandler)

	return &http.Server{
		Addr:    addr,
		Handler: instrumentedHandler,
	}
}
