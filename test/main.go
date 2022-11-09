package main

import (
	"fmt"
	"math"
)

type ErrorGeneric struct {
	blah  string
	value float64
}

type ErrNegativeSqrt struct {
	ErrorGeneric
	oneoff string
}

func (e ErrNegativeSqrt) Error() string {
	return fmt.Sprintf("error in float - %s - %s - %v", e.blah, e.oneoff, float64(e.value))
}

func Sqrt(x float64) (float64, error) {
	if x < 0 {
		return 0, ErrNegativeSqrt{
			ErrorGeneric: ErrorGeneric{
				blah:  "madhu",
				value: x,
			},
			oneoff: "oneoff",
		}
	}
	return math.Sqrt(x), nil
}

func main() {
	val, err := Sqrt(2)
	if err != nil {
		panic(err)
	}
	fmt.Println("value", val)
	val, err = Sqrt(-2)
	if err != nil {
		panic(err)
	}
	fmt.Println("value -2 - SHOULD NEVER HIT", val)
}
