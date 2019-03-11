package commons

type Error string

func (this Error) Error() string {
	return string(this)
}
