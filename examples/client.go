package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	pb "github.com/huaishan/jserver/grpc"
	"google.golang.org/grpc"
)

const (
	address            = "localhost:50051"
	base64EncodeScript = `
var keyStr = "ABCDEFGHIJKLMNOP" +
	"QRSTUVWXYZabcdef" +
	"ghijklmnopqrstuv" +
	"wxyz0123456789+/" +
	"=";
function base64_encode(input) {
	var output = "";
	var chr1, chr2, chr3 = "";
	var enc1, enc2, enc3, enc4 = "";
	var i = 0;
	do {
		chr1 = input.charCodeAt(i++);
		chr2 = input.charCodeAt(i++);
		chr3 = input.charCodeAt(i++);
		enc1 = chr1 >> 2;
		enc2 = ((chr1 & 3) << 4) | (chr2 >> 4);
		enc3 = ((chr2 & 15) << 2) | (chr3 >> 6);
		enc4 = chr3 & 63;
		if (isNaN(chr2)) {
			enc3 = enc4 = 64;
		} else if (isNaN(chr3)) {
			enc4 = 64;
		}
		output = output +
		keyStr.charAt(enc1) +
		keyStr.charAt(enc2) +
		keyStr.charAt(enc3) +
		keyStr.charAt(enc4);
		chr1 = chr2 = chr3 = "";
		enc1 = enc2 = enc3 = enc4 = "";
	} while (i < input.length);
	return output;
};
`
	SCRIPT = `
vn
var i = 0;
var a = {};
for (;;) {
	a[i] = i;
	i++;
}
`
)

func testCompile(c pb.JserverClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Compile(ctx, &pb.CompileRequest{Script: SCRIPT})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Code: %d, Resp: %s\n", r.GetCode(), r.GetMessage())
}

type JsonParams struct {
	Params []interface{} `json:"params"`
}

func testCallFunc(c pb.JserverClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	jp := JsonParams{Params: []interface{}{"我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体我能吞下玻璃而不伤身体"}}
	ps, _ := json.Marshal(jp)
	log.Println(string(ps))

	r, err := c.CallFunc(ctx, &pb.CallFuncRequest{
		Script:   base64EncodeScript,
		FuncName: "base64_encode",
		Params:   ps,
		Timeout:  100,
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Code: %d, Resp: %s, TimeConsuming: %s\n", r.GetCode(), r.GetMessage(), r.GetTimeConsuming())
	if r.GetCode() == pb.StatusCode_OK {
		log.Println(string(r.GetResult()))
	}
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewJserverClient(conn)

	// Contact the server and print out its response.
	testCompile(c)
	testCallFunc(c)
}
