package main

import (
	"github.com/henrylee2cn/faygo"
	"user/router"
)

func main() {
	router.Route(faygo.New("user"))
	faygo.Run()
}
