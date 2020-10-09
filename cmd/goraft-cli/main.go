package main

import (
	"fmt"
	"bufio"
	"os"
	"log"
	"strings"
	"encoding/json"
	"github.com/louisinger/go-raft/internal"
)

func mustHaveXParam(x int, splitted []string) bool {
	lenArgs := len(splitted[1:])
	if (lenArgs < x) {
		log.Println("Error: there is", lenArgs, "arguments but the command needs", x, "arguments.")
		return false
	}
	return true
}

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
				paramExists := mustHaveXParam(1, splitted)
				if paramExists {
					log.Println(client.Info(splitted[1]))
				}
			case "newpeer":
				paramExists := mustHaveXParam(1, splitted)
				if paramExists {
					log.Println(client.NewPeer(splitted[1]))
				}
			case "put":
				if mustHaveXParam(2, splitted) {
					data := make(map[string]interface{}) 
					data["value"] = splitted[2]
					json, _ := json.Marshal(data)
					entries := []internal.PutEntriesArgs{
						internal.PutEntriesArgs{Key: splitted[1], JsonStr: string(json)},
					}
					log.Println(client.Put(entries))
				}
			case "exit":
				os.Exit(3)
			default:
				fmt.Println("Unknow command")
			}
	}
}