package handler_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	endpointservice "cephtower/backend/internal/service/endpoint"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

func TestMutationQueuesInspectableOperation(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: contractKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "operations-api.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, contractKey, unusedProvider{})
	operations := operationservice.New(database, contractKey, nil, operationservice.Options{})
	endpoints := endpointservice.New(database, contractKey)
	if _, err := endpoints.CreateEndpoint(context.Background(), cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	h := handler.New(handler.Dependencies{Clusters: clusters, Endpoints: endpoints, Operations: operations, Database: database, AuthEnabled: func() bool { return false }})
	mux := http.NewServeMux()
	router.Register(mux, h)
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "rgw_admin", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	periodResponse := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/period/commit", fmt.Sprintf(`{"cluster_id":%d,"realm_id":"realm-explicit","expected_current_period":"old"}`, cluster.ID), "scoped-period")
	importBody := fmt.Sprintf(`{"cluster_id":%d,"name":"secondary","realm_token":"private-import-token","port":80,"placement":{},"confirm_import":true}`, cluster.ID)
	importResponse := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/realm/import", importBody, "realm-import")
	if importResponse.Code != http.StatusAccepted || strings.Contains(importResponse.Body.String(), "private-import-token") {
		t.Fatal("unsafe import queue response", importResponse.Code)
	}
	importOp, importErr := db.FindOperation(context.Background(), operationIDFromResponse(t, importResponse))
	if importErr != nil || importOp.Action != "rgw_realm.import" || importOp.Risk != "high" || importOp.MaxAttempts != 1 || strings.Contains(importOp.ParametersCiphertext, "private-import-token") {
		t.Fatal("unsafe import operation")
	}
	importPlain, importErr := security.Decrypt(importOp.ParametersCiphertext, contractKey)
	if importErr != nil || !strings.Contains(string(importPlain), "private-import-token") {
		t.Fatal("encrypted import token lost")
	}
	archiveResponse := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/realm/import", strings.TrimSuffix(importBody, "}")+`,"tier_type":"archive"}`, "archive-import")
	if archiveResponse.Code != http.StatusAccepted {
		t.Fatal("archive import rejected", archiveResponse.Code)
	}
	for _, extra := range []string{`,"tier_type":"unknown"}`, `,"unmanaged":true}`} {
		bad := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/realm/import", strings.TrimSuffix(importBody, "}")+extra, "bad-realm-import-"+extra)
		if bad.Code != http.StatusBadRequest {
			t.Fatal("unsupported import option accepted")
		}
	}
	if periodResponse.Code != http.StatusAccepted {
		t.Fatalf("period queue: %d %s", periodResponse.Code, periodResponse.Body.String())
	}
	periodOp, err := db.FindOperation(context.Background(), operationIDFromResponse(t, periodResponse))
	if err != nil || periodOp.Action != "rgw_period.commit" || periodOp.Risk != "high" || periodOp.ResourceKey != "rgw/period/commit" {
		t.Fatalf("period operation: %+v %v", periodOp, err)
	}
	unscoped := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/period/commit", fmt.Sprintf(`{"cluster_id":%d}`, cluster.ID), "unscoped-period")
	if unscoped.Code != http.StatusBadRequest {
		t.Fatalf("unscoped commit accepted: %d", unscoped.Code)
	}
	for _, mode := range []string{"attach", "detach"} {
		body := fmt.Sprintf(`{"cluster_id":%d,"account_id":"RGW12345678901234567","name":"role","owner_uid":"tenant$user","mode":%q,"policy_arn":"arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess","expected_role_id":"role-id","expected_role_arn":"arn:aws:iam::RGW12345678901234567:role/role","expected_policies":[]}`, cluster.ID, mode)
		response := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/role/managed/policy", body, "role-policy-"+mode)
		if response.Code != http.StatusAccepted {
			t.Fatalf("managed policy queue: %d %s", response.Code, response.Body.String())
		}
		op, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || op.Action != "rgw_role.managed_policy" || op.Risk != "high" || op.ResourceKey != "rgw/role/RGW12345678901234567/role" || op.LockKey != "role" {
			t.Fatalf("wrong managed policy operation: %+v %v", op, err)
		}
		bad := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/role/managed/policy", strings.Replace(body, `"expected_policies":[]`, `"expected_policies":null`, 1), "bad-role-policy-"+mode)
		if bad.Code != http.StatusBadRequest {
			t.Fatal("missing managed policy snapshot accepted")
		}
	}
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		id := base64.RawURLEncoding.EncodeToString([]byte(scope + ":events"))
		createBody, _ := json.Marshal(map[string]any{"cluster_id": cluster.ID, "topic_id": id, "topic_arn": "arn:aws:sns:zone:" + scope + ":events", "owner_uid": "user", "endpoint_secret": "https://u:create-secret@host/path", "opaque_data": "", "policy": "", "persistent": false, "time_to_live": "None", "max_retries": "None", "retry_sleep_duration": "None", "options": map[string]any{}})
		createResponse := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/topic", string(createBody), "topic-create-"+scope)
		if createResponse.Code != http.StatusAccepted || strings.Contains(createResponse.Body.String(), "create-secret") {
			t.Fatalf("unsafe creation queue %d", createResponse.Code)
		}
		createOp, err := db.FindOperation(context.Background(), operationIDFromResponse(t, createResponse))
		if err != nil || createOp.Action != "rgw_topic.create" || createOp.Risk != "high" || createOp.ResourceKey != "rgw/topic/"+id || strings.Contains(createOp.ParametersCiphertext, "create-secret") {
			t.Fatal("wrong creation operation")
		}
		body := fmt.Sprintf(`{"cluster_id":%d,"topic_id":%q,"expected_version":%q}`, cluster.ID, id, `{"tag":"t","ver":9007199254740993}`)
		response := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/topic", body, "topic-delete-"+scope)
		if response.Code != http.StatusAccepted {
			t.Fatalf("topic deletion queue: %d %s", response.Code, response.Body.String())
		}
		op, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || op.Action != "rgw_topic.delete" || op.Risk != "high" || op.ResourceKey != "rgw/topic/"+id || !strings.Contains(op.LockKey, id) {
			t.Fatalf("wrong topic operation %+v %v", op, err)
		}
		invalid := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/topic", strings.TrimSuffix(body, "}")+`,"purge":true}`, "topic-invalid-"+scope)
		if invalid.Code != http.StatusBadRequest {
			t.Fatalf("unknown deletion field accepted %d", invalid.Code)
		}
		policyBody := fmt.Sprintf(`{"cluster_id":%d,"topic_id":%q,"topic_arn":%q,"expected_policy":"{}","policy":""}`, cluster.ID, id, "arn:aws:sns:default:"+scope+":events")
		policyResponse := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/topic/policy", policyBody, "topic-policy-"+scope)
		if policyResponse.Code != http.StatusAccepted {
			t.Fatalf("policy queue: %d %s", policyResponse.Code, policyResponse.Body.String())
		}
		policyOp, err := db.FindOperation(context.Background(), operationIDFromResponse(t, policyResponse))
		if err != nil || policyOp.Action != "rgw_topic.policy" || policyOp.Risk != "high" || policyOp.ResourceKey != op.ResourceKey || policyOp.LockKey != op.LockKey {
			t.Fatalf("wrong policy operation %+v %v", policyOp, err)
		}
		policyInvalid := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/topic/policy", strings.TrimSuffix(policyBody, "}")+`,"endpoint":"other"}`, "topic-policy-invalid-"+scope)
		if policyInvalid.Code != http.StatusBadRequest {
			t.Fatalf("unknown policy field accepted: %d", policyInvalid.Code)
		}
		attributeBody := fmt.Sprintf(`{"cluster_id":%d,"topic_id":%q,"topic_arn":%q,"attribute":"persistent","expected_value":"true","value":"false"}`, cluster.ID, id, "arn:aws:sns:default:"+scope+":events")
		attributeResponse := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/topic/attribute", attributeBody, "topic-attribute-"+scope)
		if attributeResponse.Code != http.StatusAccepted {
			t.Fatalf("attribute queue: %d %s", attributeResponse.Code, attributeResponse.Body.String())
		}
		attributeOp, err := db.FindOperation(context.Background(), operationIDFromResponse(t, attributeResponse))
		if err != nil || attributeOp.Action != "rgw_topic.attribute" || attributeOp.Risk != "high" || attributeOp.ResourceKey != op.ResourceKey || attributeOp.LockKey != op.LockKey {
			t.Fatalf("wrong attribute operation %+v %v", attributeOp, err)
		}
		attributeInvalid := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/topic/attribute", strings.Replace(attributeBody, `"attribute":"persistent"`, `"attribute":"password"`, 1), "topic-attribute-invalid-"+scope)
		if attributeInvalid.Code != http.StatusBadRequest {
			t.Fatalf("unsupported attribute accepted: %d", attributeInvalid.Code)
		}
		optionBody := fmt.Sprintf(`{"cluster_id":%d,"topic_id":%q,"topic_arn":%q,"option":"verify-ssl","value":"false","expected_status":"unset","expected_value":""}`, cluster.ID, id, "arn:aws:sns:default:"+scope+":events")
		optionResponse := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/topic/option", optionBody, "topic-option-"+scope)
		if optionResponse.Code != http.StatusAccepted {
			t.Fatalf("option queue: %d", optionResponse.Code)
		}
		optionOp, err := db.FindOperation(context.Background(), operationIDFromResponse(t, optionResponse))
		if err != nil || optionOp.Action != "rgw_topic.option" || optionOp.Risk != "high" || optionOp.ResourceKey != op.ResourceKey || optionOp.LockKey != op.LockKey {
			t.Fatal("wrong option operation")
		}
		optionInvalid := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/topic/option", strings.Replace(optionBody, `"verify-ssl"`, `"password"`, 1), "topic-option-invalid-"+scope)
		if optionInvalid.Code != http.StatusBadRequest {
			t.Fatal("credential option allowed")
		}
		endpointBody := fmt.Sprintf(`{"cluster_id":%d,"topic_id":%q,"topic_arn":%q,"expected_endpoint":"https://old/path","expected_redacted":false,"expected_stored_secret":false,"endpoint_secret":"amqps://user:private-password@broker/vhost"}`, cluster.ID, id, "arn:aws:sns:default:"+scope+":events")
		endpointResponse := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/topic/endpoint", endpointBody, "topic-endpoint-"+scope)
		if endpointResponse.Code != http.StatusAccepted || strings.Contains(endpointResponse.Body.String(), "private-password") {
			t.Fatalf("unsafe endpoint queue: %d", endpointResponse.Code)
		}
		endpointOp, err := db.FindOperation(context.Background(), operationIDFromResponse(t, endpointResponse))
		if err != nil || endpointOp.Action != "rgw_topic.endpoint" || endpointOp.Risk != "high" || endpointOp.ResourceKey != op.ResourceKey || endpointOp.LockKey != op.LockKey || strings.Contains(endpointOp.ParametersCiphertext, "private-password") {
			t.Fatal("unsafe endpoint operation")
		}
		plain, err := security.Decrypt(endpointOp.ParametersCiphertext, contractKey)
		if err != nil || !strings.Contains(string(plain), "amqps://user:private-password@broker/vhost") {
			t.Fatal("endpoint credentials not preserved encrypted")
		}
	}
	for _, realm := range []string{"", "realm"} {
		response := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/zonegroup/sync/group", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"status":"enabled"}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{},"pipes":[]}`), "zonegroup-sync-"+realm)
		if response.Code != http.StatusAccepted {
			t.Fatalf("zonegroup sync queue: %d %s", response.Code, response.Body.String())
		}
		op, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || op.Action != "rgw_zonegroup.sync_group" || op.Risk != "high" || op.ResourceKey != "rgw/zonegroup/east" {
			t.Fatalf("wrong zonegroup operation: %+v %v", op, err)
		}
		created := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/zonegroup/sync/group", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"new","expected_policy":%q,"status":"allowed"}`, cluster.ID, realm, `{"groups":[]}`), "zonegroup-sync-create-"+realm)
		if created.Code != http.StatusAccepted {
			t.Fatalf("zonegroup create queue: %d %s", created.Code, created.Body.String())
		}
		creation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, created))
		if err != nil || creation.Action != "rgw_zonegroup.sync_group_create" || creation.Risk != "high" || creation.ResourceKey != op.ResourceKey || creation.LockKey != op.LockKey {
			t.Fatalf("wrong zonegroup creation: %+v %v", creation, err)
		}
		deleted := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/zonegroup/sync/group", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q}`, cluster.ID, realm, `{"id":"g","status":"forbidden","data_flow":{},"pipes":[]}`), "zonegroup-sync-delete-"+realm)
		flow := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/zonegroup/sync/flow", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"flow_type":"directional","source_zone":"a","dest_zone":"b"}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{},"pipes":[]}`), "zonegroup-flow-create-"+realm)
		if flow.Code != http.StatusAccepted {
			t.Fatalf("flow: %d %s", flow.Code, flow.Body.String())
		}
		flowOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, flow))
		pipe := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/zonegroup/sync/pipe", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"pipe_id":"p","source_zones":["*"],"dest_zones":["z"],"source_bucket":"*","dest_bucket":"photos","mode":"system"}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{},"pipes":[]}`), "zonegroup-pipe-create-"+realm)
		if pipe.Code != http.StatusAccepted {
			t.Fatalf("pipe: %d %s", pipe.Code, pipe.Body.String())
		}
		pipeOperation, pipeErr := db.FindOperation(context.Background(), operationIDFromResponse(t, pipe))
		preparation := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/zonegroup/replication/prepare", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"expected_zones":["z1","z2"],"expected_policy":"{\"groups\":[]}"}`, cluster.ID, realm), "zonegroup-replication-prepare-"+realm)
		if preparation.Code != http.StatusAccepted {
			t.Fatalf("preparation: %d %s", preparation.Code, preparation.Body.String())
		}
		preparedOperation, preparationErr := db.FindOperation(context.Background(), operationIDFromResponse(t, preparation))
		if preparationErr != nil || preparedOperation.Action != "rgw_zonegroup.replication_prepare" || preparedOperation.Risk != "high" || preparedOperation.ResourceKey != op.ResourceKey || preparedOperation.LockKey != op.LockKey {
			t.Fatalf("preparation: %+v %v", preparedOperation, preparationErr)
		}
		pipeZonesBody := fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"pipe_id":"p","source_zones":["*"],"dest_zones":["z"]}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{},"pipes":[{"id":"p"}]}`)
		pipeZones := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/zonegroup/sync/pipe/zones", pipeZonesBody, "zonegroup-pipe-zones-"+realm)
		if pipeZones.Code != http.StatusAccepted {
			t.Fatalf("pipe zones: %d %s", pipeZones.Code, pipeZones.Body.String())
		}
		pipeZoneOperation, pipeZoneErr := db.FindOperation(context.Background(), operationIDFromResponse(t, pipeZones))
		if pipeZoneErr != nil || pipeZoneOperation.Action != "rgw_zonegroup.sync_pipe_zones" || pipeZoneOperation.Risk != "high" || pipeZoneOperation.ResourceKey != op.ResourceKey || pipeZoneOperation.LockKey != op.LockKey {
			t.Fatalf("pipe zones: %+v %v", pipeZoneOperation, pipeZoneErr)
		}
		pipeUpdateBody := fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"pipe_id":"p","source_bucket":"*","dest_bucket":"photos","mode":"system"}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{},"pipes":[{"id":"p"}]}`)
		pipeUpdate := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/zonegroup/sync/pipe", pipeUpdateBody, "zonegroup-pipe-update-"+realm)
		if pipeUpdate.Code != http.StatusAccepted {
			t.Fatalf("pipe update: %d %s", pipeUpdate.Code, pipeUpdate.Body.String())
		}
		pipeUpdated, updatePipeErr := db.FindOperation(context.Background(), operationIDFromResponse(t, pipeUpdate))
		if updatePipeErr != nil || pipeUpdated.Action != "rgw_zonegroup.sync_pipe_update" || pipeUpdated.Risk != "high" || pipeUpdated.ResourceKey != op.ResourceKey || pipeUpdated.LockKey != op.LockKey {
			t.Fatalf("pipe update: %+v %v", pipeUpdated, updatePipeErr)
		}
		rejectedUpdate := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/zonegroup/sync/pipe", strings.TrimSuffix(pipeUpdateBody, "}")+`,"source_zones":["*"]}`, "zonegroup-pipe-update-zones-"+realm)
		if rejectedUpdate.Code != http.StatusBadRequest {
			t.Fatalf("zone fields accepted: %d", rejectedUpdate.Code)
		}
		pipeDeleteBody := fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"pipe_id":"p"}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{},"pipes":[{"id":"p"}]}`)
		pipeDelete := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/zonegroup/sync/pipe", pipeDeleteBody, "zonegroup-pipe-delete-"+realm)
		if pipeDelete.Code != http.StatusAccepted {
			t.Fatalf("pipe delete: %d %s", pipeDelete.Code, pipeDelete.Body.String())
		}
		pipeDeletion, deletePipeErr := db.FindOperation(context.Background(), operationIDFromResponse(t, pipeDelete))
		if deletePipeErr != nil || pipeDeletion.Action != "rgw_zonegroup.sync_pipe_delete" || pipeDeletion.Risk != "high" || pipeDeletion.ResourceKey != op.ResourceKey || pipeDeletion.LockKey != op.LockKey {
			t.Fatalf("pipe deletion: %+v %v", pipeDeletion, deletePipeErr)
		}
		withSelector := strings.TrimSuffix(pipeDeleteBody, "}") + `,"source_zones":["*"]}`
		rejectedPipe := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/zonegroup/sync/pipe", withSelector, "zonegroup-pipe-delete-selectors-"+realm)
		if rejectedPipe.Code != http.StatusBadRequest {
			t.Fatalf("selectors accepted: %d %s", rejectedPipe.Code, rejectedPipe.Body.String())
		}
		if pipeErr != nil || pipeOperation.Action != "rgw_zonegroup.sync_pipe_create" || pipeOperation.Risk != "high" || pipeOperation.ResourceKey != op.ResourceKey || pipeOperation.LockKey != op.LockKey {
			t.Fatalf("pipe: %+v %v", pipeOperation, pipeErr)
		}
		flowUpdate := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/zonegroup/sync/flow", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"flow_id":"f","zones":["z"]}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{"symmetrical":[{"id":"f","zones":["old"]}]},"pipes":[]}`), "zonegroup-flow-update-"+realm)
		if flowUpdate.Code != http.StatusAccepted {
			t.Fatalf("flow update: %d %s", flowUpdate.Code, flowUpdate.Body.String())
		}
		updatedFlow, updateErr := db.FindOperation(context.Background(), operationIDFromResponse(t, flowUpdate))
		if updateErr != nil || updatedFlow.Action != "rgw_zonegroup.sync_flow_update" || updatedFlow.Risk != "high" || updatedFlow.ResourceKey != op.ResourceKey || updatedFlow.LockKey != op.LockKey {
			t.Fatalf("flow update: %+v %v", updatedFlow, updateErr)
		}
		flowDelete := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/zonegroup/sync/flow", fmt.Sprintf(`{"cluster_id":%d,"name":"east","zonegroup_id":"zg","realm_id":%q,"group_id":"g","expected_group":%q,"flow_type":"symmetrical","flow_id":"f"}`, cluster.ID, realm, `{"id":"g","status":"allowed","data_flow":{"symmetrical":[{"id":"f","zones":["z"]}]},"pipes":[]}`), "zonegroup-flow-delete-"+realm)
		if flowDelete.Code != http.StatusAccepted {
			t.Fatalf("flow delete: %d %s", flowDelete.Code, flowDelete.Body.String())
		}
		deletionFlow, deleteErr := db.FindOperation(context.Background(), operationIDFromResponse(t, flowDelete))
		if deleteErr != nil || deletionFlow.Action != "rgw_zonegroup.sync_flow_delete" || deletionFlow.Risk != "high" || deletionFlow.ResourceKey != op.ResourceKey || deletionFlow.LockKey != op.LockKey {
			t.Fatalf("flow deletion: %+v %v", deletionFlow, deleteErr)
		}
		if err != nil || flowOperation.Action != "rgw_zonegroup.sync_flow_create" || flowOperation.Risk != "high" || flowOperation.ResourceKey != op.ResourceKey || flowOperation.LockKey != op.LockKey {
			t.Fatalf("flow operation: %+v %v", flowOperation, err)
		}
		if deleted.Code != http.StatusAccepted {
			t.Fatalf("zonegroup delete queue: %d %s", deleted.Code, deleted.Body.String())
		}
		deletion, err := db.FindOperation(context.Background(), operationIDFromResponse(t, deleted))
		if err != nil || deletion.Action != "rgw_zonegroup.sync_group_delete" || deletion.Risk != "high" || deletion.ResourceKey != op.ResourceKey || deletion.LockKey != op.LockKey {
			t.Fatalf("wrong deletion: %+v %v", deletion, err)
		}
	}
	for _, tenant := range []string{"", "team"} {
		id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00same-bucket"))
		nativeMux := http.NewServeMux()
		router.Register(nativeMux, handler.New(handler.Dependencies{Clusters: clusters, Operations: operations, Database: database, AuthEnabled: func() bool { return false }}))
		nativeResponse := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/group", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","expected_status":"allowed","status":"enabled"}`, cluster.ID, id), "sync-group-"+tenant)
		if nativeResponse.Code != http.StatusAccepted {
			t.Fatalf("native sync group should not require S3 endpoint: %d %s", nativeResponse.Code, nativeResponse.Body.String())
		}
		nativeOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, nativeResponse))
		if err != nil || nativeOperation.Action != "rgw_bucket.sync_group" || nativeOperation.Risk != "high" || nativeOperation.ResourceKey != "rgw/bucket/"+id {
			t.Fatalf("wrong sync group operation: %+v %v", nativeOperation, err)
		}
		createResponse := sendOperationRequest(t, nativeMux, http.MethodPost, "/api/v1/rgw/bucket/sync/group", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"new","status":"allowed"}`, cluster.ID, id), "sync-group-create-"+tenant)
		if createResponse.Code != http.StatusAccepted {
			t.Fatalf("sync group creation queue: %d %s", createResponse.Code, createResponse.Body.String())
		}
		createdGroup, err := db.FindOperation(context.Background(), operationIDFromResponse(t, createResponse))
		if err != nil || createdGroup.Action != "rgw_bucket.sync_group_create" || createdGroup.Risk != "high" || createdGroup.ResourceKey != nativeOperation.ResourceKey || createdGroup.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong sync group creation: %+v %v", createdGroup, err)
		}
		deletedResponse := sendOperationRequest(t, nativeMux, http.MethodDelete, "/api/v1/rgw/bucket/sync/group", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","expected_group":%q}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{},"pipes":[]}`), "sync-group-delete-"+tenant)
		if deletedResponse.Code != http.StatusAccepted {
			t.Fatalf("sync group deletion queue: %d %s", deletedResponse.Code, deletedResponse.Body.String())
		}
		deletedGroup, err := db.FindOperation(context.Background(), operationIDFromResponse(t, deletedResponse))
		if err != nil || deletedGroup.Action != "rgw_bucket.sync_group_delete" || deletedGroup.Risk != "high" || deletedGroup.ResourceKey != nativeOperation.ResourceKey || deletedGroup.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong sync group deletion: %+v %v", deletedGroup, err)
		}
		flowResponse := sendOperationRequest(t, nativeMux, http.MethodPost, "/api/v1/rgw/bucket/sync/flow", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","expected_group":%q,"flow_type":"directional","source_zone":"a","dest_zone":"b"}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{},"pipes":[]}`), "sync-flow-create-"+tenant)
		if flowResponse.Code != http.StatusAccepted {
			t.Fatalf("sync flow queue: %d %s", flowResponse.Code, flowResponse.Body.String())
		}
		flowOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, flowResponse))
		if err != nil || flowOperation.Action != "rgw_bucket.sync_flow_create" || flowOperation.Risk != "high" || flowOperation.ResourceKey != nativeOperation.ResourceKey || flowOperation.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong sync flow operation: %+v %v", flowOperation, err)
		}
		flowUpdateResponse := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/flow", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","expected_group":%q,"flow_id":"f","zones":["b"]}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{"symmetrical":[{"id":"f","zones":["A"]}]},"pipes":[]}`), "sync-flow-update-"+tenant)
		if flowUpdateResponse.Code != http.StatusAccepted {
			t.Fatalf("sync flow update queue: %d %s", flowUpdateResponse.Code, flowUpdateResponse.Body.String())
		}
		flowUpdate, err := db.FindOperation(context.Background(), operationIDFromResponse(t, flowUpdateResponse))
		if err != nil || flowUpdate.Action != "rgw_bucket.sync_flow_update" || flowUpdate.Risk != "high" || flowUpdate.ResourceKey != nativeOperation.ResourceKey || flowUpdate.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong sync flow update: %+v %v", flowUpdate, err)
		}
		flowDeleteResponse := sendOperationRequest(t, nativeMux, http.MethodDelete, "/api/v1/rgw/bucket/sync/flow", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","expected_group":%q,"flow_type":"symmetrical","flow_id":"f"}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{"symmetrical":[{"id":"f","zones":["A"]}]},"pipes":[]}`), "sync-flow-delete-"+tenant)
		if flowDeleteResponse.Code != http.StatusAccepted {
			t.Fatalf("sync flow deletion queue: %d %s", flowDeleteResponse.Code, flowDeleteResponse.Body.String())
		}
		flowDelete, err := db.FindOperation(context.Background(), operationIDFromResponse(t, flowDeleteResponse))
		if err != nil || flowDelete.Action != "rgw_bucket.sync_flow_delete" || flowDelete.Risk != "high" || flowDelete.ResourceKey != nativeOperation.ResourceKey || flowDelete.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong sync flow deletion: %+v %v", flowDelete, err)
		}
		pipeResponse := sendOperationRequest(t, nativeMux, http.MethodDelete, "/api/v1/rgw/bucket/sync/pipe", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","pipe_id":"p","expected_group":%q}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{},"pipes":[{"id":"p"}]}`), "sync-pipe-delete-"+tenant)
		if pipeResponse.Code != http.StatusAccepted {
			t.Fatalf("pipe deletion queue: %d %s", pipeResponse.Code, pipeResponse.Body.String())
		}
		pipeOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, pipeResponse))
		if err != nil || pipeOperation.Action != "rgw_bucket.sync_pipe_delete" || pipeOperation.Risk != "high" || pipeOperation.ResourceKey != nativeOperation.ResourceKey || pipeOperation.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong pipe operation: %+v %v", pipeOperation, err)
		}
		partialPipe := sendOperationRequest(t, nativeMux, http.MethodDelete, "/api/v1/rgw/bucket/sync/pipe", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","pipe_id":"p","expected_group":"{}","source_zones":["a"]}`, cluster.ID, id), "sync-pipe-partial-"+tenant)
		if partialPipe.Code != http.StatusBadRequest {
			t.Fatalf("partial pipe deletion accepted: %d %s", partialPipe.Code, partialPipe.Body.String())
		}
		pipeCreateResponse := sendOperationRequest(t, nativeMux, http.MethodPost, "/api/v1/rgw/bucket/sync/pipe", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","pipe_id":"new","expected_group":%q,"source_zones":["*"],"dest_zones":["*"],"source_bucket":"*","dest_bucket":"*","mode":"system"}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{},"pipes":[]}`), "sync-pipe-create-"+tenant)
		if pipeCreateResponse.Code != http.StatusAccepted {
			t.Fatalf("pipe creation queue: %d %s", pipeCreateResponse.Code, pipeCreateResponse.Body.String())
		}
		pipeCreated, err := db.FindOperation(context.Background(), operationIDFromResponse(t, pipeCreateResponse))
		if err != nil || pipeCreated.Action != "rgw_bucket.sync_pipe_create" || pipeCreated.Risk != "high" || pipeCreated.ResourceKey != nativeOperation.ResourceKey || pipeCreated.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong pipe creation: %+v %v", pipeCreated, err)
		}
		pipeUpdateBody := fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","pipe_id":"p","expected_group":%q,"source_bucket":"*","dest_bucket":"photos","mode":"system"}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{},"pipes":[{"id":"p"}]}`)
		pipeUpdateResponse := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe", pipeUpdateBody, "sync-pipe-update-"+tenant)
		for index, value := range []string{`"team$ns$u"`, `""`, "null", "7"} {
			response := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe", strings.TrimSuffix(pipeUpdateBody, "}")+`,"dest_owner":`+value+`}`, fmt.Sprintf("sync-acl-%s-%d", tenant, index))
			want := http.StatusAccepted
			if index >= 2 {
				want = http.StatusBadRequest
			}
			if response.Code != want {
				t.Fatalf("acl %s: %d %s", value, response.Code, response.Body.String())
			}
		}
		for index, extra := range []string{`,"tags_add":[{"key":"k","value":""}]`, `,"tags_remove":[{"key":"k","value":"v"}]`, `,"tags_add":null`, `,"tags_add":[{"key":"k"}]`, `,"tags_add":[{"key":"k","value":1}]`} {
			response := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe", strings.TrimSuffix(pipeUpdateBody, "}")+extra+`}`, fmt.Sprintf("sync-tags-%s-%d", tenant, index))
			want := http.StatusAccepted
			if index >= 2 {
				want = http.StatusBadRequest
			}
			if response.Code != want {
				t.Fatalf("tags %s: %d %s", extra, response.Code, response.Body.String())
			}
		}
		for index, extra := range []string{`,"prefix_mode":"set","source_prefix":""`, `,"prefix_mode":"remove"`, `,"prefix_mode":"unknown"`, `,"prefix_mode":"set","source_prefix":7`} {
			response := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe", strings.TrimSuffix(pipeUpdateBody, "}")+extra+`}`, fmt.Sprintf("sync-prefix-%s-%d", tenant, index))
			want := http.StatusAccepted
			if index >= 2 {
				want = http.StatusBadRequest
			}
			if response.Code != want {
				t.Fatalf("prefix %s: %d %s", extra, response.Code, response.Body.String())
			}
		}
		for index, value := range []string{`"COLD"`, `""`, "null", "7"} {
			response := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe", strings.TrimSuffix(pipeUpdateBody, "}")+`,"storage_class":`+value+`}`, fmt.Sprintf("sync-storage-class-%s-%d", tenant, index))
			want := http.StatusAccepted
			if value == "null" || value == "7" {
				want = http.StatusBadRequest
			}
			if response.Code != want {
				t.Fatalf("storage class %s: %d %s", value, response.Code, response.Body.String())
			}
		}
		for _, priority := range []string{"0", "-2147483648", "2147483647", "0.5", `"0"`, "null"} {
			response := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe", strings.TrimSuffix(pipeUpdateBody, "}")+`,"priority":`+priority+`}`, "sync-priority-"+tenant+priority)
			want := http.StatusAccepted
			if priority == "0.5" || priority == `"0"` || priority == "null" {
				want = http.StatusBadRequest
			}
			if response.Code != want {
				t.Fatalf("priority %s: %d %s", priority, response.Code, response.Body.String())
			}
		}
		if pipeUpdateResponse.Code != http.StatusAccepted {
			t.Fatalf("pipe update queue: %d %s", pipeUpdateResponse.Code, pipeUpdateResponse.Body.String())
		}
		pipeUpdated, err := db.FindOperation(context.Background(), operationIDFromResponse(t, pipeUpdateResponse))
		if err != nil || pipeUpdated.Action != "rgw_bucket.sync_pipe_update" || pipeUpdated.Risk != "high" || pipeUpdated.ResourceKey != nativeOperation.ResourceKey || pipeUpdated.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong pipe update: %+v %v", pipeUpdated, err)
		}
		pipeZonesBody := fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"group_id":"g","pipe_id":"p","expected_group":%q,"source_zones":["a"],"dest_zones":["*"]}`, cluster.ID, id, `{"id":"g","status":"allowed","data_flow":{},"pipes":[{"id":"p"}]}`)
		pipeZonesResponse := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe/zones", pipeZonesBody, "sync-pipe-zones-"+tenant)
		if pipeZonesResponse.Code != http.StatusAccepted {
			t.Fatalf("pipe zones queue: %d %s", pipeZonesResponse.Code, pipeZonesResponse.Body.String())
		}
		pipeZones, err := db.FindOperation(context.Background(), operationIDFromResponse(t, pipeZonesResponse))
		if err != nil || pipeZones.Action != "rgw_bucket.sync_pipe_zones" || pipeZones.Risk != "high" || pipeZones.ResourceKey != nativeOperation.ResourceKey || pipeZones.LockKey != nativeOperation.LockKey {
			t.Fatalf("wrong pipe zones: %+v %v", pipeZones, err)
		}
		invalidZones := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe/zones", strings.TrimSuffix(pipeZonesBody, "}")+`,"mode":"system"}`, "invalid-pipe-zones-"+tenant)
		if invalidZones.Code != http.StatusBadRequest {
			t.Fatalf("unexpected mode accepted: %d", invalidZones.Code)
		}
		invalidPipeUpdate := sendOperationRequest(t, nativeMux, http.MethodPatch, "/api/v1/rgw/bucket/sync/pipe", strings.TrimSuffix(pipeUpdateBody, "}")+`,"source_zones":["*"]}`, "invalid-pipe-update-"+tenant)
		if invalidPipeUpdate.Code != http.StatusBadRequest {
			t.Fatalf("unexpected zone update accepted: %d", invalidPipeUpdate.Code)
		}
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/bucket", fmt.Sprintf(`{"cluster_id":%d,"name":"same-bucket","tenant":%q}`, cluster.ID, tenant), "bucket-create-"+tenant)
		if response.Code != http.StatusAccepted {
			t.Fatalf("bucket creation queue: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Action != "rgw_bucket.create" || row.ResourceKey != "rgw/bucket/"+id {
			t.Fatalf("incorrect creation identity: %+v %v", row, err)
		}
		response = sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/bucket", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"versioning":"enabled"}`, cluster.ID, id), "bucket-update-"+tenant)
		if response.Code != http.StatusAccepted {
			t.Fatalf("bucket update queue: %d %s", response.Code, response.Body.String())
		}
		updated, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || updated.ResourceKey != row.ResourceKey {
			t.Fatalf("creation and update identities differ: %+v %v", updated, err)
		}
		response = sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/bucket/acl", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"acl":"private"}`, cluster.ID, id), "bucket-acl-"+tenant)
		if response.Code != http.StatusAccepted {
			t.Fatalf("ACL queue: %d %s", response.Code, response.Body.String())
		}
		aclOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || aclOperation.Action != "rgw_bucket.acl" || aclOperation.Risk != "high" || aclOperation.ResourceKey != row.ResourceKey || aclOperation.LockKey != updated.LockKey {
			t.Fatalf("incorrect ACL operation: %+v %v", aclOperation, err)
		}
		response = sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/bucket/replication", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"expected_document":""}`, cluster.ID, id), "bucket-replication-"+tenant)
		if response.Code != http.StatusAccepted {
			t.Fatalf("replication queue: %d %s", response.Code, response.Body.String())
		}
		replicationOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || replicationOperation.Action != "rgw_bucket.replication_enable" || replicationOperation.Risk != "high" || replicationOperation.ResourceKey != row.ResourceKey || replicationOperation.LockKey != updated.LockKey {
			t.Fatalf("incorrect replication operation: %+v %v", replicationOperation, err)
		}
		for _, mode := range []string{"single", "all"} {
			response = sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/bucket/notification", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"mode":%q,"notification_id":"","expected_document":"<NotificationConfiguration/>"}`, cluster.ID, id, mode), "bucket-notification-"+tenant+"-"+mode)
			if response.Code != http.StatusAccepted {
				t.Fatalf("notification queue: %d %s", response.Code, response.Body.String())
			}
			notificationOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
			if err != nil || notificationOperation.Action != "rgw_bucket.notification_delete" || notificationOperation.Risk != "high" || notificationOperation.ResourceKey != row.ResourceKey || notificationOperation.LockKey != updated.LockKey {
				t.Fatalf("incorrect notification operation: %+v %v", notificationOperation, err)
			}
		}
		response = sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/bucket/notification", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"mode":"create","expected_document":"<NotificationConfiguration/>","rule":{"id":"id","topic":"arn:aws:sns:east::topic","events":[],"filters":[]}}`, cluster.ID, id), "bucket-notification-set-"+tenant)
		if response.Code != http.StatusAccepted {
			t.Fatalf("notification write queue: %d %s", response.Code, response.Body.String())
		}
		notificationWrite, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || notificationWrite.Action != "rgw_bucket.notification_set" || notificationWrite.Risk != "high" || notificationWrite.ResourceKey != row.ResourceKey || notificationWrite.LockKey != updated.LockKey {
			t.Fatalf("incorrect notification write identity: %+v %v", notificationWrite, err)
		}
		response = sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/bucket/mfa", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"status":"Enabled","mfa_delete":"Enabled","expected_document":"<VersioningConfiguration/>","mfa_serial_secret":"private-device-serial","mfa_token":"009876"}`, cluster.ID, id), "bucket-mfa-"+tenant)
		if response.Code != http.StatusAccepted {
			t.Fatalf("MFA queue: %d %s", response.Code, response.Body.String())
		}
		mfaOperation, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || mfaOperation.Action != "rgw_bucket.mfa" || mfaOperation.Risk != "high" || mfaOperation.ResourceKey != row.ResourceKey || mfaOperation.LockKey != updated.LockKey {
			t.Fatalf("wrong MFA identity: %+v %v", mfaOperation, err)
		}
		for _, secret := range []string{"private-device-serial", "009876"} {
			if strings.Contains(response.Body.String(), secret) || strings.Contains(mfaOperation.ParametersCiphertext, secret) {
				t.Fatal("MFA secret exposed")
			}
		}
		mfaPlain, err := security.Decrypt(mfaOperation.ParametersCiphertext, contractKey)
		if err != nil || !strings.Contains(string(mfaPlain), "009876") || !strings.Contains(string(mfaPlain), "private-device-serial") {
			t.Fatal("MFA credentials not preserved encrypted")
		}
		for _, kind := range []string{"policy", "cors", "lifecycle", "encryption", "tagging"} {
			response := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/bucket/policy", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"kind":%q,"document":"raw"}`, cluster.ID, id, kind), "bucket-config-"+tenant+"-"+kind)
			if response.Code != http.StatusAccepted {
				t.Fatalf("bucket configuration queue: %d %s", response.Code, response.Body.String())
			}
			row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
			if err != nil || row.Action != "rgw_bucket_policy.update" || row.ResourceKey != "rgw/bucket/"+id+"/policy" {
				t.Fatalf("incorrect bucket configuration identity: %+v %v", row, err)
			}
			response = sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/bucket/policy", fmt.Sprintf(`{"cluster_id":%d,"bucket_id":%q,"kind":%q}`, cluster.ID, id, kind), "bucket-config-delete-"+tenant+"-"+kind)
			if response.Code != http.StatusAccepted {
				t.Fatalf("delete configuration queue: %d %s", response.Code, response.Body.String())
			}
			row, err = db.FindOperation(context.Background(), operationIDFromResponse(t, response))
			if err != nil || row.Action != "rgw_bucket_policy.delete" || row.Risk != "high" || row.ResourceKey != "rgw/bucket/"+id+"/policy" {
				t.Fatalf("unsafe deletion queue: %+v %v", row, err)
			}
		}
	}
	for index, tc := range []struct{ fields, risk string }{
		{`"email":"before@example.test"`, "medium"},
		{`"system":true`, "high"},
		{`"system":false`, "high"},
		{`"account_root":true,"expected_account_id":"RGW12345678901234567"`, "high"},
		{`"account_root":false,"expected_account_id":"RGW12345678901234567"`, "high"},
		{`"target_account_id":"RGW12345678901234567","migration_confirm_uid":"tenant$user"`, "high"},
		{`"email":"after@example.test"`, "medium"},
	} {
		response := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/user", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user",%s}`, cluster.ID, tc.fields), fmt.Sprintf("user-risk-%d", index))
		if response.Code != http.StatusAccepted {
			t.Fatalf("user update: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != tc.risk || row.Action != "rgw_user.update" || row.ResourceKey != "rgw/user/tenant$user" {
			t.Fatalf("incorrect user risk or identity: %+v %v", row, err)
		}
		if !strings.Contains(response.Body.String(), `"risk":"`+tc.risk+`"`) {
			t.Fatalf("response risk differs: %s", response.Body.String())
		}
	}
	for _, action := range []string{"add", "rm", "replace"} {
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user/caps", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","action":%q,"type":"users","permission":"read"}`, cluster.ID, action), "caps-"+action)
		if response.Code != http.StatusAccepted {
			t.Fatalf("caps mutation: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Action != "rgw_user.caps" || row.ResourceKey != "rgw/user/tenant$user" {
			t.Fatalf("wrong caps operation: %+v %v", row, err)
		}
	}
	for _, action := range []string{"modify", "rm", "rotate-swift-key"} {
		fields := ""
		if action == "modify" {
			fields = `,"subuser_permission":"read"`
		}
		if action == "rotate-swift-key" {
			fields = `,"expected_key_active":false,"secret_key":"NewSwiftSecret"`
		}
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user/subuser", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","subuser":"swift","confirm_subuser":"tenant$user:swift","action":%q%s}`, cluster.ID, action, fields), "subuser-"+action)
		if response.Code != http.StatusAccepted {
			t.Fatalf("subuser mutation: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "rgw_user.subuser" || row.ResourceKey != "rgw/user/tenant$user" {
			t.Fatalf("wrong subuser operation: %+v %v", row, err)
		}
	}
	for _, sub := range []string{"", "sub"} {
		owner, fields := "tenant$user", ""
		if sub != "" {
			owner += ":" + sub
			fields = `,"subuser":"sub"`
		}
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user/key", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","confirm_owner":%q,"access_key":"ACCESS123","secret_key":"PrivateKeySecret"%s}`, cluster.ID, owner, fields), "s3-key-"+sub)
		if response.Code != http.StatusAccepted {
			t.Fatalf("S3 key creation: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "rgw_key.create" || row.ResourceKey != "rgw/user/tenant$user/key" {
			t.Fatalf("wrong S3 key operation: %+v %v", row, err)
		}
		if strings.Contains(response.Body.String(), "PrivateKeySecret") || strings.Contains(row.ParametersCiphertext, "PrivateKeySecret") {
			t.Fatal("S3 creation credential exposed")
		}
		response = sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/rgw/user/key", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","confirm_owner":%q,"access_key":"ACCESS123"%s}`, cluster.ID, owner, fields), "s3-key-delete-"+sub)
		if response.Code != http.StatusAccepted {
			t.Fatalf("S3 key deletion: %d %s", response.Code, response.Body.String())
		}
		row, err = db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "rgw_key.delete" || row.ResourceKey != "rgw/user/tenant$user/key" {
			t.Fatalf("wrong key deletion: %+v %v", row, err)
		}
		response = sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/user/key", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","confirm_owner":%q,"access_key":"ACCESS123","secret_key":"RotatedSecret"%s}`, cluster.ID, owner, fields), "s3-key-rotate-"+sub)
		if response.Code != http.StatusAccepted {
			t.Fatalf("S3 key rotation: %d %s", response.Code, response.Body.String())
		}
		row, err = db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "rgw_key.update" || row.ResourceKey != "rgw/user/tenant$user/key" {
			t.Fatalf("wrong key rotation: %+v %v", row, err)
		}
	}
	for _, system := range []bool{true, false} {
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user", fmt.Sprintf(`{"cluster_id":%d,"uid":"new-flags-user","display_name":"User","system":%t,"suspended":true}`, cluster.ID, system), fmt.Sprintf("create-flags-user-%t", system))
		if response.Code != http.StatusAccepted {
			t.Fatalf("flag creation: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		want := "medium"
		if system {
			want = "high"
		}
		if err != nil || row.Risk != want || row.Action != "rgw_user.create" || !strings.Contains(response.Body.String(), `"risk":"`+want+`"`) {
			t.Fatalf("wrong creation risk: %+v %v", row, err)
		}
	}
	for _, account := range []bool{false, true} {
		fields := ""
		if account {
			fields = `,"account_id":"RGW12345678901234567","account_root":false`
		}
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user", fmt.Sprintf(`{"cluster_id":%d,"uid":"new-key-user","display_name":"valid-name","access_key":"ACCESS123","secret_key":"SavedCreationSecret"%s}`, cluster.ID, fields), fmt.Sprintf("create-key-user-%t", account))
		if response.Code != http.StatusAccepted {
			t.Fatalf("user key creation: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Action != "rgw_user.create" || strings.Contains(row.ParametersCiphertext, "SavedCreationSecret") || strings.Contains(response.Body.String(), "SavedCreationSecret") || strings.Contains(response.Body.String(), "ACCESS123") {
			t.Fatalf("unsafe creation response: %+v %v", row, err)
		}
	}
	for _, root := range []bool{false, true} {
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user", fmt.Sprintf(`{"cluster_id":%d,"uid":"new-user","display_name":"valid-name","account_id":"RGW12345678901234567","account_root":%t}`, cluster.ID, root), fmt.Sprintf("create-account-user-%t", root))
		if response.Code != http.StatusAccepted {
			t.Fatalf("account user create: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		want := "medium"
		if root {
			want = "high"
		}
		if err != nil || row.Risk != want {
			t.Fatalf("wrong creation risk: %+v %v", row, err)
		}
	}
	for _, kind := range []string{"s3", "swift"} {
		fields := ""
		if kind == "s3" {
			fields = `,"access_key":"ACCESS123"`
		}
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user/subuser", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","subuser":"new","confirm_subuser":"tenant$user:new","action":"create","subuser_permission":"read","key_type":%q,"secret_key":"PrivateTestSecret"%s}`, cluster.ID, kind, fields), "subuser-create-"+kind)
		if response.Code != http.StatusAccepted {
			t.Fatalf("subuser create: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "rgw_user.subuser" || row.ResourceKey != "rgw/user/tenant$user" {
			t.Fatalf("wrong creation operation: %+v %v", row, err)
		}
		if strings.Contains(row.ParametersCiphertext, "PrivateTestSecret") || strings.Contains(response.Body.String(), "PrivateTestSecret") || strings.Contains(response.Body.String(), "ACCESS123") {
			t.Fatal("creation credential exposed")
		}
	}
	for _, action := range []string{"start", "stop", "restart", "redeploy", "reconfig", "rotate-key"} {
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/daemon/action", fmt.Sprintf(`{"cluster_id":%d,"name":"osd.1","action":%q}`, cluster.ID, action), "daemon-"+action)
		if response.Code != http.StatusAccepted {
			t.Fatalf("daemon action %s: %d %s", action, response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "daemon.action" || row.ResourceKey != "daemon/osd.1/action" || row.Status != store.OperationQueued {
			t.Fatalf("incorrect daemon action risk or target: %+v %v", row, err)
		}
	}

	queued := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/pool", fmt.Sprintf(`{"cluster_id":%d,"name":"data"}`, cluster.ID), "create-data")
	if queued.Code != http.StatusAccepted {
		t.Fatalf("queue status=%d body=%s", queued.Code, queued.Body.String())
	}
	operationID := operationIDFromResponse(t, queued)
	row, err := db.FindOperation(context.Background(), operationID)
	if err != nil || row.Status != store.OperationQueued || strings.Contains(row.ParametersCiphertext, `"name":"data"`) {
		t.Fatalf("queued operation=%#v err=%v", row, err)
	}

	replayed := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/pool", fmt.Sprintf(`{"cluster_id":%d,"name":"data"}`, cluster.ID), "create-data")
	if replayed.Code != http.StatusAccepted || operationIDFromResponse(t, replayed) != operationID {
		t.Fatalf("idempotent replay status=%d body=%s", replayed.Code, replayed.Body.String())
	}
	conflict := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/pool", fmt.Sprintf(`{"cluster_id":%d,"name":"data"}`, cluster.ID), "create-data")
	if conflict.Code != http.StatusConflict {
		t.Fatalf("idempotency conflict status=%d body=%s", conflict.Code, conflict.Body.String())
	}

	get := sendOperationRequest(t, mux, http.MethodGet, "/api/v1/operation", fmt.Sprintf(`{"cluster_id":%d,"operation_id":%d}`, cluster.ID, operationID), "")
	if get.Code != http.StatusOK || strings.Contains(get.Body.String(), "parameters_ciphertext") {
		t.Fatalf("get operation status=%d body=%s", get.Code, get.Body.String())
	}
	list := sendOperationRequest(t, mux, http.MethodGet, "/api/v1/operations?status=queued&limit=10", fmt.Sprintf(`{"cluster_id":%d}`, cluster.ID), "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"operation_id":`) {
		t.Fatalf("list operations status=%d body=%s", list.Code, list.Body.String())
	}
}

func sendOperationRequest(t *testing.T, mux http.Handler, method, path, body, idempotencyKey string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func operationIDFromResponse(t *testing.T, response *httptest.ResponseRecorder) uint64 {
	t.Helper()
	var envelope struct {
		Data struct {
			OperationID uint64 `json:"operation_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope.Data.OperationID == 0 {
		t.Fatalf("decode operation response: id=%d err=%v body=%s", envelope.Data.OperationID, err, response.Body.String())
	}
	return envelope.Data.OperationID
}
