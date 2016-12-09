package main

import "github.com/JoeyEremondi/GoSesh/mockup"

// ASendToBThenBToC : A sends an int message to B, B sends the message to C
func main() {

	channelAToB := mockup.Channel{
		Name:        "fromAtoB",
		Source:      "A",
		Destination: "B"}

	channelBToA := mockup.Channel{
		Name:        "fromBtoA",
		Source:      "B",
		Destination: "A"}

	mockup.CreateStubProgram("loopUntilGood.go", "gen/ABCExample",
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
