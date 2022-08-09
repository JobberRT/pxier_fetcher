package core

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"time"
)

type Factory struct {
	executors []Executor
	database  *gorm.DB
}

func NewFactory() *Factory {
	logrus.Info("creating factory")
	f := &Factory{
		executors: make([]Executor, 0),
		database:  newDB(),
	}
	return f
}

func (f *Factory) RegisterExecutor(e Executor) {
	f.executors = append(f.executors, e)
}

func (f *Factory) Start() {
	interval := viper.GetInt64("factory.fetch_interval")
	if interval == 0 {
		interval = 10
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		logrus.Info("factory fetch")
		for _, e := range f.executors {
			exe := e
			go func() {
				f.saveToDB(exe.Fetch())
			}()
		}
		<-ticker.C
	}
}

func newDB() *gorm.DB {
	logrus.Info("start mysql")
	url := viper.GetString("mysql_url")
	if len(url) == 0 {
		logrus.Panic("mysql url is empty")
	}
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		logrus.WithError(err).Panic("failed to create db")
	}
	if err := db.AutoMigrate(&proxy{}); err != nil {
		logrus.WithError(err).Panic("failed to migrate model")
	}
	d, _ := db.DB()
	d.SetMaxIdleConns(10)
	d.SetMaxOpenConns(100)
	d.SetConnMaxLifetime(time.Hour)
	return db
}

func readBody(res *fasthttp.Response) ([]byte, error) {
	logrus.Info("reading response body")
	switch string(res.Header.Peek("Content-Type")) {
	case "br":
		return res.BodyUnbrotli()
	case "gzip":
		return res.BodyGunzip()
	case "deflate":
		return res.BodyInflate()
	default:
		return res.Body(), nil
	}
}

func (f *Factory) saveToDB(proxies []*proxy) {
	if len(proxies) == 0 {
		return
	}
	for _, pxy := range proxies {
		if db := f.database.Model(&proxy{}).
			Where("address = ? and dial_type = ?", pxy.Address, pxy.DialType).
			Update("updated_at", time.Now().Unix()); db.RowsAffected == 0 {
			pxy.ErrTimes = 0
			pxy.CreatedAt = time.Now().Unix()
			pxy.UpdatedAt = time.Now().Unix()
			f.database.Create(&pxy)
		}
	}
}
