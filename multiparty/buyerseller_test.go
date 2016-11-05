package multiparty

import (
	"net"
	"net/http"
	"net/rpc"
	"testing"

	"github.com/joey/sessions/multiparty"
)

// Run the single test with comments like:
// go test -v buyerseller_test.go
func TestQuote(test *testing.T) {
	// Create a server connection to the Seller
	seller := new(multiparty.Seller)
	rpc.Register(seller)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		panic(err)
	}
	go http.Serve(l, nil)
	// Create a client connection from Buyer to the Seller server
	buyer, err := rpc.DialHTTP("tcp", "127.0.0.1:1234")
	if err != nil {
		panic(err)
	}
	test.Log("Connected buyer")
	// Make a synchronous call to get a quote from the seller
	request := &multiparty.Request{0}
	var reply int
	test.Log("Call seller for a quote")
	err = buyer.Call("Seller.AskSeller", request, &reply)
	if err != nil {
		panic(err)
	}
	test.Log("Reply ", reply)
}
