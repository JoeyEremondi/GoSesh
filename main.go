package main

import "sessions/multiparty"

func main() {
	t := multiparty.LocalSendType{channel: "foo", value: "foo", next: multiparty.LocalEndType{}}
	println(t.stub())
}
