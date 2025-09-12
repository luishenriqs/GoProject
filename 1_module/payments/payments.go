package payments

import (
	"GoProject/1_module/payments/math"
	"fmt"
)

func Print(value any) {
	fmt.Println(value)
}



func CreatePerson(value []string) {

	for _, v := range value {
		Print(v)
	}

	fmt.Printf("%v\n", value[1:])
}


func SumAllValues() {
	s := math.Sum(10, 20)
	println("O resultado é: ", s)
}