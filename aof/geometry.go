package aof

var (
	OpenFile   = Geometry{}
	CreateFile = Geometry{Create: true}
)

var (
	SizeNowDefault    = pageSize
	SizeUpperDefault  = int64(1024 * 1024 * 16) // 16MB
	GrowthStepDefault = pageSize
)

type Geometry struct {
	SizeNow    int64
	SizeUpper  int64
	GrowthStep int64
	PageSize   int64
	Create     bool
}

func (g *Geometry) Validate() {
	if g.SizeNow < pageSize {
		g.SizeNow = pageSize
	}
	g.SizeNow = alignToPageSize(g.SizeNow)
	if g.SizeUpper <= 0 {
		g.SizeUpper = SizeUpperDefault
	}
	g.SizeUpper = alignToPageSize(g.SizeUpper)
	if g.SizeUpper < g.SizeNow {

	}
	if g.GrowthStep == 0 {
		g.GrowthStep = 1024 * 1024
	}
	g.GrowthStep = alignToPageSize(g.GrowthStep)
	if g.SizeUpper < g.SizeNow {
		g.SizeUpper = g.SizeNow
	}
	g.PageSize = pageSize
}

func (g *Geometry) Next(size int64) int64 {
	if size < 0 {
		size = 0
	}
	if g.GrowthStep <= 0 {
		g.Validate()
	}
	if size < g.SizeNow {
		return g.SizeNow
	}
	// Add the remaining in current step and add another step
	next := size + (size % g.GrowthStep) + g.GrowthStep
	if g.SizeUpper < next {
		return g.SizeUpper
	}
	return next
}
