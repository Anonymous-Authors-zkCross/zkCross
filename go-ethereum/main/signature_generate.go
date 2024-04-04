package main

import (
    "fmt"
    "log"
   // "strconv"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/crypto"
)

func main() {
    privateKey, err := crypto.HexToECDSA("f500d9e7647788765f8066efdb4a3b943a2c28b5c8fcc76ddcdb0f7b5ea77f8f")
    if err != nil {
        log.Fatal(err)
    }

    data := []byte("1")
    hash := crypto.Keccak256Hash(data)
    fmt.Println(hash.Hex()) // 0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8

    signature, err := crypto.Sign(hash.Bytes(), privateKey)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("x的类型是%T",signature)
    fmt.Println(signature)
    sig := hexutil.Encode(signature)
    fmt.Println(sig)
    var data1 []byte = []byte(sig)
    fmt.Println(data1)
}
