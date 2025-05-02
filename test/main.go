package main

type Teststruct struct {
	FieldsAny any
	PointerField **DummyStruct
	StringField string
	SliceField []string
	SliceFieldAdvanced []DummyStruct
	PointerSlice []*DummyStruct
	WrdPointerSlice []any
	Dumm AnotherDumm
}

type AnotherDumm struct {
	Dumm DummyStruct
}

type DummyStruct struct {
	TestVal string
	SecondVal string
}

func main() {
}


func waltest() {
	
}


