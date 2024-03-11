package main

import (
	"fmt"
	"github.com/olivercullimore/hikvision-anpr-alerts/app"
)

func main() {
	// Run server
	fmt.Println("Hikvision ANPR Alerts")
	fmt.Println("-----------------------------------------------------------------------------")
	app.Run()
}
