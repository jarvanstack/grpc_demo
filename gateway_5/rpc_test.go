package gateway_5

import (
	"context"
	"google.golang.org/grpc"
	"grpc_demo/gateway_5/rpc"
	"net"
	"testing"
)

type HelloServerImpl struct {}
func (h *HelloServerImpl) SayHello(ctx context.Context,req *gw.StringMessage) (*gw.StringMessage, error) {
	return &gw.StringMessage{Message: "hi"},nil
}
func TestRpcServer(t *testing.T) {
	listen, _ := net.Listen("tcp", ":50052")
	server := grpc.NewServer()
	gw.RegisterHelloHTTPServer(server,&HelloServerImpl{})
	server.Serve(listen)
}