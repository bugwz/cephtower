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

CREATE TABLE IF NOT EXISTS `ceph_data_fetch_run` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `module` text NOT NULL,
  `status` text NOT NULL,
  `source` text NOT NULL,
  `started_at` datetime NOT NULL,
  `finished_at` datetime,
  `duration_ms` integer,
  `records_upserted` integer NOT NULL DEFAULT 0,
  `records_deleted` integer NOT NULL DEFAULT 0,
  `error` text NOT NULL DEFAULT '',
  `created_at` datetime,
  CONSTRAINT `fk_ceph_data_fetch_run_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS `idx_ceph_data_fetch_run_cluster_id`
  ON `ceph_data_fetch_run` (`cluster_id`);

CREATE INDEX IF NOT EXISTS `idx_ceph_data_fetch_run_module`
  ON `ceph_data_fetch_run` (`module`);

CREATE INDEX IF NOT EXISTS `idx_ceph_data_fetch_run_started_at`
  ON `ceph_data_fetch_run` (`started_at`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_summary` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `health_status` text,
  `version` text,
  `mgr_id` text,
  `mgr_host` text,
  `have_mon_connection` numeric NOT NULL DEFAULT false,
  `executing_tasks` longtext NOT NULL,
  `finished_tasks` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_summary_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_summary_cluster_id`
  ON `ceph_cluster_summary` (`cluster_id`);

CREATE TABLE IF NOT EXISTS `ceph_cluster_health_check` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `name` text NOT NULL,
  `severity` text,
  `summary` text,
  `detail` longtext NOT NULL,
  `muted` numeric NOT NULL DEFAULT false,
  `count` integer,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_cluster_health_check_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_cluster_health_check`
  ON `ceph_cluster_health_check` (`cluster_id`, `name`);

CREATE TABLE IF NOT EXISTS `ceph_pool` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `pool_id` text,
  `pool_name` text NOT NULL,
  `type` text,
  `size` integer,
  `min_size` integer,
  `pg_num` integer,
  `pg_placement_num` integer,
  `pg_autoscale_mode` text,
  `crush_rule` text,
  `erasure_code_profile` text,
  `application_metadata` longtext NOT NULL,
  `quota_max_bytes` integer,
  `quota_max_objects` integer,
  `used_bytes` integer,
  `max_avail_bytes` integer,
  `objects` integer,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_pool_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_pool`
  ON `ceph_pool` (`cluster_id`, `pool_name`);

CREATE TABLE IF NOT EXISTS `ceph_rbd_image` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `pool_name` text NOT NULL,
  `namespace` text,
  `image_name` text NOT NULL,
  `image_spec` text NOT NULL,
  `image_id` text,
  `size_bytes` integer,
  `object_size` integer,
  `features` longtext NOT NULL,
  `stripe_count` integer,
  `stripe_unit` integer,
  `parent` longtext NOT NULL,
  `snapshots` longtext NOT NULL,
  `mirror_mode` text,
  `trash` numeric NOT NULL DEFAULT false,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_rbd_image_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_rbd_image`
  ON `ceph_rbd_image` (`cluster_id`, `image_spec`);

CREATE TABLE IF NOT EXISTS `ceph_filesystem` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `fs_id` text NOT NULL,
  `name` text NOT NULL,
  `metadata_pool` text,
  `data_pools` longtext NOT NULL,
  `mds_map` longtext NOT NULL,
  `standby_count` integer,
  `client_count` integer,
  `used_bytes` integer,
  `avail_bytes` integer,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_filesystem_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_filesystem`
  ON `ceph_filesystem` (`cluster_id`, `fs_id`);

CREATE TABLE IF NOT EXISTS `ceph_rgw_daemon` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `service_id` text NOT NULL,
  `hostname` text,
  `zone_name` text,
  `frontend_config` text,
  `version` text,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_rgw_daemon_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_rgw_daemon`
  ON `ceph_rgw_daemon` (`cluster_id`, `service_id`);

CREATE TABLE IF NOT EXISTS `ceph_rgw_user` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `uid` text NOT NULL,
  `display_name` text,
  `email` text,
  `suspended` numeric NOT NULL DEFAULT false,
  `max_buckets` integer,
  `subusers` longtext NOT NULL,
  `keys_redacted` longtext NOT NULL,
  `caps` longtext NOT NULL,
  `quota` longtext NOT NULL,
  `stats` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_rgw_user_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_rgw_user`
  ON `ceph_rgw_user` (`cluster_id`, `uid`);

CREATE TABLE IF NOT EXISTS `ceph_rgw_bucket` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `tenant` text,
  `bucket` text NOT NULL,
  `owner` text,
  `zonegroup` text,
  `placement_rule` text,
  `versioning` text,
  `object_count` integer,
  `used_bytes` integer,
  `quota` longtext NOT NULL,
  `lifecycle` longtext NOT NULL,
  `encryption` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_rgw_bucket_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_rgw_bucket`
  ON `ceph_rgw_bucket` (`cluster_id`, `tenant`, `bucket`);

CREATE TABLE IF NOT EXISTS `ceph_nvmeof_gateway` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `group_name` text NOT NULL,
  `hostname` text NOT NULL,
  `tr_addr` text NOT NULL,
  `status` text,
  `version` text,
  `listeners` longtext NOT NULL,
  `stats` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_nvmeof_gateway_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_nvmeof_gateway`
  ON `ceph_nvmeof_gateway` (`cluster_id`, `group_name`, `hostname`, `tr_addr`);

CREATE TABLE IF NOT EXISTS `ceph_nvmeof_subsystem` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `nqn` text NOT NULL,
  `serial_number` text,
  `model_number` text,
  `max_namespaces` integer,
  `namespaces` longtext NOT NULL,
  `hosts` longtext NOT NULL,
  `listeners` longtext NOT NULL,
  `connections` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_nvmeof_subsystem_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_nvmeof_subsystem`
  ON `ceph_nvmeof_subsystem` (`cluster_id`, `nqn`);

CREATE TABLE IF NOT EXISTS `ceph_iscsi_target` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `target_iqn` text NOT NULL,
  `portals` longtext NOT NULL,
  `disks` longtext NOT NULL,
  `clients` longtext NOT NULL,
  `groups` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_iscsi_target_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_iscsi_target`
  ON `ceph_iscsi_target` (`cluster_id`, `target_iqn`);

CREATE TABLE IF NOT EXISTS `ceph_nfs_export` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `cluster_id` integer NOT NULL,
  `nfs_cluster_id` text NOT NULL,
  `export_id` text NOT NULL,
  `path` text,
  `pseudo` text,
  `access_type` text,
  `squash` text,
  `protocols` longtext NOT NULL,
  `transports` longtext NOT NULL,
  `fsal` longtext NOT NULL,
  `payload` longtext NOT NULL,
  `discovered_at` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime,
  CONSTRAINT `fk_ceph_nfs_export_cluster`
    FOREIGN KEY (`cluster_id`) REFERENCES `ceph_cluster` (`id`) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS `idx_ceph_nfs_export`
  ON `ceph_nfs_export` (`cluster_id`, `nfs_cluster_id`, `export_id`);

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
