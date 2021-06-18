package rpc_call_2

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "grpc_demo/proto_1"
	"net"
	"testing"
)

// Service and implements the interface.
type UserService struct{}

func (u *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return &pb.GetUserResponse{User: &pb.UserDTO{Id: req.Id, Name: "user:" + req.Id}}, nil
}

//1.listen -> 2.server -> 3.register -> 4.serve
//1. net listen
//2. grpc.NewServer()
//3. gw.Register...Server(server,&..Service{})
//4. server.Serve(listen)
func Test_server(t *testing.T) {
	//1.listen
	listen, _ := net.Listen("tcp", ":8888")
	//2. server
	server := grpc.NewServer()
	//3.register
	pb.RegisterUserServer(server, &UserService{})
	//4.run
	server.Serve(listen)
}

//1.dail -> 2.client -> 3.rpc
//1. grpc dail
//2. gw.New...Client(conn)
//3. client.Method(params..)
func Test_client(t *testing.T) {
	conn, _ := grpc.Dial(":8888", grpc.WithInsecure())
	defer conn.Close()
	client := pb.NewUserClient(conn)
	user, _ := client.GetUser(context.Background(), &pb.GetUserRequest{Id: "1"})
	fmt.Printf("%s\n", user.String())
}
