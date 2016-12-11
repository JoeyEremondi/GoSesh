package main

import (
	"fmt"

	"github.com/JoeyEremondi/GoSesh/multiparty"
)

//import "sessions/multiparty" TODO this doesn't work in my (Jodi's) environment (OSX El capitan)

func main() {
	t := multiparty.LocalSendType{Channel: "foo", Value: "foo", Next: multiparty.LocalEndType{}}
	fmt.Println(t)
	//println(t.stub()) TODO this doesn't compile
}
