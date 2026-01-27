package main

import (
	"fmt"
	"template/DButils"
	"template/service"
)

func main() {
	fmt.Println("main.package")
	fmt.Println(DButils.DButils_package_variable)
	fmt.Println(service.ServiceUtils_package_variable)
	fmt.Println(service.Service_package_variable)

	db, err := DButils.ConnectToDB()
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()
	fmt.Println("Successfully connected to the database.")

}
