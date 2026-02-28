package main

import (
	"fmt"
	"gormal/gormal"
	"os"
	"time"
)

func main() {
	reader := make(chan string)
	gorm, err := gormal.NewGormalStdin()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(gorm)
	//err = gorm.DropFlag(gormal.ICANON)
	fmt.Println(gorm, err)

	go Reader(reader)

	for line := range reader {
		fmt.Println(line)
		if line == "x" {
			os.Exit(0)
		}
		if line == "e" {
			gorm.AppendFlag(gormal.ICANON)
		}
		if line == "n" {
			gorm.DropFlag(gormal.ICANON)
		}

	}
}

func Reader(inp chan string) {
	b := make([]byte, 1)
	for {
		os.Stdin.Read(b)
		inp <- string(b)
		time.Sleep(1000)
	}

}
