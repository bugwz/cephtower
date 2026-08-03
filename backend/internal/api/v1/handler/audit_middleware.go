package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

const auditMaxBody = 1 << 20

type auditContextKey struct{}

type auditInfo struct {
	Action       string
	ResourceKind string
	ResourceKey  string
	Risk         string
	ClusterID    *uint64
}

type auditResponseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *auditResponseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *auditResponseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	if r.body.Len() < auditMaxBody {
		_, _ = r.body.Write(data[:min(len(data), auditMaxBody-r.body.Len())])
	}
	return r.ResponseWriter.Write(data)
}

func (r *auditResponseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *auditResponseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func withAuditInfo(ctx context.Context, info *auditInfo) context.Context {
	return context.WithValue(ctx, auditContextKey{}, info)
}

func auditInfoFromRequest(r *http.Request) *auditInfo {
	info, _ := r.Context().Value(auditContextKey{}).(*auditInfo)
	return info
}

func annotateAudit(r *http.Request, action, kind, key, risk string, clusterID *uint64) {
	info := auditInfoFromRequest(r)
	if info == nil {
		return
	}
	if action != "" {
		info.Action = action
	}
	if kind != "" {
		info.ResourceKind = kind
	}
	if key != "" {
		info.ResourceKey = key
	}
	if risk != "" {
		info.Risk = risk
	}
	if clusterID != nil {
		info.ClusterID = clusterID
	}
}

func (h *Handler) auditRequest(next http.Handler, w http.ResponseWriter, r *http.Request, body []byte, started time.Time) {
	recorder := &auditResponseRecorder{ResponseWriter: w}
	next.ServeHTTP(recorder, r)
	h.recordAuditEvent(r, recorder, body, started)
}

func (h *Handler) recordAuditEvent(r *http.Request, recorder *auditResponseRecorder, body []byte, started time.Time) {
	if h == nil || h.Database == nil || h.Database() == nil {
		return
	}
	status := recorder.status
	if status == 0 {
		status = http.StatusOK
	}
	info := auditInfoFromRequest(r)
	if info == nil {
		info = &auditInfo{Action: r.Method + " " + strings.TrimPrefix(r.URL.Path, "/api/v1")}
	}
	clusterID := info.ClusterID
	parameters := auditParameters(body)
	if clusterID == nil {
		if id := auditClusterID(parameters); id != nil {
			clusterID = id
		}
	}
	var clusterName *string
	if clusterID != nil {
		if row, err := h.Database().FindCluster(r.Context(), *clusterID); err == nil {
			value := row.Name
			clusterName = &value
		}
	}
	var actorID *uint64
	actorUsername := "anonymous"
	if user, ok := CurrentUser(r); ok {
		if user.ID != 0 {
			actorID = &user.ID
		}
		actorUsername = user.Username
	}
	var parametersJSON *string
	if parameters != nil {
		if encoded, err := json.Marshal(parameters); err == nil {
			value := string(encoded)
			parametersJSON = &value
		}
	}
	errorCode := auditErrorCode(recorder.body.Bytes())
	details := map[string]any{
		"method":      r.Method,
		"path":        r.URL.Path,
		"duration_ms": time.Since(started).Milliseconds(),
	}
	if query := r.URL.RawQuery; query != "" {
		details["query"] = security.Redact(query)
	}
	detailsJSONBytes, _ := json.Marshal(details)
	detailsJSON := string(detailsJSONBytes)
	httpStatus := status
	outcome := "succeeded"
	if status == http.StatusAccepted {
		outcome = "accepted"
	} else if status >= 400 {
		outcome = "failed"
	}
	var resourceKind, resourceKey, risk *string
	if info.ResourceKind != "" {
		resourceKind = &info.ResourceKind
	}
	if info.ResourceKey != "" {
		resourceKey = &info.ResourceKey
	}
	if info.Risk != "" {
		risk = &info.Risk
	}
	var beforeGeneration *uint64
	if raw := strings.TrimSpace(r.Header.Get("If-Match")); raw != "" {
		if value, err := strconv.ParseUint(raw, 10, 64); err == nil {
			beforeGeneration = &value
		}
	}
	sourceIP := clientIP(r)
	userAgent := strings.TrimSpace(r.UserAgent())
	event := store.AuditEvent{
		OccurredAt:       time.Now().UTC(),
		EventType:        "api_request",
		RequestID:        RequestID(r),
		ActorUserID:      actorID,
		ActorUsername:    actorUsername,
		ClusterID:        clusterID,
		ClusterName:      clusterName,
		Action:           info.Action,
		ResourceKind:     resourceKind,
		ResourceKey:      resourceKey,
		Risk:             risk,
		Outcome:          outcome,
		HTTPStatus:       &httpStatus,
		ErrorCode:        errorCode,
		SourceIP:         optionalAuditString(sourceIP),
		UserAgent:        optionalAuditString(userAgent),
		BeforeGeneration: beforeGeneration,
		ParametersJSON:   parametersJSON,
		DetailsJSON:      &detailsJSON,
	}
	_ = h.Database().CreateAuditEvent(r.Context(), &event)
}

func auditParameters(body []byte) any {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return map[string]any{"raw": security.Redact(string(body))}
	}
	redacted, err := security.RedactJSON(value)
	if err != nil {
		return map[string]any{"raw": security.Redact(string(body))}
	}
	return redacted
}

func auditClusterID(parameters any) *uint64 {
	object, ok := parameters.(map[string]any)
	if !ok {
		return nil
	}
	value, ok := object["cluster_id"]
	if !ok {
		return nil
	}
	id, err := uintFromJSON(value)
	if err != nil || id == 0 {
		return nil
	}
	return &id
}

func auditErrorCode(data []byte) *string {
	var envelope struct {
		Data struct {
			ErrorCode string `json:"error_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.Data.ErrorCode == "" {
		return nil
	}
	return &envelope.Data.ErrorCode
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func optionalAuditString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
