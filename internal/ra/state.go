package ra

type state int

const (
	RELEASED state = iota
	WANTED
	HELD
)

func (s state) String() string {
	switch s {
	case RELEASED:
		return "RELEASED"
	case WANTED:
		return "WANTED"
	case HELD:
		return "HELD"
	default:
		return "?"
	}
}
