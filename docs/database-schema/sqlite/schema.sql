-- CephTower current database schema for SQLite.
-- Mirrors backend/internal/store/models.go.

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS `setting` (
  `key` text NOT NULL,
  `value` text NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  PRIMARY KEY (`key`)
);

CREATE TABLE IF NOT EXISTS `ceph_cluster` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `monitor_host` text NOT NULL,
  `keyring` text NOT NULL,
  `dashboard_username` text NOT NULL,
  `dashboard_password` text NOT NULL,
  `created_at` datetime,
  `updated_at` datetime
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_name`
  ON `ceph_cluster` (`name`);

CREATE TABLE IF NOT EXISTS `ceph_resource_snapshot` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `category` text NOT NULL,
  `resource_key` text NOT NULL,
  `payload` longtext NOT NULL,
  `last_synced_at` datetime NOT NULL,
  `last_error` text NOT NULL DEFAULT '',
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_resource_snapshot_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_resource_snapshot`
  ON `ceph_resource_snapshot` (`cluster_id`, `category`, `resource_key`);

CREATE INDEX IF NOT EXISTS `idx_ceph_resource_snapshot_last_synced_at`
  ON `ceph_resource_snapshot` (`last_synced_at`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_host` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `hostname` text NOT NULL,
  `addr` text,
  `ceph_version` text,
  `status` text,
  `labels` longtext NOT NULL,
  `sources` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_host_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_host`
  ON `ceph_cluster_host` (`cluster_id`, `hostname`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_osd` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `osd_id` text NOT NULL,
  `hostname` text,
  `status` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_osd_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_osd`
  ON `ceph_cluster_osd` (`cluster_id`, `osd_id`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_osd_flag` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `name` text NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_osd_flag_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_osd_flag`
  ON `ceph_cluster_osd_flag` (`cluster_id`, `name`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_daemon` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `name` text NOT NULL,
  `daemon_type` text,
  `hostname` text,
  `status` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_daemon_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_daemon`
  ON `ceph_cluster_daemon` (`cluster_id`, `name`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_service` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `service_name` text NOT NULL,
  `service_type` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_service_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_service`
  ON `ceph_cluster_service` (`cluster_id`, `service_name`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_mon` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `name` text NOT NULL,
  `rank` text,
  `addr` text,
  `public_addr` text,
  `status` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_mon_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_mon`
  ON `ceph_cluster_mon` (`cluster_id`, `name`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_mgr` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `name` text NOT NULL,
  `addr` text,
  `hostname` text,
  `status` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_mgr_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_mgr`
  ON `ceph_cluster_mgr` (`cluster_id`, `name`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_mds` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `name` text NOT NULL,
  `filesystem` text,
  `rank` text,
  `gid` text,
  `addr` text,
  `hostname` text,
  `state` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_mds_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_mds`
  ON `ceph_cluster_mds` (`cluster_id`, `name`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_mgr_module` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `name` text NOT NULL,
  `enabled` numeric NOT NULL DEFAULT false,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_mgr_module_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_mgr_module`
  ON `ceph_cluster_mgr_module` (`cluster_id`, `name`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_configuration` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `who` text,
  `name` text NOT NULL,
  `level` text,
  `value` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_configuration_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_configuration`
  ON `ceph_cluster_configuration` (`cluster_id`, `who`, `name`);

CREATE TABLE IF NOT EXISTS `user` (
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

CREATE UNIQUE INDEX IF NOT EXISTS `idx_user_username`
  ON `user` (`username`);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_user_email`
  ON `user` (`email`);

CREATE INDEX IF NOT EXISTS `idx_user_role`
  ON `user` (`role`);

CREATE TABLE IF NOT EXISTS `password_reset_code` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `user_id` integer NOT NULL,
  `code_hash` text NOT NULL,
  `used` numeric NOT NULL DEFAULT false,
  `expires_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_password_reset_code_user`
    FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS `idx_password_reset_code_user_id`
  ON `password_reset_code` (`user_id`);

CREATE INDEX IF NOT EXISTS `idx_password_reset_code_expires_at`
  ON `password_reset_code` (`expires_at`);

CREATE TABLE IF NOT EXISTS `user_session` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `token` text NOT NULL,
  `user_id` integer NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_user_session_user`
    FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_user_session_token`
  ON `user_session` (`token`);

CREATE INDEX IF NOT EXISTS `idx_user_session_user_id`
  ON `user_session` (`user_id`);

CREATE INDEX IF NOT EXISTS `idx_user_session_expires_at`
  ON `user_session` (`expires_at`);
