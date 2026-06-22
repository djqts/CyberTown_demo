package memory

import (
	"hash/fnv"
	"math"
)

const EmbedDim = 768

// Embed 将文本转为 768 维向量（MVP 用确定性 hash，后续可替换为 DeepSeek Embedding API）。
func Embed(text string) []float32 {
	h := fnv.New128a()
	h.Write([]byte(text))
	b := h.Sum(nil)

	vec := make([]float32, EmbedDim)
	for i := range vec {
		// 每 2 字节生成一个 float32，超出时复用
		idx := i * 2 % len(b)
		if idx+1 >= len(b) {
			idx = 0
		}
		val := float32(uint16(b[idx])<<8|uint16(b[idx+1])) / 65535.0
		vec[i] = val*2 - 1 // [-1, 1]
	}

	// L2 归一化，便于余弦相似度（Dot = Cosine when normalized）
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}
