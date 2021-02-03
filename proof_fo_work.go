// @Title 工作量证明
// @Description 工作量证明算法实现
// @Author shouchao.zheng 2021-01-23
// @Update shouchao.zheng 2021-01-23
package bc

import (
    "bytes"
    "crypto/sha256"
    "encoding/binary"
    "fmt"
    "math"
    "math/big"
)

const (
	targetBits = 24 // 算出来的hash开头必须是24个0（以二进制来计算的）
    maxNonce = math.MaxInt64 // 这个上限可真够大的，大概是 2^63 -1
)

type ProofOfWork struct {
    block  *Block
    target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
    target := big.NewInt(1)
    target.Lsh(target, uint(256-targetBits)) // 左移 (256 - targetBits) 个比特位

    return &ProofOfWork{
        target: target,
        block:  block,
    }
}

// 准备数据
// nonce 在这里做计数器解释
func (this *ProofOfWork) prepareData(nonce int) []byte {
    data := bytes.Join([][]byte{
        this.block.PrevBlockHash,
        this.block.HashTranscations(),
        IntToHex(this.block.Timestamp),
        IntToHex(int64(targetBits)),
        IntToHex(int64(nonce)),
    }, []byte{})
    return data
}

// 核心算法
func (this *ProofOfWork) Run() (int, []byte) {
    var (
    	hashInt big.Int
    	hash [32]byte
    	nonce = 0
    )

    //fmt.Printf("Mining the block containning \"%s\"\n", this.block.Data)
    for nonce < maxNonce {
        hash = sha256.Sum256(this.prepareData(nonce))
        hashInt.SetBytes(hash[:]) // 将hash结果转换成一个大整数

        if hashInt.Cmp(this.target) == -1 { // -1代表小于
            // 找到小于目标上界的值了，工作量证明结束
            fmt.Printf("\r%x", hash)
            break
        } else {
            // 计算结果大于目标上界，继续寻找
            nonce++
        }
    }
    fmt.Println("\n\n")

    return nonce, hash[:]
}

// 对工作量证明进行验证
func (this *ProofOfWork) Validate() bool {
    var hashInt big.Int

    data := this.prepareData(this.block.Nonce)
    hash := sha256.Sum256(data)
    hashInt.SetBytes(hash[:])

    return hashInt.Cmp(this.target) == -1
}

// 将一个 int64 转化为一个字节数组(byte array)
func IntToHex(num int64) []byte {
    buff := new(bytes.Buffer)
    err := binary.Write(buff, binary.BigEndian, num)
    if err != nil {
        panic(err)
    }

    return buff.Bytes()
}

func TestPow() {
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
}
