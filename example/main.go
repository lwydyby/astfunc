package main

import (
	"fmt"

	"github.com/lwydyby/astfunc"
)

func main() {
	result := astfunc.GetFuncWithDependency(".", "A")
	fmt.Println(result)
}
