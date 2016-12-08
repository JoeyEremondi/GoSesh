package main

import "github.com/JoeyEremondi/GoSesh/mockup"

// ASendToBThenBToC : A sends an int message to B, B sends the message to C
func MakeLoopStub() {

	channelAToB := mockup.Channel{
		Name:        "fromAtoB",
		Source:      "A",
		Destination: "B"}

	message := mockup.MessageType{Type: "int"}
	sendAToB := mockup.Send(channelAToB, message)

	channelBToA := mockup.Channel{
		Name:        "fromBtoA",
		Source:      "B",
		Destination: "A"}

	mockup.CreateStubProgram("ABCExample",
		mockup.Loop("testLoop",
			mockup.Send(channelAToB, mockup.MessageType{Type: "Int"}),
			mockup.Switch(channelBToA,
				mockup.Case("intIsGood",
					mockup.Break(),
				),
				mockup.Case("intIsBad",
					mockup.Continue("testLoop"),
				),
			),
		),
	)
}

func main() {
	MakeLoopStub()
}
