package main

import (
	"Monkey_1/repl"
	"fmt"
	"os"
	user1 "os/user"
)

func main() {
	user, err := user1.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is Monkey Programming language!\n", user.Username)
	fmt.Printf("Feel free to type in commands \n")
	repl.Start(os.Stdin, os.Stdout)
}
