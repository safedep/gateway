package adapters

import (
	"fmt"
	"log"
	"time"

	"github.com/safedep/gateway/services/pkg/common/utils"
	"golang.org/x/net/context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySqlAdapter struct {
	db     *gorm.DB
	config MySqlAdapterConfig
}

type MySqlAdapterConfig struct {
	Host     string
	Port     int16
	Username string
	Password string
	Database string
}

func NewMySqlAdapter(config MySqlAdapterConfig) (SqlDataAdapter, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	log.Printf("Connecting to MySQL database with %s@%s:%d", config.Username, config.Host, config.Port)

	var db *gorm.DB
	var err error

	retryN := 5
	utils.InvokeWithRetry(utils.RetryConfig{
		Count: retryN,
		Sleep: 1 * time.Second,
	}, func(n int) error {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("[%d/%d] Failed to connect to MySQL server: %v",
				n, retryN, err)
		}

		return err
	})

	if err != nil {
		return nil, err
	}

	mysqlAdapter := &MySqlAdapter{db: db, config: config}
	err = mysqlAdapter.Ping()

	return mysqlAdapter, err
}

func (m *MySqlAdapter) GetDB() (*gorm.DB, error) {
	return m.db, nil
}

func (m *MySqlAdapter) Migrate(tables ...interface{}) error {
	return m.db.AutoMigrate(tables...)
}

func (m *MySqlAdapter) Ping() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}

	ctx, cFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cFunc()

	return sqlDB.PingContext(ctx)
}
