package test

import (
    //"encoding/hex"
    //"strings"
    "crypto/sha256"
    "fmt"
    "math/big"
    "testing"

    //"github.com/stretchr/testify/assert"

    // item package
)

func TestPow(t *testing.T) {
    var nonce int = 13240266 // 假设计数器的当前值是 13240266
    var nonceStr string = fmt.Sprintf("%x", nonce) // 转换为对应的16进制值 "ca07ca"
    data1 := []byte("I like donuts")
    data2 := []byte("I like donuts" + nonceStr) // 追加计数器当前值
    targetBits := 24 // 要求前24位为0（二进制值）
    target := big.NewInt(1)
    target.Lsh(target, uint(256-targetBits)) // 左移(256-targetBits)位
    fmt.Printf("%x\n", sha256.Sum256(data1))
    fmt.Printf("%64x\n", target)
    fmt.Printf("%x\n", sha256.Sum256(data2))

/*    rawHash := "00010966776006953D5567439E5E39F86A0D273BEED61967F6"
    hash, err := hex.DecodeString(rawHash)
    if nil != err {
        t.Fatal(err)
    }

    encoded := Base58Encode(hash)
    assert.Equal(t, "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM", string(encoded))

    decoded := Base58Decode([]byte("16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM"))
    assert.Equal(t, strings.ToLower("00010966776006953D5567439E5E39F86A0D273BEED61967F6"), hex.EncodeToString(decoded))
*/
}

