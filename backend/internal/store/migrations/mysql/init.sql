CREATE TABLE IF NOT EXISTS schema_migration (version VARCHAR(64) PRIMARY KEY, checksum VARCHAR(64) NOT NULL, applied_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)) ENGINE=InnoDB;
CREATE TABLE setting (`key` VARCHAR(191) PRIMARY KEY, value LONGTEXT NOT NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)) ENGINE=InnoDB;
CREATE TABLE user (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, username VARCHAR(128) NOT NULL UNIQUE,
 display_name VARCHAR(255) NOT NULL DEFAULT '', email VARCHAR(320) NULL, password LONGTEXT NOT NULL,
 status VARCHAR(32) NOT NULL DEFAULT 'active', last_login_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 INDEX idx_user_status(status), INDEX idx_user_email(email)
) ENGINE=InnoDB;
CREATE TABLE password_reset_code (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, user_id BIGINT UNSIGNED NOT NULL, code_hash VARCHAR(64) NOT NULL UNIQUE,
 expires_at DATETIME(6) NOT NULL, consumed_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 INDEX idx_reset_user_expiry(user_id,expires_at), INDEX idx_reset_expiry_consumed(expires_at,consumed_at),
 FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE user_session (
 id VARCHAR(36) PRIMARY KEY, user_id BIGINT UNSIGNED NOT NULL, token_hash VARCHAR(64) NOT NULL UNIQUE,
 source_ip VARCHAR(64) NULL, user_agent VARCHAR(512) NULL, expires_at DATETIME(6) NOT NULL,
 last_seen_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), revoked_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 INDEX idx_session_user_expiry(user_id,expires_at), INDEX idx_session_expiry_revoked(expires_at,revoked_at),
 FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE role (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, name VARCHAR(128) NOT NULL UNIQUE, description LONGTEXT NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB;
CREATE TABLE ceph_cluster (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, name VARCHAR(128) NOT NULL UNIQUE,
 monitor_addresses VARCHAR(4096) NOT NULL, client_username VARCHAR(128) NOT NULL, client_key LONGTEXT NOT NULL,
 discovered_data LONGTEXT NOT NULL, fsid VARCHAR(36) NULL, ceph_version VARCHAR(128) NULL,
 status VARCHAR(32) NOT NULL DEFAULT 'unknown', enabled BOOLEAN NOT NULL DEFAULT TRUE, generation BIGINT UNSIGNED NOT NULL DEFAULT 0,
 last_seen_at DATETIME(6) NULL, last_error_code VARCHAR(64) NULL, last_error_message LONGTEXT NULL, observed_at DATETIME(6) NULL,
 INDEX idx_cluster_fsid(fsid), INDEX idx_cluster_status_enabled(status,enabled), INDEX idx_cluster_last_seen(last_seen_at),
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB;
CREATE TABLE user_role_binding (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, user_id BIGINT UNSIGNED NOT NULL, role_id BIGINT UNSIGNED NOT NULL,
 cluster_id BIGINT UNSIGNED NULL, scope_key VARCHAR(32) NOT NULL, created_by_user_id BIGINT UNSIGNED NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), UNIQUE KEY uq_user_role_scope(user_id,role_id,scope_key),
 INDEX idx_binding_user_cluster(user_id,cluster_id), INDEX idx_binding_role(role_id), INDEX idx_binding_created_by(created_by_user_id),
 FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE, FOREIGN KEY(role_id) REFERENCES role(id) ON DELETE CASCADE,
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE, FOREIGN KEY(created_by_user_id) REFERENCES user(id) ON DELETE SET NULL
) ENGINE=InnoDB;
CREATE TABLE ceph_cluster_credential (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, kind VARCHAR(64) NOT NULL,
 credential LONGTEXT NOT NULL, fingerprint VARCHAR(64) NOT NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 UNIQUE KEY uq_cluster_credential(cluster_id,kind), FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE ceph_cluster_endpoint (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, kind VARCHAR(64) NOT NULL,
 name VARCHAR(128) NOT NULL DEFAULT 'default', url VARCHAR(2048) NOT NULL, tls_mode VARCHAR(32) NOT NULL,
 ca_credential_id BIGINT UNSIGNED NULL, config_json LONGTEXT NULL, enabled BOOLEAN NOT NULL DEFAULT TRUE,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 UNIQUE KEY uq_cluster_endpoint(cluster_id,kind,name), INDEX idx_endpoint_enabled(cluster_id,enabled), INDEX idx_endpoint_ca(ca_credential_id),
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE,
 FOREIGN KEY(ca_credential_id) REFERENCES ceph_cluster_credential(id) ON DELETE SET NULL
) ENGINE=InnoDB;
CREATE TABLE ceph_cluster_capability (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, name VARCHAR(128) NOT NULL,
 supported BOOLEAN NOT NULL, reason VARCHAR(64) NULL, version VARCHAR(128) NULL, details_json LONGTEXT NULL,
 observed_at DATETIME(6) NOT NULL, updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 UNIQUE KEY uq_cluster_capability(cluster_id,name), INDEX idx_capability_supported(cluster_id,supported), INDEX idx_capability_observed(observed_at),
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE ceph_host (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, hostname VARCHAR(512) NOT NULL,
 ssh_address VARCHAR(255) NOT NULL, ssh_port SMALLINT UNSIGNED NOT NULL DEFAULT 22, ssh_user VARCHAR(128) NOT NULL,
 ssh_password_secret LONGTEXT NULL, address VARCHAR(255) NULL, status VARCHAR(64) NULL,
 configured_data LONGTEXT NULL,
 discovered_data LONGTEXT NOT NULL, generation BIGINT UNSIGNED NOT NULL DEFAULT 0,
 resource_version BIGINT UNSIGNED NOT NULL DEFAULT 1, source VARCHAR(32) NOT NULL DEFAULT '', source_version VARCHAR(128) NULL,
 observed_at DATETIME(6) NULL, stale_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 UNIQUE KEY uq_ceph_host(cluster_id,hostname), INDEX idx_ceph_host_cluster(cluster_id,hostname), INDEX idx_ceph_host_observed(observed_at), INDEX idx_ceph_host_stale(stale_at), INDEX idx_ceph_host_status(status),
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE ceph_collection_run (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, module VARCHAR(64) NOT NULL,
 generation BIGINT UNSIGNED NOT NULL, status VARCHAR(32) NOT NULL, source VARCHAR(32) NOT NULL,
 record_count BIGINT UNSIGNED NOT NULL DEFAULT 0, error_code VARCHAR(64) NULL, error_message LONGTEXT NULL,
 started_at DATETIME(6) NOT NULL, finished_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 UNIQUE KEY uq_collection_run(cluster_id,module,generation), INDEX idx_collection_module_started(cluster_id,module,started_at),
 INDEX idx_collection_status_started(status,started_at), INDEX idx_collection_finished(finished_at),
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE audit_event (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, occurred_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 event_type VARCHAR(32) NOT NULL, request_id VARCHAR(64) NOT NULL, actor_user_id BIGINT UNSIGNED NULL,
 actor_username VARCHAR(128) NOT NULL, cluster_id BIGINT UNSIGNED NULL, cluster_name VARCHAR(128) NULL,
 action VARCHAR(128) NOT NULL, resource_kind VARCHAR(64) NULL, resource_key VARCHAR(512) NULL,
 risk VARCHAR(16) NULL, outcome VARCHAR(32) NOT NULL, http_status INT NULL, error_code VARCHAR(64) NULL,
 source_ip VARCHAR(64) NULL, user_agent VARCHAR(512) NULL,
 before_generation BIGINT UNSIGNED NULL, after_generation BIGINT UNSIGNED NULL, parameters_json LONGTEXT NULL,
 details_json LONGTEXT NULL, previous_hash VARCHAR(64) NULL, event_hash VARCHAR(64) NOT NULL UNIQUE,
 INDEX idx_audit_occurred(occurred_at), INDEX idx_audit_actor_occurred(actor_user_id,occurred_at),
 INDEX idx_audit_cluster_occurred(cluster_id,occurred_at), INDEX idx_audit_action_occurred(action,occurred_at),
 INDEX idx_audit_resource_occurred(resource_kind,resource_key,occurred_at), INDEX idx_audit_request(request_id),
 FOREIGN KEY(actor_user_id) REFERENCES user(id) ON DELETE SET NULL,
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE SET NULL
) ENGINE=InnoDB;
