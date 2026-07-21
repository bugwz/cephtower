-- CephTower current database schema for MySQL.
-- Mirrors backend/internal/store/models.go.

CREATE TABLE IF NOT EXISTS `settings` (
  `key` varchar(128) NOT NULL,
  `value` text NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_clusters` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `description` text NOT NULL,
  `fsid` varchar(64) DEFAULT NULL,
  `enabled` boolean NOT NULL DEFAULT true,
  `dashboard_enabled` boolean NOT NULL DEFAULT false,
  `dashboard_base_url` varchar(512) DEFAULT NULL,
  `dashboard_username` varchar(128) DEFAULT NULL,
  `dashboard_password` text,
  `dashboard_insecure_tls` boolean NOT NULL DEFAULT false,
  `command_enabled` boolean NOT NULL DEFAULT false,
  `command_bin` varchar(256) NOT NULL DEFAULT 'ceph',
  `command_cluster` varchar(128) DEFAULT NULL,
  `command_conf` varchar(512) DEFAULT NULL,
  `command_name` varchar(128) DEFAULT NULL,
  `command_keyring` varchar(512) DEFAULT NULL,
  `command_keyring_content` text,
  `command_timeout_seconds` bigint NOT NULL DEFAULT 15,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_clusters_name` (`name`),
  KEY `idx_ceph_clusters_fsid` (`fsid`),
  KEY `idx_ceph_clusters_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_resource_snapshots` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `category` varchar(64) NOT NULL,
  `resource_key` varchar(256) NOT NULL,
  `payload` longtext NOT NULL,
  `last_synced_at` datetime(3) NOT NULL,
  `last_error` text NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_resource_snapshot` (`cluster_id`,`category`,`resource_key`),
  KEY `idx_ceph_resource_snapshots_last_synced_at` (`last_synced_at`),
  CONSTRAINT `fk_ceph_resource_snapshots_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_clusters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL,
  `display_name` varchar(96) NOT NULL,
  `email` varchar(128) DEFAULT NULL,
  `role` varchar(24) NOT NULL,
  `permissions` text NOT NULL,
  `password_hash` text NOT NULL,
  `enabled` boolean NOT NULL DEFAULT true,
  `last_login_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  UNIQUE KEY `idx_users_email` (`email`),
  KEY `idx_users_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `password_reset_codes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `code_hash` text NOT NULL,
  `used` boolean NOT NULL DEFAULT false,
  `expires_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_password_reset_codes_user_id` (`user_id`),
  KEY `idx_password_reset_codes_expires_at` (`expires_at`),
  CONSTRAINT `fk_password_reset_codes_user`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `token` varchar(96) NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `expires_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_sessions_token` (`token`),
  KEY `idx_user_sessions_user_id` (`user_id`),
  KEY `idx_user_sessions_expires_at` (`expires_at`),
  CONSTRAINT `fk_user_sessions_user`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
