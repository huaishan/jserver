package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"github.com/dop251/goja"
	pb "github.com/huaishan/jserver/grpc"
	json "github.com/json-iterator/go"
	"google.golang.org/grpc"
)

var port = flag.String("port", ":50051", "jserver port")

type server struct {
	pb.UnimplementedJserverServer
}

func (s *server) Compile(ctx context.Context, in *pb.CompileRequest) (*pb.CompileResponse, error) {
	vm := goja.New()
	_, err := vm.RunString(in.Script)
	if err != nil {
		return &pb.CompileResponse{
			Code:    pb.StatusCode_Failed,
			Message: err.Error(),
		}, nil
	}

	return &pb.CompileResponse{
		Code:    pb.StatusCode_OK,
		Message: pb.StatusCode_OK.String(),
	}, nil
}

type CallFuncParams struct {
	Params []interface{} `json:"params"`
}

type CallFuncResponse struct {
	Return interface{} `json:"return"`
}

func (s *server) CallFunc(ctx context.Context, in *pb.CallFuncRequest) (*pb.CallFuncResponse, error) {
	var code = pb.StatusCode_OK
	var message = pb.StatusCode_OK.String()
	var ps CallFuncParams
	var ok bool
	var f goja.Callable
	var vals []goja.Value

	var timeout int32 = 50
	if in.Timeout > 0 {
		timeout = in.Timeout
	}

	now := time.Now()

	vm := goja.New()
	time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
		vm.Interrupt("timeout")
	})

	_, err := vm.RunString(in.Script)
	if err != nil {
		code = pb.StatusCode_Failed
		message = err.Error()
		goto Error
	}

	err = json.Unmarshal(in.Params, &ps)
	if err != nil {
		code = pb.StatusCode_Failed
		message = "params malformed"
		goto Error
	}

	for _, v := range ps.Params {
		vals = append(vals, vm.ToValue(v))
	}

	if f, ok = goja.AssertFunction(vm.Get(in.FuncName)); ok {
		retVal, err := f(goja.Undefined(), vals...)
		if err != nil {
			code = pb.StatusCode_Failed
			message = err.Error()
			goto Error
		}

		ret := CallFuncResponse{Return: retVal.Export()}
		jsonRet, err := json.Marshal(ret)

		return &pb.CallFuncResponse{
			Code:          code,
			Message:       message,
			Result:        jsonRet,
			TimeConsuming: time.Since(now).String(),
		}, nil
	} else {
		code = pb.StatusCode_Failed
		message = "func_name <" + in.FuncName + "> is not found"
		goto Error
	}

Error:
	return &pb.CallFuncResponse{
		Code:          code,
		Message:       message,
		TimeConsuming: time.Since(now).String(),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterJserverServer(s, &server{})
	log.Println("Listen " + *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
