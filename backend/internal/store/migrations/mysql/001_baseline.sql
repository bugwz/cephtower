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
CREATE TABLE ceph_cluster_observation (
 cluster_id BIGINT UNSIGNED PRIMARY KEY, fsid VARCHAR(36) NULL UNIQUE, ceph_version VARCHAR(128) NULL,
 status VARCHAR(32) NOT NULL DEFAULT 'unknown', enabled BOOLEAN NOT NULL DEFAULT TRUE, generation BIGINT UNSIGNED NOT NULL DEFAULT 0,
 last_seen_at DATETIME(6) NULL, last_error_code VARCHAR(64) NULL, last_error_message LONGTEXT NULL,
 observed_at DATETIME(6) NULL, updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 INDEX idx_observation_status_enabled(status,enabled), INDEX idx_observation_last_seen(last_seen_at),
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE
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
CREATE TABLE ceph_resource_record (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, kind VARCHAR(64) NOT NULL,
 natural_key VARCHAR(512) NOT NULL, parent_kind VARCHAR(64) NULL, parent_key VARCHAR(512) NULL,
 name VARCHAR(512) NULL, status VARCHAR(64) NULL, generation BIGINT UNSIGNED NOT NULL,
 resource_version BIGINT UNSIGNED NOT NULL DEFAULT 1, source VARCHAR(32) NOT NULL, source_version VARCHAR(128) NULL,
 observed_at DATETIME(6) NOT NULL, stale_at DATETIME(6) NULL, payload_schema_version INT NOT NULL,
 payload_json LONGTEXT NOT NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), UNIQUE KEY uq_resource(cluster_id,kind,natural_key),
 INDEX idx_resource_name(cluster_id,kind,name), INDEX idx_resource_status(cluster_id,kind,status),
 INDEX idx_resource_generation(cluster_id,kind,generation), INDEX idx_resource_observed(cluster_id,kind,observed_at),
 INDEX idx_resource_parent(cluster_id,parent_kind,parent_key), INDEX idx_resource_stale(cluster_id,stale_at),
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
CREATE TABLE ceph_action_plan (
 id VARCHAR(36) PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, actor_user_id BIGINT UNSIGNED NULL,
 actor_username VARCHAR(128) NOT NULL, request_id VARCHAR(64) NOT NULL, action VARCHAR(128) NOT NULL,
 resource_kind VARCHAR(64) NOT NULL, resource_key VARCHAR(512) NOT NULL, resource_generation BIGINT UNSIGNED NOT NULL,
 risk VARCHAR(16) NOT NULL, status VARCHAR(32) NOT NULL, request_json LONGTEXT NOT NULL,
 blockers_json LONGTEXT NOT NULL, warnings_json LONGTEXT NOT NULL, expires_at DATETIME(6) NOT NULL,
 consumed_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 INDEX idx_plan_resource(cluster_id,resource_kind,resource_key), INDEX idx_plan_actor_created(actor_user_id,created_at),
 INDEX idx_plan_status_expiry(status,expires_at), INDEX idx_plan_request(request_id),
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE, FOREIGN KEY(actor_user_id) REFERENCES user(id) ON DELETE SET NULL
) ENGINE=InnoDB;
CREATE TABLE ceph_operation (
 id VARCHAR(36) PRIMARY KEY, cluster_id BIGINT UNSIGNED NULL, cluster_name VARCHAR(128) NOT NULL,
 actor_user_id BIGINT UNSIGNED NULL, actor_username VARCHAR(128) NOT NULL, plan_id VARCHAR(36) NULL,
 retry_of_id VARCHAR(36) NULL, request_id VARCHAR(64) NOT NULL, action VARCHAR(128) NOT NULL,
 resource_kind VARCHAR(64) NOT NULL, resource_key VARCHAR(512) NOT NULL, resource_generation BIGINT UNSIGNED NULL,
 risk VARCHAR(16) NOT NULL, status VARCHAR(32) NOT NULL DEFAULT 'queued', stage VARCHAR(64) NOT NULL DEFAULT 'queued',
 progress INT NOT NULL DEFAULT 0 CHECK(progress BETWEEN 0 AND 100), attempt INT NOT NULL DEFAULT 0,
 max_attempts INT NOT NULL DEFAULT 1, idempotency_key_hash VARCHAR(64) NULL,
 idempotency_scope_hash VARCHAR(64) NULL UNIQUE, request_json LONGTEXT NOT NULL, result_json LONGTEXT NULL,
 error_code VARCHAR(64) NULL, error_message LONGTEXT NULL, error_details_json LONGTEXT NULL,
 retryable BOOLEAN NOT NULL DEFAULT FALSE, cancel_requested_at DATETIME(6) NULL, scheduled_at DATETIME(6) NOT NULL,
 started_at DATETIME(6) NULL, heartbeat_at DATETIME(6) NULL, completed_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 INDEX idx_operation_cluster_status_created(cluster_id,status,created_at), INDEX idx_operation_actor_created(actor_user_id,created_at),
 INDEX idx_operation_resource_created(resource_kind,resource_key,created_at), INDEX idx_operation_status_scheduled(status,scheduled_at),
 INDEX idx_operation_heartbeat(heartbeat_at), INDEX idx_operation_plan(plan_id), INDEX idx_operation_retry_of(retry_of_id),
 INDEX idx_operation_request(request_id), FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE SET NULL,
 FOREIGN KEY(actor_user_id) REFERENCES user(id) ON DELETE SET NULL, FOREIGN KEY(plan_id) REFERENCES ceph_action_plan(id) ON DELETE SET NULL,
 FOREIGN KEY(retry_of_id) REFERENCES ceph_operation(id) ON DELETE SET NULL
) ENGINE=InnoDB;
CREATE TABLE ceph_operation_event (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, operation_id VARCHAR(36) NOT NULL, sequence BIGINT UNSIGNED NOT NULL,
 event_type VARCHAR(32) NOT NULL, stage VARCHAR(64) NOT NULL, progress INT NULL,
 message LONGTEXT NOT NULL, data_json LONGTEXT NULL, error_code VARCHAR(64) NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), UNIQUE KEY uq_operation_event(operation_id,sequence),
 INDEX idx_event_operation_created(operation_id,created_at), INDEX idx_event_created(created_at),
 FOREIGN KEY(operation_id) REFERENCES ceph_operation(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE ceph_operation_lock (
 lock_key VARCHAR(191) PRIMARY KEY, cluster_id BIGINT UNSIGNED NOT NULL, resource_kind VARCHAR(64) NOT NULL,
 resource_key VARCHAR(512) NOT NULL, operation_id VARCHAR(36) NOT NULL, fencing_token BIGINT UNSIGNED NOT NULL,
 lease_expires_at DATETIME(6) NOT NULL, acquired_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), INDEX idx_lock_resource(cluster_id,resource_kind,resource_key),
 INDEX idx_lock_operation(operation_id), INDEX idx_lock_expiry(lease_expires_at),
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE, FOREIGN KEY(operation_id) REFERENCES ceph_operation(id) ON DELETE CASCADE
) ENGINE=InnoDB;
CREATE TABLE audit_event (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY, occurred_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 event_type VARCHAR(32) NOT NULL, request_id VARCHAR(64) NOT NULL, actor_user_id BIGINT UNSIGNED NULL,
 actor_username VARCHAR(128) NOT NULL, cluster_id BIGINT UNSIGNED NULL, cluster_name VARCHAR(128) NULL,
 action VARCHAR(128) NOT NULL, resource_kind VARCHAR(64) NULL, resource_key VARCHAR(512) NULL,
 risk VARCHAR(16) NULL, outcome VARCHAR(32) NOT NULL, http_status INT NULL, error_code VARCHAR(64) NULL,
 source_ip VARCHAR(64) NULL, user_agent VARCHAR(512) NULL, plan_id VARCHAR(36) NULL, operation_id VARCHAR(36) NULL,
 before_generation BIGINT UNSIGNED NULL, after_generation BIGINT UNSIGNED NULL, parameters_json LONGTEXT NULL,
 details_json LONGTEXT NULL, previous_hash VARCHAR(64) NULL, event_hash VARCHAR(64) NOT NULL UNIQUE,
 INDEX idx_audit_occurred(occurred_at), INDEX idx_audit_actor_occurred(actor_user_id,occurred_at),
 INDEX idx_audit_cluster_occurred(cluster_id,occurred_at), INDEX idx_audit_action_occurred(action,occurred_at),
 INDEX idx_audit_resource_occurred(resource_kind,resource_key,occurred_at), INDEX idx_audit_request(request_id),
 INDEX idx_audit_operation(operation_id), FOREIGN KEY(actor_user_id) REFERENCES user(id) ON DELETE SET NULL,
 FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE SET NULL
) ENGINE=InnoDB;
