package beat

type CalculatorFactory struct {
}

func (c *CalculatorFactory) newBlkioCalculator(old BlkioData, new BlkioData) BlkioCalculator {
	return BlkioCalculator{old, new}
}

func (c *CalculatorFactory) newCPUCalculator(old CPUData, new CPUData) CPUCalculator {
	return CPUCalculator{old, new}
}

func (c *CalculatorFactory) newNetworkCalculator(old NetworkData, new NetworkData) NetworkCalculator {
	return NetworkCalculator{old, new}
}
