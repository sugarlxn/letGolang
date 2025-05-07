package main

import (
	"fmt"
	"os"
	"strings"
)

// 自定义error 自定义错误类型因该以Error结尾
type MyError []error

func (m MyError) Error() string {
	var errStr []string
	for _, err := range m {
		errStr = append(errStr, err.Error())
	}
	return strings.Join(errStr, ", ")
}

func proverbs(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, "Error are values")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, "don't just check errors, handle them gracefully!")
	return err
}

func main() {
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}
	for _, file := range files {
		if file.IsDir() {
			fmt.Println("Directory:", file.Name())
		} else {
			fmt.Println("File:", file.Name())
		}
	}

	err = proverbs("proverbs.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
}
