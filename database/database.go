package database

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

type Users struct {
	SessionId string `gorm:"session_id"`
	Frequency float64
	Timestamp time.Time
}

type PgConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"dbname"`
}

func ConnectToDB() (*gorm.DB, error) {
	var pg PgConfig
	pg.GetConf()

	cfg := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", pg.Host, pg.Username, pg.Password, pg.Database, pg.Port)
	db, err := gorm.Open(postgres.Open(cfg), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Users{})
	return db, nil
}

func (p *PgConfig) GetConf() *PgConfig {
	conf, err := os.ReadFile("../../database/pg.yaml")
	if err != nil {
		log.Fatalf("file not found: %v", err)
	}
	err = yaml.Unmarshal(conf, p)
	if err != nil {
		log.Fatalf("error unmarshalling yaml file: %v", err)
	}
	return p
}
