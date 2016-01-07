package beat

type BlkioCalculator interface {
	getRead() uint64
	getWrite() uint64
	getTotal() uint64
}

type BlkioCalculatorImpl struct {
	old BlkioData
	new BlkioData
}

type BlkioData struct {
	reads  uint64
	writes uint64
	totals uint64
}

func (c BlkioCalculatorImpl) getRead() uint64 {
	return c.new.reads - c.old.reads
}

func (c BlkioCalculatorImpl) getWrite() uint64 {
	return c.new.writes - c.old.writes
}

func (c BlkioCalculatorImpl) getTotal() uint64 {
	return c.new.totals - c.old.totals
}
