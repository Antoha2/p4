package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"p4/internal/config"
	"p4/internal/helper"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	run()
}

func run() {
	cfg := config.GetConfig()
	dbx, err := initDb(cfg)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	//select id, name from Employees  where id = 1
	err = SelectData(dbx)
	if err != nil {
		log.Println(err)
	}

	//INSERT into Departments (dep_name) values('АХЧ')
	err = InsertData(dbx)
	if err != nil {
		log.Println(err)
	}

	// update Employees set name='Robert' where name='Rob'
	err = UpdateData(dbx)
	if err != nil {
		log.Println(err)
	}

	//transaction - insert into Employees (name, DepartamentId, ProjectId) values('Pol', 1, 2) and
	// insert into Departments (dep_name) values('Logistic')
	err = Transaction(dbx)
	if err != nil {
		log.Println(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
}

// SELECT id, name from Employees  where id = 1
func SelectData(dbx *gorm.DB) error {

	query := "select id, name from Employees where id = $1"
	result := dbx.Table("Employees").Raw(query, 1)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("record not found")
	}
	log.Println("SELECT ok")

	return nil
}

// INSERT into Departments (dep_name) values('АХЧ')
func InsertData(dbx *gorm.DB) error {

	dep := &helper.Departament{DepName: "АХЧ"}

	result := dbx.Table("departaments").Create(dep).Scan(&dep)
	if errors.Is(result.Error, gorm.ErrInvalidValue) {
		return errors.New("invalid value")
	}
	log.Println("INSERT ok")
	return nil

}

// UPDATE Employees set name='Robert' where name='Rob'
func UpdateData(dbx *gorm.DB) error {

	result := dbx.Table("employees").Model(helper.Employee{}).Where("name=?", "Rob").Updates(helper.Employee{Name: "Robert"})
	if errors.Is(result.Error, gorm.ErrInvalidData) {
		return errors.New("unsupported data")
	}
	log.Println("UPDATE ok")
	return nil
}

// transaction - insert into Employees (name, DepartamentId, ProjectId) values('Pol', 1, 2) and
// insert into Departments (dep_name) values('Logistic')
func Transaction(dbx *gorm.DB) error {

	tx := dbx.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Table("employees").Create(&helper.Employee{Name: "Pol", DepartamentId: 1, ProjectId: 2}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Table("departaments").Create(&helper.Departament{DepName: "Logistic"}).Error; err != nil {
		tx.Rollback()
		return err
	}

	log.Println("TRANSACTION ok")
	return tx.Commit().Error
}

func initDb(cfg *config.Config) (*gorm.DB, error) {

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=%d",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Dbname,
		cfg.DB.Sslmode,
		5,
	)

	connConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	// Make connections
	dbx, err := sqlx.Open("pgx", stdlib.RegisterConnConfig(connConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection db: %v", err)
	}

	dbx.SetMaxIdleConns(10)
	dbx.SetMaxOpenConns(100)
	dbx.SetConnMaxLifetime(time.Hour)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbx,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open(): %v", err)
	}

	err = dbx.Ping()
	if err != nil {
		return nil, fmt.Errorf("error to ping connection pool: %v", err)
	}
	log.Printf("Подключение к базе данных на http://127.0.0.1:%d\n", cfg.DB.Port)
	return gormDB, nil
}
