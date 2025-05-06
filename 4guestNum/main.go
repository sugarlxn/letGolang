package main

import (
	"fmt"
	"math/rand"
)

func main() {
	fmt.Println("Guest Num Game!!")
	var target_num int = rand.Intn(100) + 1
	for {
		fmt.Println("输入你猜测的数字：")
		var guess int
		_, err := fmt.Scanln(&guess)
		if err != nil {
			fmt.Println("输入错误，请输入一个有效数字")
			//清除输入缓冲区
			var discard string
			fmt.Scanln(&discard)
			continue
		}

		if guess < target_num {
			fmt.Println("太小了，再大一点")
		} else if guess > target_num {
			fmt.Println("太大了，再小一点")
		} else {
			fmt.Println("猜对了！！ 数字就是 ", target_num)
			break
		}
	}
}
