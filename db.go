package bitcoin_go

import (
	"bytes"
	"encoding/gob"
)

func (this *Block) Serialize() []byte {
    var result bytes.Buffer
    encoder := gob.NewEncoder(&result)

    if err := encoder.Encode(this); nil != err {
        panic(err)
    }

    return result.Bytes()
}

func DeSerialize(data []byte) *Block {
    var block Block

    decoder := gob.NewDecoder(bytes.NewReader(data))
    if err := decoder.Decode(&block); nil != err {
        panic(err)
    }

    return &block
}
