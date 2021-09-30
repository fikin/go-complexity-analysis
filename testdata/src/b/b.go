// Package b is test package
package b

// Comment 1

import "fmt"

// Comment 1

// f5 is func
func f5() { // want "Cyclomatic complexity: 1, Halstead difficulty: 3.750, volume: 39.863"
	const aa = `
AA
BB
`
	fmt.Printf("%s", aa)
}

// f6 is function to be ignored
//complexity:ignore
func func6() { // want "ignored"

}
