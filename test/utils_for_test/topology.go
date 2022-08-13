package utilsForTest

// List of instruction for launching a shoset.
type ShosetCreation struct {
	Lname           string
	ShosetType      string
	LocalAddress    string
	RemoteAddresses []string
	ProtocolType    string // type of connection (pki (master), link, join, ...)
	Launched        bool   // true when the shoset has been launched
	//Used to extend the list after the first launch by calling again CreateManyShosets or CreateShosetFromTopology.
}

// #### List of topologies

var Line2 = []*ShosetCreation{
	{Lname: "A", ShosetType: "cl", LocalAddress: "localhost:8001", RemoteAddresses: []string{""}, ProtocolType: "pki", Launched: false},
	{Lname: "B", ShosetType: "cl", LocalAddress: "localhost:8002", RemoteAddresses: []string{"localhost:8001"}, ProtocolType: "link", Launched: false},
}

var Line3 = []*ShosetCreation{
	{Lname: "A", ShosetType: "cl", LocalAddress: "localhost:8001", RemoteAddresses: []string{""}, ProtocolType: "pki", Launched: false},
	{Lname: "B", ShosetType: "cl", LocalAddress: "localhost:8002", RemoteAddresses: []string{"localhost:8001"}, ProtocolType: "link", Launched: false},
	{Lname: "C", ShosetType: "cl", LocalAddress: "localhost:8003", RemoteAddresses: []string{"localhost:8002"}, ProtocolType: "link", Launched: false},
}

var StraightLine = []*ShosetCreation{
	{Lname: "A", ShosetType: "cl", LocalAddress: "localhost:8001", RemoteAddresses: []string{""}, ProtocolType: "pki", Launched: false},
	{Lname: "B", ShosetType: "cl", LocalAddress: "localhost:8002", RemoteAddresses: []string{"localhost:8001"}, ProtocolType: "link", Launched: false},
	{Lname: "C", ShosetType: "cl", LocalAddress: "localhost:8003", RemoteAddresses: []string{"localhost:8002"}, ProtocolType: "link", Launched: false},
	{Lname: "D", ShosetType: "cl", LocalAddress: "localhost:8004", RemoteAddresses: []string{"localhost:8003"}, ProtocolType: "link", Launched: false},
	{Lname: "E", ShosetType: "cl", LocalAddress: "localhost:8005", RemoteAddresses: []string{"localhost:8004"}, ProtocolType: "link", Launched: false},
}

var IPbyLname = map[string]string{
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

var LinkedCircles = []*ShosetCreation{
	{Lname: "A", ShosetType: "cl", LocalAddress: IPbyLname["A"], RemoteAddresses: []string{IPbyLname["D"]}, ProtocolType: "pki", Launched: false},
	{Lname: "B", ShosetType: "cl", LocalAddress: IPbyLname["B"], RemoteAddresses: []string{IPbyLname["A"]}, ProtocolType: "link", Launched: false},
	{Lname: "C", ShosetType: "cl", LocalAddress: IPbyLname["C"], RemoteAddresses: []string{IPbyLname["A"]}, ProtocolType: "link", Launched: false},
	{Lname: "D", ShosetType: "cl", LocalAddress: IPbyLname["D"], RemoteAddresses: []string{IPbyLname["B"], IPbyLname["C"]}, ProtocolType: "link", Launched: false},
	{Lname: "E", ShosetType: "cl", LocalAddress: IPbyLname["E"], RemoteAddresses: []string{IPbyLname["D"], IPbyLname["F"]}, ProtocolType: "link", Launched: false},
	{Lname: "F", ShosetType: "cl", LocalAddress: IPbyLname["F"], RemoteAddresses: []string{IPbyLname["E"]}, ProtocolType: "link", Launched: false},
	{Lname: "G", ShosetType: "cl", LocalAddress: IPbyLname["G"], RemoteAddresses: []string{IPbyLname["F"]}, ProtocolType: "link", Launched: false},
	{Lname: "H", ShosetType: "cl", LocalAddress: IPbyLname["H"], RemoteAddresses: []string{IPbyLname["F"]}, ProtocolType: "link", Launched: false},
	{Lname: "I", ShosetType: "cl", LocalAddress: IPbyLname["I"], RemoteAddresses: []string{IPbyLname["G"], IPbyLname["H"]}, ProtocolType: "link", Launched: false},
}

var Circle = []*ShosetCreation{
	{Lname: "A", ShosetType: "cl", LocalAddress: IPbyLname["A"], RemoteAddresses: []string{}, ProtocolType: "pki", Launched: false},
	{Lname: "B", ShosetType: "cl", LocalAddress: IPbyLname["B"], RemoteAddresses: []string{IPbyLname["A"]}, ProtocolType: "link", Launched: false},
	{Lname: "C", ShosetType: "cl", LocalAddress: IPbyLname["C"], RemoteAddresses: []string{IPbyLname["A"]}, ProtocolType: "link", Launched: false},
	{Lname: "D", ShosetType: "cl", LocalAddress: IPbyLname["D"], RemoteAddresses: []string{IPbyLname["C"]}, ProtocolType: "link", Launched: false},
	{Lname: "E", ShosetType: "cl", LocalAddress: IPbyLname["E"], RemoteAddresses: []string{IPbyLname["D"], IPbyLname["B"]}, ProtocolType: "link", Launched: false},
}
