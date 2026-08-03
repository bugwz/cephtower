CREATE TABLE IF NOT EXISTS schema_migration (
  version TEXT PRIMARY KEY,
  checksum TEXT NOT NULL,
  applied_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE TABLE setting (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE TABLE user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL DEFAULT '',
  email TEXT NULL,
  password TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  last_login_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_user_status ON user(status);
CREATE INDEX idx_user_email ON user(email);
CREATE TABLE password_reset_code (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
  code_hash TEXT NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  consumed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_reset_user_expiry ON password_reset_code(user_id, expires_at);
CREATE INDEX idx_reset_expiry_consumed ON password_reset_code(expires_at, consumed_at);
CREATE TABLE user_session (
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  source_ip TEXT NULL,
  user_agent TEXT NULL,
  expires_at DATETIME NOT NULL,
  last_seen_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  revoked_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_session_user_expiry ON user_session(user_id, expires_at);
CREATE INDEX idx_session_expiry_revoked ON user_session(expires_at, revoked_at);
CREATE TABLE role (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE TABLE ceph_cluster (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  monitor_addresses TEXT NOT NULL,
  client_username TEXT NOT NULL,
  client_key TEXT NOT NULL,
  discovered_data TEXT NOT NULL DEFAULT '{}',
  fsid TEXT NULL,
  ceph_version TEXT NULL,
  status TEXT NOT NULL DEFAULT 'unknown',
  enabled INTEGER NOT NULL DEFAULT 1 CHECK(enabled IN (0,1)),
  generation INTEGER NOT NULL DEFAULT 0,
  last_seen_at DATETIME NULL,
  last_error_code TEXT NULL,
  last_error_message TEXT NULL,
  observed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_cluster_fsid ON ceph_cluster(fsid);
CREATE INDEX idx_cluster_status_enabled ON ceph_cluster(status, enabled);
CREATE INDEX idx_cluster_last_seen ON ceph_cluster(last_seen_at);
CREATE TABLE user_role_binding (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
  role_id INTEGER NOT NULL REFERENCES role(id) ON DELETE CASCADE,
  cluster_id INTEGER NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  scope_key TEXT NOT NULL,
  created_by_user_id INTEGER NULL REFERENCES user(id) ON DELETE SET NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(user_id, role_id, scope_key)
);
CREATE INDEX idx_binding_user_cluster ON user_role_binding(user_id, cluster_id);
CREATE INDEX idx_binding_role ON user_role_binding(role_id);
CREATE INDEX idx_binding_created_by ON user_role_binding(created_by_user_id);
CREATE TABLE ceph_cluster_credential (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id INTEGER NOT NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  credential TEXT NOT NULL,
  fingerprint TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(cluster_id, kind)
);
CREATE TABLE ceph_cluster_endpoint (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id INTEGER NOT NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT 'default',
  url TEXT NOT NULL,
  tls_mode TEXT NOT NULL,
  ca_credential_id INTEGER NULL REFERENCES ceph_cluster_credential(id) ON DELETE SET NULL,
  config_json TEXT NULL,
  enabled INTEGER NOT NULL DEFAULT 1 CHECK(enabled IN (0,1)),
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(cluster_id, kind, name)
);
CREATE INDEX idx_endpoint_enabled ON ceph_cluster_endpoint(cluster_id, enabled);
CREATE INDEX idx_endpoint_ca ON ceph_cluster_endpoint(ca_credential_id);
CREATE TABLE ceph_cluster_capability (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id INTEGER NOT NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  supported INTEGER NOT NULL CHECK(supported IN (0,1)),
  reason TEXT NULL,
  version TEXT NULL,
  details_json TEXT NULL,
  observed_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(cluster_id, name)
);
CREATE INDEX idx_capability_supported ON ceph_cluster_capability(cluster_id, supported);
CREATE INDEX idx_capability_observed ON ceph_cluster_capability(observed_at);
CREATE TABLE ceph_host (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id INTEGER NOT NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  hostname TEXT NOT NULL,
  ssh_address TEXT NOT NULL,
  ssh_port INTEGER NOT NULL DEFAULT 22,
  ssh_user TEXT NOT NULL,
  ssh_password_secret TEXT NULL,
  address TEXT NULL,
  status TEXT NULL,
  configured_data TEXT NULL,
  discovered_data TEXT NOT NULL DEFAULT '{}',
  generation INTEGER NOT NULL DEFAULT 0,
  resource_version INTEGER NOT NULL DEFAULT 1,
  source TEXT NOT NULL DEFAULT '',
  source_version TEXT NULL,
  observed_at DATETIME NULL,
  stale_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(cluster_id, hostname)
);
CREATE INDEX idx_ceph_host_cluster ON ceph_host(cluster_id, hostname);
CREATE INDEX idx_ceph_host_observed ON ceph_host(observed_at);
CREATE INDEX idx_ceph_host_stale ON ceph_host(stale_at);
CREATE INDEX idx_ceph_host_status ON ceph_host(status);
CREATE TABLE ceph_collection_run (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id INTEGER NOT NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  module TEXT NOT NULL,
  generation INTEGER NOT NULL,
  status TEXT NOT NULL,
  source TEXT NOT NULL,
  record_count INTEGER NOT NULL DEFAULT 0,
  error_code TEXT NULL,
  error_message TEXT NULL,
  started_at DATETIME NOT NULL,
  finished_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE(cluster_id, module, generation)
);
CREATE INDEX idx_collection_module_started ON ceph_collection_run(cluster_id, module, started_at);
CREATE INDEX idx_collection_status_started ON ceph_collection_run(status, started_at);
CREATE INDEX idx_collection_finished ON ceph_collection_run(finished_at);
CREATE TABLE audit_event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  occurred_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  event_type TEXT NOT NULL,
  request_id TEXT NOT NULL,
  actor_user_id INTEGER NULL REFERENCES user(id) ON DELETE SET NULL,
  actor_username TEXT NOT NULL,
  cluster_id INTEGER NULL REFERENCES ceph_cluster(id) ON DELETE SET NULL,
  cluster_name TEXT NULL,
  action TEXT NOT NULL,
  resource_kind TEXT NULL,
  resource_key TEXT NULL,
  risk TEXT NULL,
  outcome TEXT NOT NULL,
  http_status INTEGER NULL,
  error_code TEXT NULL,
  source_ip TEXT NULL,
  user_agent TEXT NULL,
  before_generation INTEGER NULL,
  after_generation INTEGER NULL,
  parameters_json TEXT NULL,
  details_json TEXT NULL,
  previous_hash TEXT NULL,
  event_hash TEXT NOT NULL UNIQUE
);
CREATE INDEX idx_audit_occurred ON audit_event(occurred_at);
CREATE INDEX idx_audit_actor_occurred ON audit_event(actor_user_id, occurred_at);
CREATE INDEX idx_audit_cluster_occurred ON audit_event(cluster_id, occurred_at);
CREATE INDEX idx_audit_action_occurred ON audit_event(action, occurred_at);
CREATE INDEX idx_audit_resource_occurred ON audit_event(resource_kind, resource_key, occurred_at);
CREATE INDEX idx_audit_request ON audit_event(request_id);
