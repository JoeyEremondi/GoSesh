/**
An example mockup for a 2-party interaction which loops,
with A sending an integer to B until B gives back a label accepting the integer.
*/
package main

import "github.com/JoeyEremondi/GoSesh/mockup"

// ASendToBThenBToC : A sends an int message to B, B sends the message to C
func main() {

	channelAToB := mockup.Channel{
		Name:        "127.0.0.1:24602",
		Source:      "A",
		Destination: "B"}

	channelBToA := mockup.Channel{
		Name:        "127.0.0.1:24601",
		Source:      "B",
		Destination: "A"}

	mockup.CreateStubProgram("loopUntilGood.go", "gen/loopUntilGoodPgm",
		mockup.Loop("testLoop",
			mockup.Send(channelAToB, mockup.MessageType{Type: "int"}),
			mockup.Switch(channelBToA,
				mockup.Case("intIsBad",
					mockup.Continue("testLoop"),
				),
				mockup.Case("intIsGood",
					mockup.Break(),
				),
			),
		),
	)
}
