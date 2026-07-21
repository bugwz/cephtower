-- CephTower current database schema for SQLite.
-- Mirrors backend/internal/store/models.go.

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS `settings` (
  `key` text NOT NULL,
  `value` text NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  PRIMARY KEY (`key`)
);

CREATE TABLE IF NOT EXISTS `ceph_clusters` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `description` text NOT NULL DEFAULT '',
  `fsid` text,
  `enabled` numeric NOT NULL DEFAULT true,
  `dashboard_enabled` numeric NOT NULL DEFAULT false,
  `dashboard_base_url` text,
  `dashboard_username` text,
  `dashboard_password` text,
  `dashboard_insecure_tls` numeric NOT NULL DEFAULT false,
  `command_enabled` numeric NOT NULL DEFAULT false,
  `command_bin` text NOT NULL DEFAULT 'ceph',
  `command_cluster` text,
  `command_conf` text,
  `command_name` text,
  `command_keyring` text,
  `command_keyring_content` text,
  `command_timeout_seconds` integer NOT NULL DEFAULT 15,
  `created_at` datetime,
  `updated_at` datetime
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_clusters_name`
  ON `ceph_clusters` (`name`);

CREATE INDEX IF NOT EXISTS `idx_ceph_clusters_fsid`
  ON `ceph_clusters` (`fsid`);

CREATE INDEX IF NOT EXISTS `idx_ceph_clusters_enabled`
  ON `ceph_clusters` (`enabled`);

CREATE TABLE IF NOT EXISTS `ceph_resource_snapshots` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `category` text NOT NULL,
  `resource_key` text NOT NULL,
  `payload` longtext NOT NULL,
  `last_synced_at` datetime NOT NULL,
  `last_error` text NOT NULL DEFAULT '',
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_resource_snapshots_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_clusters` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_resource_snapshot`
  ON `ceph_resource_snapshots` (`cluster_id`, `category`, `resource_key`);

CREATE INDEX IF NOT EXISTS `idx_ceph_resource_snapshots_last_synced_at`
  ON `ceph_resource_snapshots` (`last_synced_at`);

CREATE TABLE IF NOT EXISTS `users` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `username` text NOT NULL,
  `display_name` text NOT NULL,
  `email` text,
  `role` text NOT NULL,
  `permissions` text NOT NULL,
  `password_hash` text NOT NULL,
  `enabled` numeric NOT NULL DEFAULT true,
  `last_login_at` datetime,
  `created_at` datetime,
  `updated_at` datetime
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_users_username`
  ON `users` (`username`);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_users_email`
  ON `users` (`email`);

CREATE INDEX IF NOT EXISTS `idx_users_role`
  ON `users` (`role`);

CREATE TABLE IF NOT EXISTS `password_reset_codes` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `user_id` integer NOT NULL,
  `code_hash` text NOT NULL,
  `used` numeric NOT NULL DEFAULT false,
  `expires_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_password_reset_codes_user`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS `idx_password_reset_codes_user_id`
  ON `password_reset_codes` (`user_id`);

CREATE INDEX IF NOT EXISTS `idx_password_reset_codes_expires_at`
  ON `password_reset_codes` (`expires_at`);

CREATE TABLE IF NOT EXISTS `user_sessions` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `token` text NOT NULL,
  `user_id` integer NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_user_sessions_user`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_user_sessions_token`
  ON `user_sessions` (`token`);

CREATE INDEX IF NOT EXISTS `idx_user_sessions_user_id`
  ON `user_sessions` (`user_id`);

CREATE INDEX IF NOT EXISTS `idx_user_sessions_expires_at`
  ON `user_sessions` (`expires_at`);
