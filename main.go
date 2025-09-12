package main

import (
	"GoProject/1_module/payments"
	"fmt"

	"github.com/google/uuid"
)


func main() {

	var family [4]string
	family[0] = "Sorahia"
	family[1] = "Luís Henrique"
	family[2] = "Lisandra"
	family[3] = "Matheus"

	person := []string{"Sorahia", "Luís Henrique", "Lisandra"}

	person = append(person, "Matheus")

	// professions := []string{"developer", "QA", "BA", "Tech Lead"}

	// payments.CreatePerson(person)
	// payments.CreatePerson(professions)

	payments.SumAllValues()

	fmt.Println(uuid.New())

	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}

	for k, v := range person {
		fmt.Println(k, v)
	}


}