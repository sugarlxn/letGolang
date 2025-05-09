package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)



type auth_user struct {
	id        int
	username  string
	email     string
	is_active bool
}

func (u auth_user) String() string {
	return fmt.Sprintf("ID: %d, Username: %s, Email: %s, IsActive: %t", u.id, u.username, u.email, u.is_active)
}

// 单笔查询
func getOne(db *sql.DB, id int) (auth_user, error) {
	var user auth_user
	query := "SELECT id, username, email, is_active FROM auth_user WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&user.id, &user.username, &user.email, &user.is_active)
	if err != nil {
		return auth_user{}, err
	}
	return user, nil
}

// 多笔查询
func getMany(db *sql.DB, id int) ([]auth_user, error) {
	var users []auth_user
	query := "SELECT id, username, email, is_active FROM auth_user WHERE id > ?"
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user auth_user
		if err := rows.Scan(&user.id, &user.username, &user.email, &user.is_active); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func main() {

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, server, port, database)
	fmt.Println(connStr)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()
	ctx := context.Background()
	// Check if the connection is alive
	if err := db.PingContext(ctx); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Connected to the database successfully!")

	// Query the database 单笔查询
	var user auth_user
	query_err := db.QueryRow("SELECT id,username,email,is_active FROM auth_user WHERE id = ?", 1).Scan(
		&user.id, &user.username, &user.email, &user.is_active)

	if query_err != nil {
		log.Fatalln(query_err)
	}
	fmt.Println(user)

}
