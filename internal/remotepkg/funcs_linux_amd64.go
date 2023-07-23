package remotepkg

func Func(b Bool, r Rune, i Int) error {
	return nil
}

type Method struct{}

func (Method) Method(b Bool, r Rune, i Int) error {
	return nil
}
