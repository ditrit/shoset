package shoset

import "testing"

var cfg *Config

func TestNewConfig(t *testing.T) {
	cfg = NewConfig("test")
	if cfg == nil {
		t.Errorf("TestNewConfig didn't work")
	}
	cfg.SetFileName("test")
}

func TestInitFolders(t *testing.T) {
	TestNewConfig(t)
	_, err := cfg.InitFolders(cfg.GetFileName())
	if err != nil {
		t.Errorf("TestInitFolders didn't work")
	}
}

func TestSet(t *testing.T) {
	TestInitFolders(t)
	cfg.AppendToKey("protocol_test", []string{"address_test"})
}

func TestWriteConfig(t *testing.T) {
	TestSet(t)
	err := cfg.WriteConfig(cfg.GetFileName())
	if err != nil {
		t.Errorf("TestWriteConfig didn't work" + err.Error())
	}
}

func TestReadConfig(t *testing.T) {
	TestWriteConfig(t)
	err := cfg.ReadConfig(cfg.GetFileName())
	if err != nil {
		t.Errorf("TestReadConfig didn't work " + err.Error())
	}
}

func TestGetSlice(t *testing.T) {
	TestWriteConfig(t)
	addresses := cfg.GetSlice("protocol_test")
	if addresses[0] != "address_test" {
		t.Errorf("TestGetSlice didn't work ")
	}
}
