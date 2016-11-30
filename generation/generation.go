package generation

import "bytes"

// Channel is a struct for the developer to use
type Channel struct {
	Name        string
	Source      string
	Destination string
}

// ValueType is a struct for the developer to use
type ValueType struct {
	Type string // instead of Sort which is non-intuitive
}

// LocalType is a struct for the developer to use
type LocalType struct {
	// TODO What forms can this take on? Does this need some interface?
	Anything string
}

//LocalSendType : perform unwrapping for the developer
func LocalSendType(channel Channel, valueType ValueType, next LocalType) string {
	return channel.Source + " --> " + channel.Destination + " : " + channel.Name
}

//LocalReceiveType : perform unwrapping for the developer
func LocalReceiveType(channel Channel, valueType ValueType, next LocalType) string {
	return channel.Destination + " --> " + channel.Source + " : " + channel.Name
}

//LocalBranchingType : perform unwrapping for the developer
func LocalBranchingType(channel Channel, branches map[string]LocalType) string {
	var buffer bytes.Buffer
	var i int
	// TODO
	for _, v := range branches {
		i++
		buffer.WriteString(v.Anything)
		if i != len(branches) {
			buffer.WriteString("\n + \n ")
		}
	}

	return buffer.String()
}

//LocalSelectionType : perform unwrapping for the developer
func LocalSelectionType(channel Channel, branches map[string]LocalType) string {
	var buffer bytes.Buffer
	// TODO
	for _, v := range branches {
		buffer.WriteString(v.Anything)
	}

	return buffer.String()
}

// TODO are we implementing the LocalRecursiveType type?
// TODO implicitly end the operations for the dev LocalEndType
