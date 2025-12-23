package mathx

import "math/bits"

func FastLog2(x int) int {
	// 边界处理：非正整数直接返回最小值 n=0（2^0=1）
	if x <= 0 {
		return 0 // 2^0=1 是所有非正整数的解
	}

	// 将 int 转为 uint（确保位操作安全）
	ux := uint(x)
	ux-- // 处理 x 恰好是 2^k 的情况

	// 位填充：将最高位以下所有低位设为 1
	ux |= ux >> 1
	ux |= ux >> 2
	ux |= ux >> 4
	ux |= ux >> 8
	ux |= ux >> 16
	if bits.UintSize == 64 { // 自动适配 64 位系统
		ux |= ux >> 32
	}

	// 计算位数：填充后数值长度即为 n+1
	return bits.Len(ux)
}
