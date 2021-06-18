package stream_3

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

type UserServerImpl struct{}

//one req and stream resp.
func (u *UserServerImpl) GetUserList(req *GetUserRequest, stream User_GetUserListServer) error {
	for i := 0; i < 10; i++ {
		resp := GetUserResponse{User: &UserDTO{Id: strconv.Itoa(i), Name: "name" + strconv.Itoa(i)}}
		stream.Send(&resp)
	}
	return nil
}

//stream req and stream resp.
func (u *UserServerImpl) GetUserByStream(stream User_GetUserByStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		resp := GetUserResponse{User: &UserDTO{Id: req.Id, Name: "name" + req.Id}}
		stream.Send(&resp)
	}
}

//1.listen -> 2.server -> register -> serve
func Test_server(t *testing.T) {
	//1. listen
	listen, _ := net.Listen("tcp", ":8889")
	//2. server
	server := grpc.NewServer()
	//3.register
	RegisterUserServer(server, &UserServerImpl{})
	//4.serve
	server.Serve(listen)
}
//1.dail 2.client 3.rpc
func Test_client_1(t *testing.T) {
	//1.dail
	conn, _ := grpc.Dial(":8889", grpc.WithInsecure())
	//2.client
	client := NewUserClient(conn)
	//3.call
	stream, _ := client.GetUserList(context.Background(), &GetUserRequest{Id: "1"})
	for  {
		resp, err := stream.Recv()
		if err == io.EOF  {
			break
		}
		if err != nil {
			fmt.Printf("%s\n", "error stream recv")
			break
		}
		fmt.Printf("resp.User.Name=%#v\n", resp.User.Name)
	}
}
//this just like async.
//1.dail 2.client 3.rpc
func Test_client_2(t *testing.T) {
	conn, _ := grpc.Dial(":8889", grpc.WithInsecure())
	client := NewUserClient(conn)
	stream, _ := client.GetUserByStream(context.Background())
	go func() {
		for  {
			resp, err := stream.Recv()
			if err == io.EOF  {
				break
			}
			if err != nil {
				fmt.Printf("%s\n", "error stream recv")
				break
			}
			fmt.Printf("resp.User.Name=%#v\n", resp.User.Name)
		}
	}()
	for i := 0; i < 10; i++ {
		stream.Send(&GetUserRequest{Id: strconv.Itoa(i)})
	}
	time.Sleep(time.Second)

}
