package nginx

const (
	// L5dHeaderHTTP ..
	L5dHeaderHTTP = "proxy_set_header l5d-dst-override $service_name.$namespace.svc.cluster.local:$service_port;"
	// L5dHeaderGRPC ..
	L5dHeaderGRPC = "grpc_set_header l5d-dst-override $service_name.$namespace.svc.cluster.local:$service_port;"
)
