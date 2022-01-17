package kernel

import "fmt"

func GetKeyHeight() []byte {
	return []byte(KeyHeight)
}

func GetKeyBlock(height Height) []byte {
	return []byte(fmt.Sprintf(KeyBlock, height))
}

func GetKeyTX(hash Hash) []byte {
	return []byte(fmt.Sprintf(KeyTX, hash))
}

func GetKeyMempoolHeight() []byte {
	return []byte(KeyMempoolHeight)
}

func GetKeyMempoolTX(hash Hash) []byte {
	return []byte(fmt.Sprintf(KeyMempoolTX, hash))
}
