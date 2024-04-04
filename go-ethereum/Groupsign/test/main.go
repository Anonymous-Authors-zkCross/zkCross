package main

import (
	"fmt"
	zktx "github.com/ethereum/go-ethereum/zktx"
)

func main(){
	fmt.Println(zktx.VerifyHighFeeProof())
}
