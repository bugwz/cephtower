# Database Schema

This directory records the current CephTower database table structure for each
supported database engine.

- `mysql/schema.sql`: MySQL DDL
- `sqlite/schema.sql`: SQLite DDL

The schema mirrors the GORM models in `backend/internal/store/models.go`.
