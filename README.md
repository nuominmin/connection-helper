# connection-helper

## 概览

## 功能

## 安装
```bash
go get github.com/nuominmin/connection-helper
```

## 简单示例(Mysql)
```go
package database

import (
	"database/sql"
	"errors"
	connectionhelper "github.com/nuominmin/connection-helper"
	"google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Config struct {
	Source string
	// 日志等级. 默认 silent
	// 1. silent
	// 2. error
	// 3. warn
	// 4. info
	LogLevel int64
	// 最大空闲连接数. 默认: 10
	MaxIdleConns int64
	// 最大活动连接数. 默认: 100
	MaxOpenConns int64
	// 连接最大存活时间. 默认: 300s
	ConnMaxLifetime *durationpb.Duration
}
type Connector struct {
	conf *Config
	conn *gorm.DB
}

func New(conf *Config) (*connectionhelper.RetryExecutor[*gorm.DB], error) {
	return connectionhelper.New[*gorm.DB](&Connector{
		conf: conf,
	})
}

func (c *Connector) Connect() (*gorm.DB, error) {
	conn, err := gorm.Open(mysql.Open(c.conf.Source), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 设置日志等级
	conn.Logger = conn.Logger.LogMode(logger.LogLevel(c.conf.LogLevel))

	// 设置数据库配置
	var db *sql.DB
	if db, err = conn.DB(); err != nil {
		return nil, err
	}

	if c.conf.MaxIdleConns != 0 {
		db.SetMaxIdleConns(int(c.conf.MaxIdleConns))
	} else {
		db.SetMaxIdleConns(10)
	}

	if c.conf.MaxOpenConns != 0 {
		db.SetMaxOpenConns(int(c.conf.MaxOpenConns))
	} else {
		db.SetMaxOpenConns(100)
	}

	if c.conf.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(c.conf.ConnMaxLifetime.AsDuration())
	} else {
		db.SetConnMaxLifetime(time.Second * 300)
	}

	return conn, nil
}

func (c *Connector) GetConn() *gorm.DB {
	return c.conn
}

func (c *Connector) Close() error {
	sqlDB, err := c.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (c *Connector) IsConnectionError(err error) bool {
	return errors.Is(err, gorm.ErrInvalidField)
}

```

```
// 初始化
data := &Data{}
databaseConf := &database.Config{
	Source:          c.Database.Source,
	LogLevel:        c.Database.LogLevel,
	MaxIdleConns:    c.Database.MaxIdleConns,
	MaxOpenConns:    c.Database.MaxOpenConns,
	ConnMaxLifetime: c.Database.ConnMaxLifetime,
}

var err error
if data.executor, err = database.New(databaseConf); err != nil {
	return nil, nil, err
}


// 逻辑方法内调用
err := r.data.executor.ExecWithRetry(ctx, func(ctx context.Context, conn *gorm.DB) error {
	return conn.WithContext(ctx).FirstOrCreate(&address{}, address{Address: addresses[i]}).Error
})
if err != nil {
	return err
}

```




