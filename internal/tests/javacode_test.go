package test

import (
	"fmt"
	"log"
	"os"
	"time"

	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"

	"p4/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	once sync.Once
)

func NewTransSingleton() {
	once.Do(func() {
		InitTest()
	})
}

func InitTest() {

	cfg := config.GetConfig()
	dbx := MustInitDb(cfg)

	dbx.Raw("drop table projects")
	dbx.Raw("drop table departaments")
	dbx.Raw("drop table employees")

}

func MustInitDb(cfg *config.Config) *gorm.DB {

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
		os.Exit(1)
	}

	// Make connections
	dbx, err := sqlx.Open("pgx", stdlib.RegisterConnConfig(connConfig))
	if err != nil {
		os.Exit(1)
	}

	dbx.SetMaxIdleConns(10)
	dbx.SetMaxOpenConns(100)
	dbx.SetConnMaxLifetime(time.Hour)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbx,
	}), &gorm.Config{})
	if err != nil {
		os.Exit(1)
	}

	err = dbx.Ping()
	if err != nil {
		os.Exit(1)
	}
	log.Printf("Подключение к базе данных на http://127.0.0.1:%d\n", cfg.DB.Port)
	return gormDB
}

//test ChangeSum
func SelectData(t *testing.T) {
	tests := []struct {
		want    string
		wantErr bool
	}{

		{"select ok", false},
		{"record not found", true},
	}

	NewTransSingleton()
	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.inputSum))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("User-Role", tc.inputRole)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := s.ChangingSumHandler(c)

			if (err != nil) != tc.wantErr { // если ошибка не нил , и не ждем ошибку
				t.Fatalf("error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if (err != nil) && tc.wantErr { // если ошибка не нил , и ждем ошибку
				return
			}
			if !reflect.DeepEqual(rec.Body.String(), tc.want) { //если нет ошибки , то сравниваем значения
				t.Fatalf("expected: %v, got: %v", rec.Body.String(), tc.want)
			}
		})
	}
}
