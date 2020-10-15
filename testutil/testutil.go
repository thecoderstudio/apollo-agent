package testutil

import "strings"

// BlockUntilContains will block and continuously read the channel until the subString is
// found in the TOTAL output of the channel. Transform is used to transform a singular channel
// output to a string.
func BlockUntilContains(
	channel <-chan interface{},
	transform func(interface{}) string,
	subString string,
) {
	readOutput := ""
	for {
		output := <-channel
		readOutput = readOutput + transform(output)
		if strings.ContainsAny(readOutput, subString) {
			break
		}
	}
}
