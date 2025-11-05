package main

import configs "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/config"

func main() {
	config, err := configs.LoadConfig(".") // (.) path do .env - diretório atual
	if err != nil {
		println(err)
	}
	println(config.DBDriver)
}
