package theory

func b() {
	panic("hello")
}

func TheoryStart() (value int) {
	value = 12

	defer func() int {
		if err := recover(); err != nil {
			value = value * 2
		}

		return -1
	}()

	b()

	return value
}
