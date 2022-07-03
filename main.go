package main

import (
	r "github.com/io-boxies/io-app-engine/router"
)

func main() {
	router := r.InitRoutes()
	router.Run()
}
