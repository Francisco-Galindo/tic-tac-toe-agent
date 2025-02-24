package main

import "fmt"

func main() {
	var num uint16 = 0b0000000000000000
	fmt.Println(^num, num)

	var a uint16 = 0b1111111000000000
	var b uint16 = 0b0000000111111111
	fmt.Printf("x = %b\n",a)
	fmt.Printf("y = %b\n",b)

	fmt.Printf("x ^ y is %b\n",a ^ (^b))
	for i := range(4) {
		fmt.Println(i)
	}

}
