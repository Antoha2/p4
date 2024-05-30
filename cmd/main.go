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
	err = selectWhere(dbx)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	//INSERT into Departments (dep_name) values('АХЧ')
	err = insertData(dbx)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// update Employees set name='Robert' where name='Rob'
	err = updateData(dbx)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// update Employees set name='Robert' where name='Rob'
	err = transaction(dbx)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
}

func selectWhere(dbx *gorm.DB) error {

	emp := &helper.Employee{}
	query := "select id, name from Employees  where id = $1"
	result := dbx.Table("Employees").Raw(query, 1).Scan(emp)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("record not found")
	}
	log.Println("select id, name from Employees where id = 1 - ok")

	return nil
}

func insertData(dbx *gorm.DB) error {

	dep := &helper.Departament{DepName: "АХЧ"}

	result := dbx.Table("departaments").Create(dep).Scan(&dep)
	if errors.Is(result.Error, gorm.ErrInvalidValue) {
		return errors.New("invalid value")
	}
	log.Println("INSERT into Departments (dep_name) values('АХЧ') - ok")
	return nil

}

func updateData(dbx *gorm.DB) error {

	result := dbx.Table("employees").Model(helper.Employee{}).Where("name=?", "Rob").Updates(helper.Employee{Name: "Robert"})
	if errors.Is(result.Error, gorm.ErrInvalidData) {
		return errors.New("unsupported data")
	}
	log.Println("update Employees set name='Robert' where name='Rob' ok")
	return nil
}

func transaction(dbx *gorm.DB) error {

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

	log.Println("Используйте транзакции для вставки нового отдела и сотрудника - ok")
	return tx.Commit().Error
}

func initDb(cfg *config.Config) (*gorm.DB, error) {

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Dbname,
		cfg.DB.Sslmode,
	)

	connConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("1 failed to parse config: %v", err)
	}

	// Make connections
	dbx, err := sqlx.Open("pgx", stdlib.RegisterConnConfig(connConfig))
	if err != nil {
		return nil, fmt.Errorf("2 failed to create connection db: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbx,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("3 gorm.Open(): %v", err)
	}

	err = dbx.Ping()
	if err != nil {
		return nil, fmt.Errorf("4 error to ping connection pool: %v", err)
	}
	log.Printf("Подключение к базе данных на http://127.0.0.1:%d\n", cfg.DB.Port)
	return gormDB, nil
}
