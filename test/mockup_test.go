package test

/* Run a single test file with verbose comments like:
 * go test -v mockup_test.go
 *
 * Test the mockup functions by logging their output
 */

import (
	"testing"

	"github.com/JoeyEremondi/GoSesh/mockup"
)

func TestSend(test *testing.T) {
	channel := mockup.Channel{
		Name:        "testchannel",
		Source:      "A",
		Destination: "B"}

	message := mockup.MessageType{Type: "int"}

	send := mockup.Send(channel, message)
	test.Log(send)
}

/*
func TestLocalBranchingType(test *testing.T) {
	channel := mockup.Channel{Name: "testchannel"}
	event1 := mockup.Event{Event: "Event1"}
	event2 := mockup.Event{Event: "Event2"}
	event3 := mockup.Event{Event: "Event3"}
	// TODO what is this key supposed to be used for?

	EventMap := map[string]mockup.Event{
		"1": event1,
		"2": event2,
		"3": event3}
	localBranchingType := mockup.Branch(channel, EventMap)
	test.Log(localBranchingType)
}*/
