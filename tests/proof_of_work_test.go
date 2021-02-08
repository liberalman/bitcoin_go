package test

import (
    // system package
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "math/big"
    "testing"

    // third package
    "github.com/stretchr/testify/assert"

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
    data1Hash := sha256.Sum256(data1)
    data2Hash := sha256.Sum256(data2)

    assert.Equal(t, "f80867f6efd4484c23b0e7184e53fe4af6ab49b97f5293fcd50d5b2bfa73a4d0",
        hex.EncodeToString(data1Hash[:]))
    fmt.Printf("%x\n", data1Hash)
    assert.Equal(t, "010000000000000000000000000000000000000000000000000000000000",
        hex.EncodeToString(target.Bytes()))
    fmt.Printf("%64x\n", target) // 注意这是个大整数，打印的时候1前面是没有0的
    assert.Equal(t, "0000002f7c1fe31cb82acdc082cfec47620b7e4ab94f2bf9e096c436fc8cee06",
        hex.EncodeToString(data2Hash[:]))
    fmt.Printf("%x\n", data2Hash)
}

