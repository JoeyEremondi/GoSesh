package main

import "github.com/JoeyEremondi/GoSesh/mockup"

func main() {

	channelAToB := mockup.Channel{
		Name:        "127.0.0.1:24602",
		Source:      "A",
		Destination: "B"}

	channelBToA := mockup.Channel{
		Name:        "127.0.0.1:24601",
		Source:      "B",
		Destination: "A"}

	channelAToC := mockup.Channel{
		Name:        "127.0.0.1:24603",
		Source:      "A",
		Destination: "C"}

	channelCToA := mockup.Channel{
		Name:        "127.0.0.1:24601",
		Source:      "C",
		Destination: "A"}

	mockup.CreateStubProgram("2pc.go", "gen/2pc",
		mockup.Send(channelAToB, mockup.MessageType{Type: "string"}),
		mockup.Switch(channelBToA,
			mockup.Case("B-Fail",
				mockup.Send(channelAToC, mockup.MessageType{Type: "string"}),
				mockup.Switch(channelCToA,
					mockup.Case("C-Fail",
						mockup.Send(channelAToB, mockup.MessageType{Type: "string"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "string"}),
					),
					mockup.Case("C-Commit",
						mockup.Send(channelAToB, mockup.MessageType{Type: "string"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "string"}),
					),
				),
			),
			mockup.Case("B-Commit",
				mockup.Send(channelAToC, mockup.MessageType{Type: "string"}),
				mockup.Switch(channelCToA,
					mockup.Case("C-Fail",
						mockup.Send(channelAToB, mockup.MessageType{Type: "string"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "string"})),
					mockup.Case("C-Commit",
						mockup.Send(channelAToB, mockup.MessageType{Type: "string"}),
						mockup.Send(channelAToC, mockup.MessageType{Type: "string"}),
					),
				),
			),
		),
	)
}
