package main

import (
	"fmt"
	"errors"
)
func main() {
	err := haha()
	fmt.Println(err)
}

func haha()  error {
	err := test()
	name := "testhhhh"
	return errors.New("can not found storagecluster for volumeattachment: " + name + err.Error())
}

func test() error{
	return errors.New("can ")
}
