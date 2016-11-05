package multiparty

import (
	"math/rand"
)

// Request example
type Request struct {
	Quote int
}

// Response example
type Response struct {
	Quote, Order, Delivery int
}

// Buyer example type
type Buyer int

// Seller example type
type Seller int

// AskSeller to send some request type to the buyer
func (s *Seller) AskSeller(requestType *Request, reply *int) error {
	// random selling price
	*reply = rand.New(rand.NewSource(1000)).Int()
	return nil
}

// Receive a quote from the seller and randomly decide to accept it
/*func (b *Buyer) DecideQuote(req *Request, reply *int) error {
	if req.Quote < rand.New(rand.NewSource(100)).Int() {
		*reply = 1
	} else {
		*reply = 0
	}
	return nil
}*/
