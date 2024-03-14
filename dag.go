package merkledag

import (
	"bytes"
	"hash"
	"math"
)

const (
	Blob = 256 * 1024
)

// 生成MerkleDAG的功能
func Add(store KVStore, node Node, h hash.Hash) ([]byte, error) {
	// 根据节点类型执行相应的操作
	switch node.Type() {
	case FILE:
		// 将节点转换为File接口类型
		file := node.(File)
		// 获取文件内容
		fileContent := file.Bytes()

		// 计算分片的数量
		numBlob := int(math.Ceil(float64(len(fileContent)) / float64(Blob)))
		childLinks := make([]Link, numBlob)

		// 分片计算哈希值
		for i := 0; i < numBlob; i++ {
			start := i * Blob
			end := int(math.Min(float64(start+Blob), float64(len(fileContent))))
			chunk := fileContent[start:end] // 分片内容

			h.Reset()
			h.Write(chunk)
			hashValue := h.Sum(nil)

			// 将分片的哈希值存储在KVStore中
			store.Put(hashValue, chunk)

			childLinks[i] = Link{Hash: hashValue, Name: ""}
		}

		// 计算文件的Merkle Root
		fileHash := h.Sum(bytes.Join(childLinks.ToBytes(), nil))
		store.Put(fileHash, nil)

		node.SetData("blob")

		return fileHash, nil

	case DIR:
		// 将节点转换为Dir接口类型
		dir := node.(Dir)

		// 创建一个空的链接数组
		childLinks := make([]Link, 0)

		// 获取文件夹迭代器
		iterator := dir.It()

		// 遍历文件夹下的文件/文件夹
		for iterator.Next() {
			child := iterator.Node()

			// 递归调用Add函数生成子节点的Merkle Root，并将其添加到链接数组中
			childHash, err := Add(store, child, h)
			if err != nil {
				return nil, err
			}

			childLinks = append(childLinks, Link{Hash: childHash, Name: child.Name()})
		}

		// 计算链接数组的哈希值作为文件夹的Merkle Root
		dirHash := h.Sum(bytes.Join(childLinks.ToBytes(), nil))

		// 将文件夹的Merkle Root写入KVStore
		store.Put(dirHash, nil)

		node.SetData("tree")

		return dirHash, nil
	}

	return nil, nil
}

