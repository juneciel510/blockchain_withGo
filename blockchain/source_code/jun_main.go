package main

import "os"


func main() {
	
	deadsig := make(chan os.Signal, 1)
	users:=NewUsers()
	go ClientLoop(users)
	go users.HandleChannel()
	<-deadsig



}