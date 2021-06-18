package auth_4

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "grpc_demo/proto_1"
	"net"
	"testing"
)

const (
	certificatePath = "/root/go/src/grpc_demo/auth_4"
)
type UserService struct{}

func (u *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	fmt.Printf("%s\n", "getuser()")
	return &pb.GetUserResponse{User: &pb.UserDTO{Id: req.Id, Name: "user:" + req.Id}}, nil
}

//1.listen 2.server(add auth here) 3.register 4.serve
func Test_server(t *testing.T) {
	listen, _ := net.Listen("tcp", ":8888")
	creds, _ := credentials.NewServerTLSFromFile(certificatePath+"/server.crt", certificatePath+"/server.key")
	server := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterUserServer(server,&UserService{})
	server.Serve(listen)
}
//1.dail(add auth here) 2.client 3.rpc
func Test_client(t *testing.T) {
	creds, _ := credentials.NewClientTLSFromFile(certificatePath+"/ca.crt", "www.lixueduan.com")
	conn, _ := grpc.Dial(":8888", grpc.WithTransportCredentials(creds))
	client := pb.NewUserClient(conn)
	resp, _ := client.GetUser(context.Background(), &pb.GetUserRequest{Id: "1"})
	fmt.Printf("resp.User.Name=%#v\n", resp.User.Name)
}