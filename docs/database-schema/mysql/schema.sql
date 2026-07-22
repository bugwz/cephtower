-- CephTower current database schema for MySQL.
-- Mirrors backend/internal/store/models.go.

CREATE TABLE IF NOT EXISTS `setting` (
  `key` varchar(128) NOT NULL,
  `value` text NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `monitor_host` text NOT NULL,
  `keyring` text NOT NULL,
  `dashboard_username` varchar(128) NOT NULL,
  `dashboard_password` text NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_resource_snapshot` (
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
  KEY `idx_ceph_resource_snapshot_last_synced_at` (`last_synced_at`),
  CONSTRAINT `fk_ceph_resource_snapshot_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_host` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `hostname` varchar(256) NOT NULL,
  `addr` varchar(128) DEFAULT NULL,
  `ceph_version` varchar(128) DEFAULT NULL,
  `status` varchar(64) DEFAULT NULL,
  `labels` longtext NOT NULL,
  `sources` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_host` (`cluster_id`,`hostname`),
  KEY `idx_ceph_cluster_host_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_host_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_osd` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `osd_id` varchar(64) NOT NULL,
  `hostname` varchar(256) DEFAULT NULL,
  `status` varchar(64) DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_osd` (`cluster_id`,`osd_id`),
  KEY `idx_ceph_cluster_osd_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_osd_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_osd_flag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_osd_flag` (`cluster_id`,`name`),
  KEY `idx_ceph_cluster_osd_flag_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_osd_flag_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_daemon` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(256) NOT NULL,
  `daemon_type` varchar(64) DEFAULT NULL,
  `hostname` varchar(256) DEFAULT NULL,
  `status` varchar(64) DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_daemon` (`cluster_id`,`name`),
  KEY `idx_ceph_cluster_daemon_daemon_type` (`daemon_type`),
  KEY `idx_ceph_cluster_daemon_hostname` (`hostname`),
  KEY `idx_ceph_cluster_daemon_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_daemon_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_service` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `service_name` varchar(256) NOT NULL,
  `service_type` varchar(64) DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_service` (`cluster_id`,`service_name`),
  KEY `idx_ceph_cluster_service_service_type` (`service_type`),
  KEY `idx_ceph_cluster_service_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_service_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_mon` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  `rank` varchar(64) DEFAULT NULL,
  `addr` varchar(256) DEFAULT NULL,
  `public_addr` varchar(256) DEFAULT NULL,
  `status` varchar(64) DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_mon` (`cluster_id`,`name`),
  KEY `idx_ceph_cluster_mon_status` (`status`),
  KEY `idx_ceph_cluster_mon_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_mon_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_mgr` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  `addr` varchar(256) DEFAULT NULL,
  `hostname` varchar(256) DEFAULT NULL,
  `status` varchar(64) DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_mgr` (`cluster_id`,`name`),
  KEY `idx_ceph_cluster_mgr_hostname` (`hostname`),
  KEY `idx_ceph_cluster_mgr_status` (`status`),
  KEY `idx_ceph_cluster_mgr_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_mgr_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_mds` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  `filesystem` varchar(128) DEFAULT NULL,
  `rank` varchar(64) DEFAULT NULL,
  `gid` varchar(64) DEFAULT NULL,
  `addr` varchar(256) DEFAULT NULL,
  `hostname` varchar(256) DEFAULT NULL,
  `state` varchar(64) DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_mds` (`cluster_id`,`name`),
  KEY `idx_ceph_cluster_mds_filesystem` (`filesystem`),
  KEY `idx_ceph_cluster_mds_hostname` (`hostname`),
  KEY `idx_ceph_cluster_mds_state` (`state`),
  KEY `idx_ceph_cluster_mds_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_mds_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_mgr_module` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  `enabled` boolean NOT NULL DEFAULT false,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_mgr_module` (`cluster_id`,`name`),
  KEY `idx_ceph_cluster_mgr_module_enabled` (`enabled`),
  KEY `idx_ceph_cluster_mgr_module_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_mgr_module_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_configuration` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `who` varchar(128) DEFAULT NULL,
  `name` varchar(256) NOT NULL,
  `level` varchar(64) DEFAULT NULL,
  `value` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_configuration` (`cluster_id`,`who`,`name`),
  KEY `idx_ceph_cluster_configuration_discovered_at` (`discovered_at`),
  CONSTRAINT `fk_ceph_cluster_configuration_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user` (
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
  UNIQUE KEY `idx_user_username` (`username`),
  UNIQUE KEY `idx_user_email` (`email`),
  KEY `idx_user_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `password_reset_code` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `code_hash` text NOT NULL,
  `used` boolean NOT NULL DEFAULT false,
  `expires_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_password_reset_code_user_id` (`user_id`),
  KEY `idx_password_reset_code_expires_at` (`expires_at`),
  CONSTRAINT `fk_password_reset_code_user`
    FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_session` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `token` varchar(96) NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `expires_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_session_token` (`token`),
  KEY `idx_user_session_user_id` (`user_id`),
  KEY `idx_user_session_expires_at` (`expires_at`),
  CONSTRAINT `fk_user_session_user`
    FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
