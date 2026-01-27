package DButils

import (
	"fmt"
	"testing"
)

// 单元测试放在每个package下，文件名以_test.go结尾, 单元测试函数名以Test开头
func TestDButilsPackageVariable(t *testing.T) {
	expected := "DButils.package"
	if DButils_package_variable != expected {
		t.Errorf("Expected %s, but got %s", expected, DButils_package_variable)
	} else {
		fmt.Println("DButils_package_variable test passed.")
	}
}

func TestConnectToDB(t *testing.T) {
	db, err := ConnectToDB()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	} else {
		fmt.Println("ConnectToDB test passed.")
	}
	if err = db.Ping(); err != nil {
		t.Errorf("Error pinging database: %v", err)
	} else {
		fmt.Println("Ping test passed.")
	}
	defer db.Close()
}
