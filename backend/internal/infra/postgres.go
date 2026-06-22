package infra

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	config "backend/internal/config"
	"backend/internal/logger"
)

// PostgresDB 包装 *gorm.DB。
type PostgresDB struct {
	DB *gorm.DB
}

// NewPostgresClient 创建并验证 PostgreSQL 连接。初始化过程通过 appLog 记录。
func NewPostgresClient(appLog *logger.AppLogger) (*PostgresDB, error) {
	dsn := config.AppConfig.PostgreSQL.DSN
	appLog.Info("正在连接 PostgreSQL", "dsn", dsn)

	dialector := postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: false,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		appLog.Error(err, "PostgreSQL 连接失败")
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		appLog.Error(err, "获取 sql.DB 失败")
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		appLog.Error(err, "PostgreSQL Ping 失败")
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	appLog.Info("PostgreSQL 连接成功")
	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) Close(appLog *logger.AppLogger) error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
