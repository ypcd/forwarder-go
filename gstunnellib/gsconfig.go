package gstunnellib

import (
	"encoding/json"
	"io"
	"os"

	. "github.com/ypcd/gforwarder/gstunnellib/gsrand"
)

/*
	type gsConfig_1 struct {
		Listen             string
		Server             string
		Key                string
		Debug              bool
		Tmr_display_time   int
		Tmr_changekey_time int
		g_Mt_model           bool
	}
*/
type GsConfig struct {
	Listen  string
	Servers []string
	Key     string

	Tmr_display_time   int
	Tmr_changekey_time int
	NetworkTimeout     int

	Debug    bool
	Mt_model bool
	IOModel  string
}

func (gs *GsConfig) GetServer_rand() string {
	return gs.Servers[GetRDCInt_max(int64(len(gs.Servers)))]
}
func (gs *GsConfig) GetServers() []string {
	return gs.Servers
}

func CreateGsconfig(confg string) *GsConfig {
	f, err := os.Open(confg)
	CheckError_exit(err)
	defer f.Close()

	buf, err := io.ReadAll(f)
	CheckError(err)

	//fmt.Println(string(buf))
	var gsconfig GsConfig

	gsconfig.Tmr_display_time = 6
	gsconfig.Tmr_changekey_time = 60
	gsconfig.NetworkTimeout = 60
	gsconfig.Debug = false
	gsconfig.Mt_model = true
	gsconfig.IOModel = "rw"

	err = json.Unmarshal(buf, &gsconfig)
	CheckError(err)

	if gsconfig.Servers == nil || gsconfig.Listen == "" {
		g_logger.Fatalln("Gstunnel config is error.")
	}
	if gsconfig.IOModel != "rw" && gsconfig.IOModel != "copy" {
		panic("gsconfig.IOModel != 'rw' && gsconfig.IOModel != 'copy'")
	}
	return &gsconfig
}
