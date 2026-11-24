package infas

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // Registers the "sqlite" driver name
)

type DB_connection struct {
	DB *sql.DB
}

func (app *DB_connection) OpenDB() error {
	db, err := sql.Open("sqlite", "/home/duylong/code/mocatestdeploy/container_src/moca_test.db")
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	app.DB = db
	return nil
}
func (app *DB_connection) InitDB() error {
	schema, err := os.ReadFile("infas/schema.sql") // hoặc đường dẫn tới file schema.sql
	if err != nil {
		return fmt.Errorf("không đọc được schema: %w", err)
	}
	_, err = app.DB.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("lỗi khởi tạo bảng: %w", err)
	}
	return nil
}
