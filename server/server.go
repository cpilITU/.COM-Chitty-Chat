package main

import (
	"context"
	"flag"
	"log"
	"net"
	"strconv"
	proto "github.com/cpilITU/Chitty-Chat/proto"
	"google.golang.org/grpc"
)

var (
	connections []*Connection
	Port        = flag.Int("port", 0, "server port number")
)

type Server struct {
	proto.UnimplementedChittyChatServer // Nessecery
	Connection                          []*Connection
	Port                                int
	LamportTime                         int
}

type Connection struct {
	stream proto.ChittyChat_CreateConnectionServer
	id     string
	active bool
	error  chan error
}

func main() {
	flag.Parse()

	server := &Server{
		Connection: connections,
		Port:       *Port,
		LamportTime:    0,
	}

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(server.Port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.Port)

	proto.RegisterChittyChatServer(grpcServer, server)
	grpcServer.Serve(listener)
}

func (s *Server) CreateConnection(user *proto.User, stream proto.ChittyChat_CreateConnectionServer) error {
	//Update Lamport time when message is recived
	max(user.LamportTime, s)
	conn := &Connection{
		stream: stream,
		id:     user.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, conn)
	
	s.LamportTime++
	s.BroadcastMessage(&proto.Message{
		Id:          user.Id,
		Content:     "Participant " + user.Id + " joined Chitty-Chat at Lamport time " + strconv.FormatInt(user.LamportTime, 10),
		LamportTime: int64(s.LamportTime),
	}, stream)
	return <-conn.error
}

func (s *Server) BroadcastMessage(msg *proto.Message, stream proto.ChittyChat_BroadcastMessageServer) error {
	//Update Lamport time when message is recived
	max(msg.LamportTime, s)
	log.Printf("Client %s publishes message {msg : %s} to the server at Lamport time {time : %d}", msg.Id, msg.Content, s.LamportTime)
	// Updateing Lamport when sending a message
	s.LamportTime++
	for _, conn := range s.Connection {
		if conn.active {	
			conn.stream.Send(&proto.Message{
				Id:          msg.Id,
				Content:     msg.Content,
				LamportTime: int64(s.LamportTime),
			})
			log.Printf("Server broadcasts message {msg : %s} to client %s at Lamport time {time : %d}", msg.Content, conn.id, s.LamportTime)
		}
	}
	return nil
}

func (s *Server) UserLeft(ctx context.Context, user *proto.User) (*proto.Close, error) {
	//Update Lamport time when message is recived
	max(user.LamportTime, s)
	for _, conn := range s.Connection {
		if conn.id == user.Id {
			conn.active = false		
		}
	}
	return &proto.Close{}, nil
}

func max(ClientLamport int64, server *Server) {
	if int(ClientLamport) > server.LamportTime {
		server.LamportTime = int(ClientLamport) + 1
		return
	}
	server.LamportTime++
}

