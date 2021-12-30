package main

import (
	"fmt"
	zipkintrace "github.com/asim/go-micro/examples/v4/wrapper/trace"
	//"google.golang.org/grpc/profiling/service"
	"log"

	"context"
	proto "github.com/asim/go-micro/examples/v4/service/proto"
	"go-micro.dev/v4"
	"go-micro.dev/v4/server"

)

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *proto.Request, rsp *proto.Response) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}

// logWrapper is a handler wrapper
func logWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		log.Printf("[wrapper] server request: %v", req.Endpoint())
		err := fn(ctx, req, rsp)
		return err
	}
}


func main() {
	serverName:="go.micro.srv.wrapper"
	hostPort:="localhost:6008"

	zipkintrace.InitZipKinTrace(serverName,hostPort)

	service := micro.NewService(
		micro.Name(serverName),
		// wrap the handler
		micro.WrapHandler(logWrapper, zipkintrace.TraceWrappers),
	)

	service.Init()

	proto.RegisterGreeterHandler(service.Server(), new(Greeter))

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
