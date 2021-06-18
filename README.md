## grpc_demo

Here is grpc demo.

### 1. quick start

grpc package

```bash
google.golang.org/grpc v1.38.0
```
> NOTE: you also can get the latest version.

### 2. proto 

new protobuf file. `proto_1/user.proto`

```protobuf
syntax = "proto3";
package proto1;
option go_package = "proto1/";
service User {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
}
message GetUserRequest {
  string id = 1;
}
message GetUserResponse {
  UserDTO user = 1;
}
message UserDTO {
  string id = 1 ;
  string name = 2;
}
```


run the `generate.sh` to generate all protobuf go file.

```shell
for x in **/*.gw; do protoc --go_out=plugins=grpc,paths=source_relative:. $x; done
```

### 2. rpc_call

#### server

```go
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
```

#### client

```go
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
```

### 3. stream 

proto buf

```protobuf
syntax = "proto3";
package proto1;
option go_package = "proto1/";
service User {
  //one req and stream resp.
  rpc GetUserList(GetUserRequest) returns (stream GetUserResponse){}
  //stream req and stream resp.
  rpc GetUserByStream(stream GetUserRequest) returns (stream GetUserResponse){}
}
message GetUserRequest {
  string id = 1;
}
message GetUserResponse {
  UserDTO user = 1;
}
message UserDTO {
  string id = 1 ;
  string name = 2;
}
```


(1) Server-client stream

(2) bidirectional stream(double stream)

all the code see [stream_3/stream_test.go](stream_3/stream_test.go)

```go
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
```

### 4.  Authentication.

Authentication.(Auth)

> Application Layer Transport Security.(ALTS) is only used for google cloud. we do not care that here.

> reference: https://lixueduan.com/post/grpc/04-encryption-tls/

#### (1) CA certificate.


```shell
cd auth_4
```


```bash
# 生成.key  私钥文件
openssl genrsa -out ca.key 2048

# 生成.csr 证书签名请求文件
openssl req -new -key ca.key -out ca.csr  -subj "/C=GB/L=China/O=lixd/CN=www.lixueduan.com"

# 自签名生成.crt 证书文件
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt  -subj "/C=GB/L=China/O=lixd/CN=www.lixueduan.com"
```

#### (2) server certificate 

```shell
# find where is your openssl.cnf
find / -name "openssl.cnf"
/etc/pki/tls/openssl.cnf
# 1. key
openssl genrsa -out server.key 2048
# 2. .csr
openssl req -new -key server.key -out server.csr \
	-subj "/C=GB/L=China/O=lixd/CN=www.lixueduan.com" \
	-reqexts SAN \
	-config <(cat /etc/pki/tls/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.lixueduan.com,DNS:*.refersmoon.com"))
# 3. .crt
openssl x509 -req -days 3650 \
   -in server.csr -out server.crt \
   -CA ca.crt -CAkey ca.key -CAcreateserial \
   -extensions SAN \
   -extfile <(cat /etc/ssl/openssl.cnf <(printf "\n[SAN]\nsubjectAltName=DNS:*.lixueduan.com,DNS:*.refersmoon.com"))
```

we only used these files

1. server.key

2. server.crt

3. ca.crt


#### code.

```go
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
```

complete code see `auth_4/auth_test.go`  file.


### 5. gateway

> reverse proxy grpc to http at same time.

> but I think is useless for me to build my micro server, I can create my own api at gateway.

> reference: https://github.com/grpc-ecosystem/grpc-gateway


#### (1) install dependencies.

```shell
go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

### install protobuf compiler 

there are two ways to install your protobuf compiler.

1. protoc + protoc-gen-go(not recommend, if you use this you will care where is "google/api/annotations.proto", you should find them and add "-I ." to protoc , fuck ! )

> issues: https://github.com/grpc-ecosystem/grpc-gateway/issues/1935
> ...

[how to install protoc and protoc-gen-go in linux.](https://blog.csdn.net/jarvan5/article/details/118026721)

2. buf (recommend, and I will use this way here)

[how to install buf in linux](https://blog.csdn.net/jarvan5/article/details/117918779)


### create protobuf 

```protobuf
syntax = "proto3";
package gateway;
option go_package = "/gw";
import "google/api/annotations.proto";
// 定义Hello服务
service HelloHTTP {
  // 定义SayHello方法
  rpc SayHello(StringMessage) returns (StringMessage) {
    // http option
    option (google.api.http) = {
      post: "/example/echo"
      body: "*"
    };
  }
}
message StringMessage {
  string message = 1;
}
```
### generate go code by buf

you should create two files

```bush
buf.gen.yaml
buf.yaml
buf.lock
```

#### buf.gen.yaml

```yaml
version: v1beta1
plugins:
  - name: go
    out: ./
    opt:
      - paths=source_relative
  - name: go-grpc
    out: ./
    opt:
      - paths=source_relative
  - name: grpc-gateway
    out: ./
    opt:
      - paths=source_relative
      # user custom api config you do not need it.
      #- grpc_api_configuration=path/to/config.yaml
      - standalone=true
```

#### buf.yaml (add deps)

```yaml
version: v1beta1
name: buf.build/yourorg/myprotos
deps:
  # it will download this dep if you run buf generate
  - buf.build/beta/googleapis
```

#### buf.lock

run `buf beta mod update` to update file   `buf.lock`

### generate code

run 

```shell
[root@c03 gw]# buf generate
[root@c03 gw]# ls
buf.gen.yaml  buf.lock  buf.yaml  hello_grpc.pb.go  hello.pb.go  hello.pb.gw.go  hello.proto

```

### how to do reverse proxy in code.

```go

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	gw "grpc_demo/gateway_5/gw"

	"net/http"
	"testing"
)


//http
func TestHttpServer(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// grpc服务地址
	endpoint := "127.0.0.1:50052"
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	// HTTP转grpc
	err := gw.RegisterHelloHTTPHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		grpclog.Fatalf("Register handler err:%v\n", err)
	}

	http.ListenAndServe(":8080", mux)
}

```

**IMPORTANT: you also need to create your rpc server! this is just a proxy !**

all code see [gateway_5/](gateway_5)