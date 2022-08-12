package utilsForTest

type ShosetCreation struct {
	Lname    string
	Stype    string
	Src      string
	Dst      []string
	Ptype    string
	Launched bool
}

var Line2 = []*ShosetCreation{
	{Lname: "A", Stype: "cl", Src: "localhost:8001", Dst: []string{""}, Ptype: "pki", Launched: false},
	{Lname: "B", Stype: "cl", Src: "localhost:8002", Dst: []string{"localhost:8001"}, Ptype: "link", Launched: false},
}

var Line3 = []*ShosetCreation{
	{Lname: "A", Stype: "cl", Src: "localhost:8001", Dst: []string{""}, Ptype: "pki", Launched: false},
	{Lname: "B", Stype: "cl", Src: "localhost:8002", Dst: []string{"localhost:8001"}, Ptype: "link", Launched: false},
	{Lname: "C", Stype: "cl", Src: "localhost:8003", Dst: []string{"localhost:8002"}, Ptype: "link", Launched: false},
}

//Configuration  1 : (straight line)

var StraightLine = []*ShosetCreation{
	{Lname: "A", Stype: "cl", Src: "localhost:8001", Dst: []string{""}, Ptype: "pki", Launched: false},
	{Lname: "B", Stype: "cl", Src: "localhost:8002", Dst: []string{"localhost:8001"}, Ptype: "link", Launched: false},
	{Lname: "C", Stype: "cl", Src: "localhost:8003", Dst: []string{"localhost:8002"}, Ptype: "link", Launched: false},
	{Lname: "D", Stype: "cl", Src: "localhost:8004", Dst: []string{"localhost:8003"}, Ptype: "link", Launched: false},
	{Lname: "E", Stype: "cl", Src: "localhost:8005", Dst: []string{"localhost:8004"}, Ptype: "link", Launched: false},
}

var LnameiIP = map[string]string{
	"A": "localhost:8001",
	"B": "localhost:8002",
	"C": "localhost:8003",
	"D": "localhost:8004",
	"E": "localhost:8005",
	"F": "localhost:8006",
	"G": "localhost:8007",
	"H": "localhost:8008",
	"I": "localhost:8009",
}

//Configuration  2 : (two-linked cricles)

var LinkedCircles = []*ShosetCreation{
	{Lname: "A", Stype: "cl", Src: LnameiIP["A"], Dst: []string{LnameiIP["D"]}, Ptype: "pki", Launched: false},
	{Lname: "B", Stype: "cl", Src: LnameiIP["B"], Dst: []string{LnameiIP["A"]}, Ptype: "link", Launched: false},
	{Lname: "C", Stype: "cl", Src: LnameiIP["C"], Dst: []string{LnameiIP["A"]}, Ptype: "link", Launched: false},
	{Lname: "D", Stype: "cl", Src: LnameiIP["D"], Dst: []string{LnameiIP["B"], LnameiIP["C"]}, Ptype: "link", Launched: false},
	{Lname: "E", Stype: "cl", Src: LnameiIP["E"], Dst: []string{LnameiIP["D"], LnameiIP["F"]}, Ptype: "link", Launched: false},
	{Lname: "F", Stype: "cl", Src: LnameiIP["F"], Dst: []string{LnameiIP["E"]}, Ptype: "link", Launched: false},
	{Lname: "G", Stype: "cl", Src: LnameiIP["G"], Dst: []string{LnameiIP["F"]}, Ptype: "link", Launched: false},
	{Lname: "H", Stype: "cl", Src: LnameiIP["H"], Dst: []string{LnameiIP["F"]}, Ptype: "link", Launched: false},
	{Lname: "I", Stype: "cl", Src: LnameiIP["I"], Dst: []string{LnameiIP["G"], LnameiIP["H"]}, Ptype: "link", Launched: false},
}

//Configuration  3 : (circle)

var Circle = []*ShosetCreation{
	{Lname: "A", Stype: "cl", Src: LnameiIP["A"], Dst: []string{}, Ptype: "pki", Launched: false},
	{Lname: "B", Stype: "cl", Src: LnameiIP["B"], Dst: []string{LnameiIP["A"]}, Ptype: "link", Launched: false},
	{Lname: "C", Stype: "cl", Src: LnameiIP["C"], Dst: []string{LnameiIP["A"]}, Ptype: "link", Launched: false},
	{Lname: "D", Stype: "cl", Src: LnameiIP["D"], Dst: []string{LnameiIP["C"]}, Ptype: "link", Launched: false},
	{Lname: "E", Stype: "cl", Src: LnameiIP["E"], Dst: []string{LnameiIP["D"], LnameiIP["B"]}, Ptype: "link", Launched: false},
}
