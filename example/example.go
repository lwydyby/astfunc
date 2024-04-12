package main

func A(a string) string {
	return B()
}

func B() string {
	return "test"
}
