package a

import "fmt"

func f0() { // want "Cyclomatic complexity: 1"
}

func f1() { // want "Cyclomatic complexity: 3"
	if false {

	} else {

	}
}

func f2() { // want "Cyclomatic complexity: 8"
	for true {
		if false {

		} else if false {

		} else if false {

		} else if false {
			n := 0
			switch n {
			case 0:
			case 1:
			default:
			}
		} else {

		}
	}
}

func f3() { // want "Cyclomatic complexity: 4"
	if false || true {
		if false {

		}
	}
}

func f4() { // want "Cyclomatic complexity: 2"
	n := 0
	switch n {
	case 0:
	case 1:
	case 2:
	case 3:
	case 4:
	case 5:
	case 6:
	case 7:
	case 8:
	case 9:
	}
}

// f5 is func
func f5() { // want "Cyclomatic complexity: 1, Halstead difficulty: 3.750, volume: 39.863"
	const aa = `
AA
BB
`
	fmt.Printf("%s", aa)
}
