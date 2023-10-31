package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"strconv"
	"time"
	"fmt"
	proto "github.com/cpilITU/Chitty-Chat/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	}

	client := proto.NewChittyChatClient(conn)
	user := createUser()

	stream := createSomething(client, user)

	go func(str proto.ChittyChat_CreateConnectionClient) {
		for {
			msg, _ := str.Recv()
			//Update Lamport time when message is recived
			max(msg.LamportTime, user)
			fmt.Printf("%v : %s {Recived at Lamport time %d}\n", msg.Id, msg.Content, user.LamportTime)
		}
	}(stream)

	writeMessage(client, user)
}

func createUser() *proto.User {
	
	id := sha256.Sum256([]byte(time.Now().String()))

	return &proto.User{
		Id:          hex.EncodeToString(id[0:3]),
		LamportTime: 0,
	}
}

func createSomething(client proto.ChittyChatClient, user *proto.User) proto.ChittyChat_BroadcastMessageClient {
	user.LamportTime++
	
	stream, err := client.CreateConnection(context.Background(), &proto.User{
		Id:          user.Id,
		LamportTime: user.LamportTime,
	})
	if err != nil {
		log.Printf("There was an error creating user %s", user.Id)
		os.Exit(1)
	}

	return stream
}

func writeMessage(client proto.ChittyChatClient, user *proto.User) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var tempHolderForUserMSG = scanner.Text()
		if len(tempHolderForUserMSG) <= 128 {
			// Updateing Lamport when sending a message
			user.LamportTime++
			if tempHolderForUserMSG == "quit" {
				client.UserLeft(context.Background(), user)
				user.LamportTime++
				
				break
			} else {
				_, err := client.BroadcastMessage(context.Background(), &proto.Message{
					Id:          user.Id,
					Content:     scanner.Text() + " at Lamport time " + strconv.FormatInt(user.LamportTime, 10),
					LamportTime: user.LamportTime,
				})
				if err != nil {
					fmt.Printf("Error Sending Message: %v", err)
					break
				}
			}
		} else {
			fmt.Println("Invalid message : Message must be between 0 and 128 characters")
		}
	}
}

func max(ServerLamport int64, user *proto.User) {
	if ServerLamport > user.LamportTime {
		user.LamportTime = ServerLamport + 1
		return
	}
	user.LamportTime++
}
