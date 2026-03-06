package sorter

type months int32

const (
	_                = iota
	january   months = iota
	february  months = iota
	march     months = iota
	april     months = iota
	may       months = iota
	june      months = iota
	july      months = iota
	august    months = iota
	september months = iota
	october   months = iota
	november  months = iota
	december  months = iota
)

const step = 10

const (
	_   = iota
	kiB = 1 << (step * iota)
	miB = 1 << (step * iota)
	giB = 1 << (step * iota)
	tiB = 1 << (step * iota)
)

const minMonthLength = 3

const defaultMaxLineSize = 10 << 20

// memoryThreshold is the amount of data (in bytes) accumulated before switching to disk-based sorting.
const memoryThreshold = 256 * miB

// scannerInitBufSize is the initial capacity of the bufio.Scanner buffer.
const scannerInitBufSize = 64 * 1024

// minDeduplicateLen is the minimum slice length that requires deduplication.
const minDeduplicateLen = 2
