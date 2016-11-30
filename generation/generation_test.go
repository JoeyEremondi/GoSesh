package generation

import (
	"testing"

	"github.com/joey/sessions/generation"
)

// Run a single test file with verbose comments like:
// go test -v generation_test.go
func TestLocalSendType(test *testing.T) {
	channel := generation.Channel{
		Name:        "testchannel",
		Source:      "A",
		Destination: "B"}
	valueType := generation.ValueType{Type: "int"}
	localType := generation.LocalType{Anything: "nexttype"}

	localSendType := generation.LocalSendType(channel, valueType, localType)
	test.Log(localSendType)
}

func TestLocalReceiveType(test *testing.T) {
	channel := generation.Channel{
		Name:        "testchannel",
		Source:      "A",
		Destination: "B"}
	valueType := generation.ValueType{Type: "int"}
	localType := generation.LocalType{Anything: "nexttype"}

	localReceiveType := generation.LocalReceiveType(channel, valueType, localType)
	test.Log(localReceiveType)
}

func TestLocalBranchingType(test *testing.T) {
	channel := generation.Channel{Name: "testchannel"}
	localType1 := generation.LocalType{Anything: "localtype1"}
	localType2 := generation.LocalType{Anything: "localtype2"}
	localType3 := generation.LocalType{Anything: "localtype3"}
	// TODO what is this key supposed to be used for?

	localTypeMap := map[string]generation.LocalType{
		"1": localType1,
		"2": localType2,
		"3": localType3}
	localBranchingType := generation.LocalBranchingType(channel, localTypeMap)
	test.Log(localBranchingType)
}

// TODO test selection type
