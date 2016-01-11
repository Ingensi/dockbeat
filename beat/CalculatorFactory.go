package beat

type CalculatorFactory interface {
	newBlkioCalculator(old BlkioData, new BlkioData) BlkioCalculator
	newCPUCalculator(old CPUData, new CPUData) CPUCalculator
	newNetworkCalculator(old NetworkData, new NetworkData) NetworkCalculator
}

type CalculatorFactoryImpl struct {
}

func (c CalculatorFactoryImpl) newBlkioCalculator(old BlkioData, new BlkioData) BlkioCalculator {
	return BlkioCalculatorImpl{old: old, new: new}
}

func (c CalculatorFactoryImpl) newCPUCalculator(old CPUData, new CPUData) CPUCalculator {
	return CPUCalculatorImpl{old: old, new: new}
}

func (c CalculatorFactoryImpl) newNetworkCalculator(old NetworkData, new NetworkData) NetworkCalculator {
	return NetworkCalculatorImpl{old: old, new: new}
}
