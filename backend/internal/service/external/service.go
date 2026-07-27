package external

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/iscsi"
	"cephtower/backend/internal/integration/ceph/monitoring"
	"cephtower/backend/internal/integration/ceph/nvmeof"
	"cephtower/backend/internal/integration/ceph/s3"
	"cephtower/backend/internal/security"
	endpointservice "cephtower/backend/internal/service/endpoint"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Service struct {
	endpoints     *endpointservice.Service
	encryptionKey string
	transport     http.RoundTripper
}
type httpCredential struct {
	Token             string `json:"token"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	CACertificate     string `json:"ca_certificate"`
	ClientCertificate string `json:"client_certificate"`
	ClientKey         string `json:"client_key"`
	AccessKey         string `json:"access_key"`
	SecretKey         string `json:"secret_key"`
	SessionToken      string `json:"session_token"`
	Region            string `json:"region"`
}

func Supports(action string) bool {
	switch action {
	case "silence.create", "silence.delete",
		"rgw_bucket.create", "rgw_bucket.update", "rgw_bucket.delete", "rgw_bucket_policy.update",
		"iscsi_target.create", "iscsi_target.update", "iscsi_target.delete",
		"nvmeof_subsystem.create", "nvmeof_subsystem.update", "nvmeof_subsystem.delete",
		"nvmeof_namespace.create", "nvmeof_namespace.update", "nvmeof_namespace.delete",
		"nvmeof_listener.create", "nvmeof_listener.delete",
		"nvmeof_host.create", "nvmeof_host.delete":
		return true
	default:
		return false
	}
}

func New(endpoints *endpointservice.Service, operations *operationservice.Service, encryptionKeys ...string) *Service {
	key := ""
	if len(encryptionKeys) > 0 {
		key = encryptionKeys[0]
	}
	s := &Service{endpoints: endpoints, encryptionKey: key}
	if operations != nil {
		for _, action := range []string{"silence.create", "silence.delete", "rgw_bucket.create", "rgw_bucket.update", "rgw_bucket.delete", "rgw_bucket_policy.update", "iscsi_target.create", "iscsi_target.update", "iscsi_target.delete", "nvmeof_subsystem.create", "nvmeof_subsystem.update", "nvmeof_subsystem.delete", "nvmeof_namespace.create", "nvmeof_namespace.update", "nvmeof_namespace.delete", "nvmeof_listener.create", "nvmeof_listener.delete", "nvmeof_host.create", "nvmeof_host.delete"} {
			_ = operations.Register(action, s.apply)
		}
	}
	return s
}

// Read performs protocol-native reads for resources that are not reconciled
// from Ceph CLI state. The returned value is already shaped for API data.
func (s *Service) Read(ctx context.Context, clusterID uint64, kind, key string, query url.Values) (any, error) {
	switch kind {
	case "metric":
		return s.readMetric(ctx, clusterID, key, query)
	case "alert", "alert_rule", "silence", "grafana":
		return s.readMonitoring(ctx, clusterID, kind)
	case "iscsi_gateway", "iscsi_target":
		return s.readISCSI(ctx, clusterID, kind, key)
	case "nvmeof_gateway", "nvmeof_subsystem", "nvmeof_namespace", "nvmeof_listener", "nvmeof_host", "nvmeof_connection":
		return s.readNVMeOF(ctx, clusterID, kind, key, query)
	case "rgw_bucket_policy":
		return s.readBucketPolicy(ctx, clusterID, key, query)
	default:
		return nil, failure("capability_unavailable", "external resource is unavailable", false)
	}
}

func (s *Service) CheckPlan(ctx context.Context, request operationservice.PlanRequest) ([]string, []string, error) {
	switch request.Action {
	case "rgw_bucket.delete":
		endpoint, credential, client, err := s.httpClient(ctx, request.ClusterID, "s3")
		if err != nil {
			return nil, nil, err
		}
		api, err := s3.New(endpoint.URL, s3.Credentials{AccessKey: credential.AccessKey, SecretKey: credential.SecretKey, SessionToken: credential.SessionToken, Region: credential.Region}, client)
		if err != nil {
			return nil, nil, err
		}
		bucket, err := decodeBucketID(pathAfter(request.ResourceKey, "bucket"))
		if err != nil {
			return nil, nil, err
		}
		if err := api.HeadBucket(ctx, bucket); err != nil {
			return nil, nil, err
		}
	case "iscsi_target.delete":
		endpoint, credential, client, err := s.httpClient(ctx, request.ClusterID, "iscsi")
		if err != nil {
			return nil, nil, err
		}
		api, err := iscsi.New(endpoint.URL, credential.Username, credential.Password, client)
		if err != nil {
			return nil, nil, err
		}
		if _, err := api.Target(ctx, pathAfter(request.ResourceKey, "target")); err != nil {
			return nil, nil, err
		}
	case "nvmeof_subsystem.delete", "nvmeof_namespace.delete", "nvmeof_listener.delete", "nvmeof_host.delete":
		kind := strings.TrimSuffix(request.Action, ".delete")
		key := pathAfter(request.ResourceKey, "subsystem")
		if kind == "nvmeof_namespace" {
			key += "\x00" + pathAfter(request.ResourceKey, "namespace")
		}
		if _, err := s.readNVMeOF(ctx, request.ClusterID, kind, key, nil); err != nil {
			return nil, nil, err
		}
	}
	return nil, nil, nil
}

func (s *Service) monitoringClient(ctx context.Context, clusterID uint64, endpointKind string) (*monitoring.Client, error) {
	endpoint, credential, client, err := s.httpClient(ctx, clusterID, endpointKind)
	if err != nil {
		return nil, err
	}
	api, err := monitoring.New(endpoint.URL, credential.Token, client)
	if err != nil {
		return nil, failure("invalid_endpoint", err.Error(), false)
	}
	return api, nil
}

func (s *Service) readMetric(ctx context.Context, clusterID uint64, key string, query url.Values) (any, error) {
	allowed := map[string]bool{"metric_id": true, "time": true, "start": true, "end": true, "step": true}
	if err := validateQuery(query, allowed); err != nil {
		return nil, err
	}
	api, err := s.monitoringClient(ctx, clusterID, "prometheus")
	if err != nil {
		return nil, err
	}
	metricID := strings.TrimSpace(query.Get("metric_id"))
	if metricID == "" {
		return nil, failure("invalid_request", "metric_id is required", false)
	}
	var result monitoring.PrometheusResult
	if strings.Contains(key, "range") {
		start, err := time.Parse(time.RFC3339Nano, query.Get("start"))
		if err != nil {
			return nil, failure("invalid_request", "start must be RFC3339", false)
		}
		end, err := time.Parse(time.RFC3339Nano, query.Get("end"))
		if err != nil {
			return nil, failure("invalid_request", "end must be RFC3339", false)
		}
		step, err := time.ParseDuration(query.Get("step"))
		if err != nil {
			return nil, failure("invalid_request", "step must be a Go duration such as 30s", false)
		}
		result, err = api.QueryRange(ctx, metricID, start, end, step)
	} else {
		var at *time.Time
		if raw := query.Get("time"); raw != "" {
			parsed, parseErr := time.Parse(time.RFC3339Nano, raw)
			if parseErr != nil {
				return nil, failure("invalid_request", "time must be RFC3339", false)
			}
			at = &parsed
		}
		result, err = api.Query(ctx, metricID, at)
	}
	if err != nil {
		return nil, failure("prometheus_failed", err.Error(), true)
	}
	return map[string]any{"result_type": result.Data.ResultType, "series": result.Data.Result, "meta": observedMeta()}, nil
}

func (s *Service) readMonitoring(ctx context.Context, clusterID uint64, kind string) (any, error) {
	endpointKind := "alertmanager"
	if kind == "grafana" {
		endpointKind = "grafana"
	} else if kind == "alert_rule" {
		endpointKind = "prometheus"
	}
	api, err := s.monitoringClient(ctx, clusterID, endpointKind)
	if err != nil {
		return nil, err
	}
	var items any
	switch kind {
	case "alert":
		items, err = api.Alerts(ctx)
	case "alert_rule":
		items, err = api.Rules(ctx)
	case "silence":
		items, err = api.Silences(ctx)
	case "grafana":
		items, err = api.Dashboards(ctx)
	}
	if err != nil {
		return nil, failure(endpointKind+"_failed", err.Error(), true)
	}
	return listResult(items), nil
}

func (s *Service) readISCSI(ctx context.Context, clusterID uint64, kind, key string) (any, error) {
	endpoint, credential, client, err := s.httpClient(ctx, clusterID, "iscsi")
	if err != nil {
		return nil, err
	}
	api, err := iscsi.New(endpoint.URL, credential.Username, credential.Password, client)
	if err != nil {
		return nil, failure("invalid_endpoint", err.Error(), false)
	}
	if kind == "iscsi_gateway" {
		result, err := api.Health(ctx)
		if err != nil {
			return nil, failure("iscsi_failed", err.Error(), true)
		}
		return result, nil
	}
	if key != "" {
		result, err := api.Target(ctx, key)
		if err != nil {
			return nil, failure("iscsi_failed", err.Error(), true)
		}
		return result, nil
	}
	items, err := api.Targets(ctx)
	if err != nil {
		return nil, failure("iscsi_failed", err.Error(), true)
	}
	return listResult(items), nil
}

func (s *Service) readNVMeOF(ctx context.Context, clusterID uint64, kind, key string, query url.Values) (any, error) {
	if err := validateQuery(query, map[string]bool{}); err != nil {
		return nil, err
	}
	endpoint, err := s.endpoints.Endpoint(ctx, clusterID, "nvmeof")
	if err != nil {
		return nil, failure("endpoint_unavailable", "NVMe-oF endpoint is not configured", false)
	}
	var credential httpCredential
	if err := s.decryptOptionalCredential(ctx, clusterID, "nvmeof", &credential); err != nil {
		return nil, err
	}
	if err := s.applyEndpointCA(ctx, clusterID, endpoint, &credential); err != nil {
		return nil, err
	}
	parsed, err := url.Parse(endpoint.URL)
	if err != nil {
		return nil, failure("invalid_endpoint", "NVMe-oF endpoint URL is invalid", false)
	}
	api, err := nvmeof.Dial(ctx, nvmeof.DialConfig{Address: parsed.Host, ServerName: parsed.Hostname(), BearerToken: credential.Token, CACertificate: []byte(credential.CACertificate), ClientCertificate: []byte(credential.ClientCertificate), ClientKey: []byte(credential.ClientKey)})
	if err != nil {
		return nil, failure("nvmeof_unavailable", err.Error(), true)
	}
	defer api.Close()
	nqn := key
	childID := ""
	if parts := strings.SplitN(key, "\x00", 2); len(parts) == 2 {
		nqn, childID = parts[0], parts[1]
	}
	var response proto.Message
	switch kind {
	case "nvmeof_gateway":
		response, err = api.Gateway().GetGatewayInfo(ctx, &nvmeof.GetGatewayInfoReq{})
	case "nvmeof_subsystem":
		var filter *string
		if nqn != "" {
			filter = &nqn
		}
		response, err = api.ListSubsystems(ctx, filter)
	case "nvmeof_namespace":
		request := &nvmeof.ListNamespacesReq{Subsystem: nqn}
		if childID != "" {
			parsedID, parseErr := strconv.ParseUint(childID, 10, 32)
			if parseErr != nil {
				return nil, failure("invalid_request", "nsid must be an unsigned integer", false)
			}
			value := uint32(parsedID)
			request.Nsid = &value
		}
		response, err = api.ListNamespaces(ctx, request)
	case "nvmeof_listener":
		response, err = api.ListListeners(ctx, &nvmeof.ListListenersReq{Subsystem: nqn})
	case "nvmeof_host":
		response, err = api.ListHosts(ctx, &nvmeof.ListHostsReq{Subsystem: nqn})
	case "nvmeof_connection":
		response, err = api.ListConnections(ctx, &nvmeof.ListConnectionsReq{Subsystem: nqn})
	}
	if err != nil {
		return nil, failure("nvmeof_failed", err.Error(), true)
	}
	encoded, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(response)
	if err != nil {
		return nil, failure("nvmeof_failed", "NVMe-oF response could not be normalized", false)
	}
	var data any
	if err := json.Unmarshal(encoded, &data); err != nil {
		return nil, failure("nvmeof_failed", "NVMe-oF response could not be normalized", false)
	}
	return map[string]any{"result": data, "meta": observedMeta()}, nil
}

func (s *Service) readBucketPolicy(ctx context.Context, clusterID uint64, key string, query url.Values) (any, error) {
	if err := validateQuery(query, map[string]bool{"kind": true}); err != nil {
		return nil, err
	}
	endpoint, credential, client, err := s.httpClient(ctx, clusterID, "s3")
	if err != nil {
		return nil, err
	}
	api, err := s3.New(endpoint.URL, s3.Credentials{AccessKey: credential.AccessKey, SecretKey: credential.SecretKey, SessionToken: credential.SessionToken, Region: credential.Region}, client)
	if err != nil {
		return nil, failure("invalid_credential", err.Error(), false)
	}
	bucket, err := decodeBucketID(key)
	if err != nil {
		return nil, failure("invalid_request", "bucket_id is invalid", false)
	}
	kind := query.Get("kind")
	if kind == "" {
		kind = "policy"
	}
	body, contentType, err := api.GetBucketConfiguration(ctx, bucket, kind)
	if err != nil {
		return nil, failure("s3_failed", err.Error(), true)
	}
	return map[string]any{"bucket_id": key, "kind": kind, "document": string(body), "content_type": contentType}, nil
}

func validateQuery(values url.Values, allowed map[string]bool) error {
	for name := range values {
		if !allowed[name] {
			return failure("invalid_request", "unsupported query parameter "+name, false)
		}
		if len(values[name]) != 1 {
			return failure("invalid_request", "query parameter "+name+" must appear once", false)
		}
	}
	return nil
}
func observedMeta() map[string]any {
	return map[string]any{"observed_at": time.Now().UTC(), "stale": false, "stale_reason": nil}
}
func listResult(items any) map[string]any {
	return map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": observedMeta()}
}
func (s *Service) apply(ctx context.Context, operation store.CephOperation) (cephdomain.OperationResult, error) {
	if operation.ClusterID == nil {
		return cephdomain.OperationResult{}, failure("endpoint_unavailable", "cluster endpoint is unavailable", false)
	}
	var stored any
	if err := json.Unmarshal([]byte(operation.RequestJSON), &stored); err != nil {
		return cephdomain.OperationResult{}, err
	}
	var err error
	if s.encryptionKey != "" {
		stored, err = security.UnprotectJSON(stored, s.encryptionKey)
		if err != nil {
			return cephdomain.OperationResult{}, failure("invalid_credential", "operation secrets could not be decrypted", false)
		}
	}
	parameters, ok := stored.(map[string]any)
	if !ok {
		return cephdomain.OperationResult{}, failure("invalid_request", "operation parameters must be an object", false)
	}
	switch {
	case strings.HasPrefix(operation.Action, "silence."):
		return s.alertmanager(ctx, *operation.ClusterID, operation, parameters)
	case strings.HasPrefix(operation.Action, "rgw_bucket"):
		return s.s3(ctx, *operation.ClusterID, operation, parameters)
	case strings.HasPrefix(operation.Action, "iscsi_"):
		return s.iscsi(ctx, *operation.ClusterID, operation, parameters)
	case strings.HasPrefix(operation.Action, "nvmeof_"):
		return s.nvmeof(ctx, *operation.ClusterID, operation, parameters)
	default:
		return cephdomain.OperationResult{}, failure("capability_unavailable", "external action is unavailable", false)
	}
}
func (s *Service) s3(ctx context.Context, clusterID uint64, operation store.CephOperation, parameters map[string]any) (cephdomain.OperationResult, error) {
	endpoint, credential, client, err := s.httpClient(ctx, clusterID, "s3")
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	api, err := s3.New(endpoint.URL, s3.Credentials{AccessKey: credential.AccessKey, SecretKey: credential.SecretKey, SessionToken: credential.SessionToken, Region: credential.Region}, client)
	if err != nil {
		return cephdomain.OperationResult{}, failure("invalid_credential", err.Error(), false)
	}
	bucket := ""
	if operation.Action == "rgw_bucket.create" {
		bucket, _ = parameters["name"].(string)
		if bucket == "" {
			bucket, _ = parameters["bucket"].(string)
		}
	} else {
		bucket, err = decodeBucketID(last(operation.ResourceKey))
		if err != nil {
			return cephdomain.OperationResult{}, failure("invalid_request", "bucket_id is invalid", false)
		}
	}
	switch operation.Action {
	case "rgw_bucket.create":
		err = api.CreateBucket(ctx, bucket)
	case "rgw_bucket.delete":
		err = api.DeleteBucket(ctx, bucket)
	case "rgw_bucket.update":
		versioning, _ := parameters["versioning"].(string)
		status := map[string]string{"enabled": "Enabled", "suspended": "Suspended"}[strings.ToLower(versioning)]
		if status == "" {
			return cephdomain.OperationResult{}, failure("invalid_request", "versioning must be enabled or suspended", false)
		}
		body := []byte("<VersioningConfiguration xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\"><Status>" + status + "</Status></VersioningConfiguration>")
		err = api.PutBucketConfiguration(ctx, bucket, "versioning", body)
	case "rgw_bucket_policy.update":
		kind, _ := parameters["kind"].(string)
		if kind == "" {
			kind = "policy"
		}
		document, exists := parameters["document"]
		if !exists {
			document = parameters[kind]
		}
		var body []byte
		if text, ok := document.(string); ok {
			body = []byte(text)
		} else {
			body, err = json.Marshal(document)
		}
		if err == nil && (len(body) == 0 || string(body) == "null") {
			err = fmt.Errorf("bucket configuration document is required")
		}
		if err == nil {
			err = api.PutBucketConfiguration(ctx, bucket, kind, body)
		}
	}
	if err != nil {
		return cephdomain.OperationResult{}, failure("s3_failed", err.Error(), true)
	}
	encodedID := base64.RawURLEncoding.EncodeToString([]byte("\x00" + bucket))
	return cephdomain.OperationResult{ResourceURL: fmt.Sprintf("/api/v1/cluster/%d/rgw/bucket/%s", clusterID, encodedID)}, nil
}
func (s *Service) alertmanager(ctx context.Context, clusterID uint64, operation store.CephOperation, parameters map[string]any) (cephdomain.OperationResult, error) {
	endpoint, credential, client, err := s.httpClient(ctx, clusterID, "alertmanager")
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	api, err := monitoring.New(endpoint.URL, credential.Token, client)
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	if operation.Action == "silence.delete" {
		if err := api.DeleteSilence(ctx, last(operation.ResourceKey)); err != nil {
			return cephdomain.OperationResult{}, failure("alertmanager_failed", err.Error(), true)
		}
		return cephdomain.OperationResult{}, nil
	}
	encoded, _ := json.Marshal(parameters)
	var request monitoring.Silence
	if err := json.Unmarshal(encoded, &request); err != nil {
		return cephdomain.OperationResult{}, failure("invalid_request", "invalid silence request", false)
	}
	id, err := api.CreateSilence(ctx, request)
	if err != nil {
		return cephdomain.OperationResult{}, failure("alertmanager_failed", err.Error(), true)
	}
	return cephdomain.OperationResult{ResourceURL: fmt.Sprintf("/api/v1/cluster/%d/silence/%s", clusterID, id), Details: map[string]any{"silence_id": id}}, nil
}
func (s *Service) iscsi(ctx context.Context, clusterID uint64, operation store.CephOperation, parameters map[string]any) (cephdomain.OperationResult, error) {
	endpoint, credential, client, err := s.httpClient(ctx, clusterID, "iscsi")
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	api, err := iscsi.New(endpoint.URL, credential.Username, credential.Password, client)
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	iqn := last(operation.ResourceKey)
	if operation.Action == "iscsi_target.delete" {
		if err := api.DeleteTarget(ctx, iqn); err != nil {
			return cephdomain.OperationResult{}, failure("iscsi_failed", err.Error(), true)
		}
		return cephdomain.OperationResult{}, nil
	}
	encoded, _ := json.Marshal(parameters)
	var target iscsi.Target
	if err := json.Unmarshal(encoded, &target); err != nil {
		return cephdomain.OperationResult{}, failure("invalid_request", "invalid iSCSI target", false)
	}
	if target.IQN == "" {
		target.IQN = iqn
	}
	if err := api.ApplyTarget(ctx, target); err != nil {
		return cephdomain.OperationResult{}, failure("iscsi_failed", err.Error(), true)
	}
	return cephdomain.OperationResult{ResourceURL: fmt.Sprintf("/api/v1/cluster/%d/iscsi/target/%s", clusterID, url.PathEscape(target.IQN))}, nil
}
func (s *Service) nvmeof(ctx context.Context, clusterID uint64, operation store.CephOperation, parameters map[string]any) (cephdomain.OperationResult, error) {
	endpoint, err := s.endpoints.Endpoint(ctx, clusterID, "nvmeof")
	if err != nil {
		return cephdomain.OperationResult{}, failure("endpoint_unavailable", "NVMe-oF endpoint is not configured", false)
	}
	var credential httpCredential
	if err := s.decryptOptionalCredential(ctx, clusterID, "nvmeof", &credential); err != nil {
		return cephdomain.OperationResult{}, err
	}
	if err := s.applyEndpointCA(ctx, clusterID, endpoint, &credential); err != nil {
		return cephdomain.OperationResult{}, err
	}
	parsed, err := url.Parse(endpoint.URL)
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	api, err := nvmeof.Dial(ctx, nvmeof.DialConfig{Address: parsed.Host, ServerName: parsed.Hostname(), BearerToken: credential.Token, CACertificate: []byte(credential.CACertificate), ClientCertificate: []byte(credential.ClientCertificate), ClientKey: []byte(credential.ClientKey)})
	if err != nil {
		return cephdomain.OperationResult{}, failure("nvmeof_unavailable", err.Error(), true)
	}
	defer api.Close()
	parameters["subsystem_nqn"] = pathAfter(operation.ResourceKey, "subsystem")
	var response proto.Message
	switch operation.Action {
	case "nvmeof_subsystem.create":
		var request nvmeof.CreateSubsystemReq
		if err := decodeProto(parameters, &request); err != nil {
			return cephdomain.OperationResult{}, err
		}
		response, err = api.CreateSubsystem(ctx, &request)
	case "nvmeof_subsystem.update":
		return cephdomain.OperationResult{}, failure("invalid_request", "NVMe-oF subsystem fields are immutable; manage child resources instead", false)
	case "nvmeof_subsystem.delete":
		request := nvmeof.DeleteSubsystemReq{SubsystemNqn: pathAfter(operation.ResourceKey, "subsystem")}
		response, err = api.DeleteSubsystem(ctx, &request)
	case "nvmeof_namespace.create":
		var request nvmeof.NamespaceAddReq
		if err := decodeProto(parameters, &request); err != nil {
			return cephdomain.OperationResult{}, err
		}
		response, err = api.AddNamespace(ctx, &request)
	case "nvmeof_namespace.update":
		var request nvmeof.NamespaceResizeReq
		if err := decodeProto(parameters, &request); err != nil {
			return cephdomain.OperationResult{}, err
		}
		if value := pathAfter(operation.ResourceKey, "namespace"); request.Nsid == 0 {
			parsed, _ := strconv.ParseUint(value, 10, 32)
			request.Nsid = uint32(parsed)
		}
		response, err = api.Gateway().NamespaceResize(ctx, &request)
	case "nvmeof_namespace.delete":
		var request nvmeof.NamespaceDeleteReq
		parameters["nsid"], _ = strconv.ParseUint(pathAfter(operation.ResourceKey, "namespace"), 10, 32)
		if err := decodeProto(parameters, &request); err != nil {
			return cephdomain.OperationResult{}, err
		}
		response, err = api.DeleteNamespace(ctx, &request)
	case "nvmeof_listener.create":
		var request nvmeof.CreateListenerReq
		if err := decodeProto(parameters, &request); err != nil {
			return cephdomain.OperationResult{}, err
		}
		response, err = api.CreateListener(ctx, &request)
	case "nvmeof_listener.delete":
		var request nvmeof.DeleteListenerReq
		if err := decodeProto(parameters, &request); err != nil {
			return cephdomain.OperationResult{}, err
		}
		response, err = api.DeleteListener(ctx, &request)
	case "nvmeof_host.create":
		var request nvmeof.AddHostReq
		if err := decodeProto(parameters, &request); err != nil {
			return cephdomain.OperationResult{}, err
		}
		response, err = api.AddHost(ctx, &request)
	case "nvmeof_host.delete":
		request := nvmeof.RemoveHostReq{SubsystemNqn: parameters["subsystem_nqn"].(string), HostNqn: pathAfter(operation.ResourceKey, "host")}
		response, err = api.RemoveHost(ctx, &request)
	}
	if err != nil {
		return cephdomain.OperationResult{}, failure("nvmeof_failed", err.Error(), true)
	}
	var details any
	encoded, _ := protojson.Marshal(response)
	_ = json.Unmarshal(encoded, &details)
	return cephdomain.OperationResult{Details: map[string]any{"gateway_response": details}}, nil
}
func (s *Service) httpClient(ctx context.Context, clusterID uint64, kind string) (store.CephClusterEndpoint, httpCredential, *http.Client, error) {
	endpoint, err := s.endpoints.Endpoint(ctx, clusterID, kind)
	if err != nil {
		return endpoint, httpCredential{}, nil, failure("endpoint_unavailable", kind+" endpoint is not configured", false)
	}
	var credential httpCredential
	if err := s.decryptOptionalCredential(ctx, clusterID, kind, &credential); err != nil {
		return endpoint, credential, nil, err
	}
	if err := s.applyEndpointCA(ctx, clusterID, endpoint, &credential); err != nil {
		return endpoint, credential, nil, err
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if credential.CACertificate != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(credential.CACertificate)) {
			return endpoint, credential, nil, failure("invalid_credential", "CA certificate is invalid", false)
		}
		tlsConfig.RootCAs = pool
	}
	if credential.ClientCertificate != "" {
		certificate, err := tls.X509KeyPair([]byte(credential.ClientCertificate), []byte(credential.ClientKey))
		if err != nil {
			return endpoint, credential, nil, failure("invalid_credential", "client certificate is invalid", false)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	timeout := 20 * time.Second
	if endpoint.ConfigJSON != nil {
		var config struct {
			TimeoutSeconds int `json:"timeout_seconds"`
		}
		if err := json.Unmarshal([]byte(*endpoint.ConfigJSON), &config); err == nil && config.TimeoutSeconds > 0 {
			timeout = time.Duration(config.TimeoutSeconds) * time.Second
		}
	}
	transport := s.transport
	if transport == nil {
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	}
	return endpoint, credential, &http.Client{Timeout: timeout, Transport: transport}, nil
}
func (s *Service) decryptOptionalCredential(ctx context.Context, clusterID uint64, kind string, target *httpCredential) error {
	err := s.endpoints.DecryptCredential(ctx, clusterID, kind, target)
	if errors.Is(err, store.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return failure("invalid_credential", kind+" credential could not be decrypted", false)
	}
	return nil
}
func (s *Service) applyEndpointCA(ctx context.Context, clusterID uint64, endpoint store.CephClusterEndpoint, target *httpCredential) error {
	if endpoint.CACredentialID == nil {
		return nil
	}
	var ca httpCredential
	if err := s.endpoints.DecryptCredentialByID(ctx, clusterID, *endpoint.CACredentialID, &ca); err != nil {
		return failure("invalid_credential", "endpoint CA credential could not be decrypted", false)
	}
	if strings.TrimSpace(ca.CACertificate) == "" {
		return failure("invalid_credential", "endpoint CA certificate is empty", false)
	}
	target.CACertificate = ca.CACertificate
	return nil
}
func decodeProto(value any, target proto.Message) error {
	encoded, _ := json.Marshal(value)
	if err := (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(encoded, target); err != nil {
		return failure("invalid_request", "invalid NVMe-oF request", false)
	}
	return nil
}
func pathAfter(path, segment string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := range parts {
		if parts[i] == segment && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
func last(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return parts[len(parts)-1]
}
func decodeBucketID(value string) (string, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) == 0 {
		return "", fmt.Errorf("invalid bucket ID")
	}
	parts := bytes.SplitN(decoded, []byte{0}, 2)
	bucket := parts[len(parts)-1]
	if len(bucket) == 0 || bytes.ContainsAny(bucket, "/\x00") {
		return "", fmt.Errorf("invalid bucket ID")
	}
	return string(bucket), nil
}
func failure(code, message string, retryable bool) error {
	return &cephdomain.OperationError{Code: code, Message: message, Retryable: retryable}
}
