package main

import (
	"fmt"
	"bufio"
	"os"
	"log"
	"strings"
	"github.com/louisinger/go-raft/internal"
)

func main()  {
	fmt.Println("Welcome inside the goraft cli.")
	if (len(os.Args) < 2) {
		log.Fatal("You must specify the server path (domain:port)")
	}
	serverPath := os.Args[1]

	client, err := internal.NewClient(serverPath)
	if (err != nil) {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Input command:")
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		splitted := strings.Split(text, " ")

		switch splitted[0] {
		case "info":
			log.Println(client.Info(splitted[1]))
		case "exit":
			break
		default:
			fmt.Println("Unknow command")
		}
	}
}