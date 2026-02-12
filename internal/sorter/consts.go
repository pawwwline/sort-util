package sorter

type months int

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
