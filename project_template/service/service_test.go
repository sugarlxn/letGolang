package service

import (
	"fmt"
	"testing"
)

// 单元测试放在每个package下，文件名以_test.go结尾，单元测试函数名以Test开头
func TestService(t *testing.T) {
	exceptedUtils := "service_utils.package"
	if ServiceUtils_package_variable != exceptedUtils {
		t.Errorf("Expected %s, but got %s", exceptedUtils, ServiceUtils_package_variable)
	} else {
		fmt.Println("ServiceUtils_package_variable test passed.")
	}

	exceptedService := "service.package"
	if Service_package_variable != exceptedService {
		t.Errorf("Expected %s, but got %s", exceptedService, Service_package_variable)
	} else {
		fmt.Println("Service_package_variable test passed.")
	}
}
