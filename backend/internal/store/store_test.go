package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cephtower/backend/internal/config"
)

func TestTransactionRollsBackDomainWrites(t *testing.T) {
	db, err := Open(config.DatabaseConfig{Engine: EngineSQLite, SQLite: config.SQLiteConfig{Path: filepath.Join(t.TempDir(), "rollback.db")}})
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	wantErr := errors.New("rollback")
	err = db.Transaction(func(tx *Database) error {
		user := User{Username: "rollback", DisplayName: "Rollback", Role: UserRoleAdmin, Permissions: "[]", PasswordHash: "hash", Enabled: true}
		if err := tx.CreateUser(context.Background(), &user); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Transaction() error = %v", err)
	}
	hasUsers, err := db.HasUsers(context.Background())
	if err != nil || hasUsers {
		t.Fatalf("HasUsers() = %v, %v; want false", hasUsers, err)
	}
}

func TestOpenSQLiteResolvesRelativePathUnderWorkDir(t *testing.T) {
	workDir := t.TempDir()
	db, err := Open(config.DatabaseConfig{
		Engine: EngineSQLite,
		SQLite: config.SQLiteConfig{Path: "data/cephtower.db"},
	}, workDir)
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	if err := Close(db); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	dbPath := filepath.Join(workDir, "data", "cephtower.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("SQLite database was not created under work directory: %v", err)
	}
}

func TestOpenSQLiteCreatesDatabaseAndMigrates(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "cephtower.db")
	db, err := Open(config.DatabaseConfig{
		Engine: EngineSQLite,
		SQLite: config.SQLiteConfig{
			Path: dbPath,
		},
	})
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	if !db.db.Migrator().HasTable(&Setting{}) {
		t.Fatal("setting table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephCluster{}) {
		t.Fatal("ceph_cluster table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterHost{}) {
		t.Fatal("ceph_cluster_host table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterOSD{}) {
		t.Fatal("ceph_cluster_osd table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterOSDFlag{}) {
		t.Fatal("ceph_cluster_osd_flag table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterDaemon{}) {
		t.Fatal("ceph_cluster_daemon table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterService{}) {
		t.Fatal("ceph_cluster_service table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterMon{}) {
		t.Fatal("ceph_cluster_mon table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterMgr{}) {
		t.Fatal("ceph_cluster_mgr table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterMDS{}) {
		t.Fatal("ceph_cluster_mds table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterMgrModule{}) {
		t.Fatal("ceph_cluster_mgr_module table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephClusterConfiguration{}) {
		t.Fatal("ceph_cluster_configuration table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephDataFetchRun{}) {
		t.Fatal("ceph_data_fetch_run table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephPool{}) {
		t.Fatal("ceph_pool table was not migrated")
	}
	if !db.db.Migrator().HasTable(&CephRBDImage{}) {
		t.Fatal("ceph_rbd_image table was not migrated")
	}
	if !db.db.Migrator().HasTable(&User{}) {
		t.Fatal("user table was not migrated")
	}
	if !db.db.Migrator().HasTable(&UserSession{}) {
		t.Fatal("user_session table was not migrated")
	}
	if !db.db.Migrator().HasTable(&PasswordResetCode{}) {
		t.Fatal("password_reset_code table was not migrated")
	}

	var users int64
	if err := db.db.Model(&User{}).Count(&users).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 0 {
		t.Fatalf("users = %d, want empty database before setup", users)
	}

	cluster := CephCluster{
		Name:              "primary",
		MonitorHost:       "10.0.0.11:6789,10.0.0.12:6789",
		Keyring:           "[client.admin]\nkey = secret",
		DashboardUsername: "admin",
		DashboardPassword: "secret",
	}
	if err := db.db.Create(&cluster).Error; err != nil {
		t.Fatalf("create ceph cluster: %v", err)
	}
	var saved CephCluster
	if err := db.db.Where("name = ?", "primary").First(&saved).Error; err != nil {
		t.Fatalf("load ceph cluster: %v", err)
	}
	if saved.MonitorHost == "" || saved.Keyring == "" || saved.DashboardUsername != "admin" || saved.DashboardPassword != "secret" {
		t.Fatalf("saved cluster = %#v, want cluster connection fields persisted", saved)
	}
}
