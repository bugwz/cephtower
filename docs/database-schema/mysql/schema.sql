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

CREATE TABLE IF NOT EXISTS `ceph_data_fetch_run` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `module` varchar(64) NOT NULL,
  `status` varchar(32) NOT NULL,
  `source` varchar(32) NOT NULL,
  `started_at` datetime(3) NOT NULL,
  `finished_at` datetime(3) DEFAULT NULL,
  `duration_ms` bigint DEFAULT NULL,
  `records_upserted` bigint NOT NULL DEFAULT 0,
  `records_deleted` bigint NOT NULL DEFAULT 0,
  `error` text NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ceph_data_fetch_run_cluster_id` (`cluster_id`),
  KEY `idx_ceph_data_fetch_run_module` (`module`),
  KEY `idx_ceph_data_fetch_run_status` (`status`),
  KEY `idx_ceph_data_fetch_run_started_at` (`started_at`),
  CONSTRAINT `fk_ceph_data_fetch_run_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_summary` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `health_status` varchar(64) DEFAULT NULL,
  `version` varchar(128) DEFAULT NULL,
  `mgr_id` varchar(128) DEFAULT NULL,
  `mgr_host` varchar(256) DEFAULT NULL,
  `have_mon_connection` boolean NOT NULL DEFAULT false,
  `executing_tasks` longtext NOT NULL,
  `finished_tasks` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_summary_cluster_id` (`cluster_id`),
  KEY `idx_ceph_cluster_summary_health_status` (`health_status`),
  CONSTRAINT `fk_ceph_cluster_summary_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_cluster_health_check` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `name` varchar(256) NOT NULL,
  `severity` varchar(64) DEFAULT NULL,
  `summary` text,
  `detail` longtext NOT NULL,
  `muted` boolean NOT NULL DEFAULT false,
  `count` bigint DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_cluster_health_check` (`cluster_id`,`name`),
  KEY `idx_ceph_cluster_health_check_severity` (`severity`),
  CONSTRAINT `fk_ceph_cluster_health_check_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_pool` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `pool_id` varchar(64) DEFAULT NULL,
  `pool_name` varchar(256) NOT NULL,
  `type` varchar(64) DEFAULT NULL,
  `size` bigint DEFAULT NULL,
  `min_size` bigint DEFAULT NULL,
  `pg_num` bigint DEFAULT NULL,
  `pg_placement_num` bigint DEFAULT NULL,
  `pg_autoscale_mode` varchar(64) DEFAULT NULL,
  `crush_rule` varchar(128) DEFAULT NULL,
  `erasure_code_profile` varchar(128) DEFAULT NULL,
  `application_metadata` longtext NOT NULL,
  `quota_max_bytes` bigint DEFAULT NULL,
  `quota_max_objects` bigint DEFAULT NULL,
  `used_bytes` bigint DEFAULT NULL,
  `max_avail_bytes` bigint DEFAULT NULL,
  `objects` bigint DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_pool` (`cluster_id`,`pool_name`),
  KEY `idx_ceph_pool_pool_id` (`pool_id`),
  KEY `idx_ceph_pool_type` (`type`),
  CONSTRAINT `fk_ceph_pool_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_rbd_image` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `pool_name` varchar(256) NOT NULL,
  `namespace` varchar(256) DEFAULT NULL,
  `image_name` varchar(256) NOT NULL,
  `image_spec` varchar(768) NOT NULL,
  `image_id` varchar(128) DEFAULT NULL,
  `size_bytes` bigint DEFAULT NULL,
  `object_size` bigint DEFAULT NULL,
  `features` longtext NOT NULL,
  `stripe_count` bigint DEFAULT NULL,
  `stripe_unit` bigint DEFAULT NULL,
  `parent` longtext NOT NULL,
  `snapshots` longtext NOT NULL,
  `mirror_mode` varchar(64) DEFAULT NULL,
  `trash` boolean NOT NULL DEFAULT false,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_rbd_image` (`cluster_id`,`image_spec`),
  KEY `idx_ceph_rbd_image_pool_name` (`pool_name`),
  KEY `idx_ceph_rbd_image_image_id` (`image_id`),
  KEY `idx_ceph_rbd_image_trash` (`trash`),
  CONSTRAINT `fk_ceph_rbd_image_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_filesystem` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `fs_id` varchar(64) NOT NULL,
  `name` varchar(256) NOT NULL,
  `metadata_pool` varchar(256) DEFAULT NULL,
  `data_pools` longtext NOT NULL,
  `mds_map` longtext NOT NULL,
  `standby_count` bigint DEFAULT NULL,
  `client_count` bigint DEFAULT NULL,
  `used_bytes` bigint DEFAULT NULL,
  `avail_bytes` bigint DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_filesystem` (`cluster_id`,`fs_id`),
  KEY `idx_ceph_filesystem_name` (`name`),
  CONSTRAINT `fk_ceph_filesystem_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_rgw_daemon` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `service_id` varchar(256) NOT NULL,
  `hostname` varchar(256) DEFAULT NULL,
  `zone_name` varchar(256) DEFAULT NULL,
  `frontend_config` text,
  `version` varchar(128) DEFAULT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_rgw_daemon` (`cluster_id`,`service_id`),
  KEY `idx_ceph_rgw_daemon_hostname` (`hostname`),
  KEY `idx_ceph_rgw_daemon_zone_name` (`zone_name`),
  CONSTRAINT `fk_ceph_rgw_daemon_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_rgw_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `uid` varchar(256) NOT NULL,
  `display_name` varchar(256) DEFAULT NULL,
  `email` varchar(256) DEFAULT NULL,
  `suspended` boolean NOT NULL DEFAULT false,
  `max_buckets` bigint DEFAULT NULL,
  `subusers` longtext NOT NULL,
  `keys_redacted` longtext NOT NULL,
  `caps` longtext NOT NULL,
  `quota` longtext NOT NULL,
  `stats` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_rgw_user` (`cluster_id`,`uid`),
  KEY `idx_ceph_rgw_user_email` (`email`),
  KEY `idx_ceph_rgw_user_suspended` (`suspended`),
  CONSTRAINT `fk_ceph_rgw_user_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_rgw_bucket` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `tenant` varchar(256) NOT NULL DEFAULT '',
  `bucket` varchar(256) NOT NULL,
  `owner` varchar(256) DEFAULT NULL,
  `zonegroup` varchar(256) DEFAULT NULL,
  `placement_rule` varchar(256) DEFAULT NULL,
  `versioning` varchar(64) DEFAULT NULL,
  `object_count` bigint DEFAULT NULL,
  `used_bytes` bigint DEFAULT NULL,
  `quota` longtext NOT NULL,
  `lifecycle` longtext NOT NULL,
  `encryption` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_rgw_bucket` (`cluster_id`,`tenant`,`bucket`),
  KEY `idx_ceph_rgw_bucket_owner` (`owner`),
  CONSTRAINT `fk_ceph_rgw_bucket_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_nvmeof_gateway` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `group_name` varchar(256) NOT NULL,
  `hostname` varchar(256) NOT NULL,
  `tr_addr` varchar(256) NOT NULL,
  `status` varchar(64) DEFAULT NULL,
  `version` varchar(128) DEFAULT NULL,
  `listeners` longtext NOT NULL,
  `stats` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_nvmeof_gateway` (`cluster_id`,`group_name`,`hostname`,`tr_addr`),
  KEY `idx_ceph_nvmeof_gateway_status` (`status`),
  CONSTRAINT `fk_ceph_nvmeof_gateway_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_nvmeof_subsystem` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `nqn` varchar(512) NOT NULL,
  `serial_number` varchar(128) DEFAULT NULL,
  `model_number` varchar(128) DEFAULT NULL,
  `max_namespaces` bigint DEFAULT NULL,
  `namespaces` longtext NOT NULL,
  `hosts` longtext NOT NULL,
  `listeners` longtext NOT NULL,
  `connections` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_nvmeof_subsystem` (`cluster_id`,`nqn`),
  CONSTRAINT `fk_ceph_nvmeof_subsystem_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_iscsi_target` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `target_iqn` varchar(512) NOT NULL,
  `portals` longtext NOT NULL,
  `disks` longtext NOT NULL,
  `clients` longtext NOT NULL,
  `groups` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_iscsi_target` (`cluster_id`,`target_iqn`),
  CONSTRAINT `fk_ceph_iscsi_target_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `ceph_nfs_export` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cluster_id` bigint unsigned NOT NULL,
  `nfs_cluster_id` varchar(256) NOT NULL,
  `export_id` varchar(128) NOT NULL,
  `path` varchar(1024) DEFAULT NULL,
  `pseudo` varchar(1024) DEFAULT NULL,
  `access_type` varchar(64) DEFAULT NULL,
  `squash` varchar(128) DEFAULT NULL,
  `protocols` longtext NOT NULL,
  `transports` longtext NOT NULL,
  `fsal` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime(3) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_ceph_nfs_export` (`cluster_id`,`nfs_cluster_id`,`export_id`),
  CONSTRAINT `fk_ceph_nfs_export_cluster`
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
