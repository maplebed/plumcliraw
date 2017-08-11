package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	flag "github.com/jessevdk/go-flags"
	"github.com/maplebed/libplumraw"
)

type Options struct {
	Email    string `short:"e" long:"email" descrption:"Email address to authenticate with the Plum Web API"`
	Password string `short:"p" long:"password" descrption:"Password to authenticate with the Plum Web API"`
	ID       string `long:"id" description:"For commands that require an ID, use this flag to set it"`

	LightpadIP string `long:"lpip" description:"Lightpad IP Address"`
	Port       int    `long:"port" description:"Lightpad Port" default:"8443"`
	HAT        string `long:"hat" description:"House Access Token - get from --action GetHouse"`
	Conf       string `long:"conf" description:"JSON used for Lightpad Set commands"`

	ListActions bool   `short:"l" long:"list_actions" description:"List available actions"`
	Action      string `short:"a" long:"action" description:"Call to make to the API or Lgihtpad"`

	TestMode bool `long:"test" description:"Run this CLI in Test mode"`
}

const version = "0.0.1"

func main() {
	var options Options
	flagParser := flag.NewParser(&options, flag.Default)
	flagParser.Parse()

	libplumraw.UserAgentAddition = fmt.Sprintf("rawcli/%s", version)

	if options.ListActions {
		fmt.Printf(`Available actions:

Web:
  * GetHouses               - get a list of all House IDs
  * GetHouse --id <id>     - get the description of a House
  * GetScenes               - get a list of all Scene IDs
  * GetScene --id <id>     - get the description of a Scene
  * GetRoom --id <id>      - get the description of a Room
  * GetLoad --id <id>     - get the description of a Load
  * GetLightpad --id <id> - get the description of a Lightpad

Lightpad - all require --lpip, --port, and --hat:
  * GetLoadMetrics                     - Get metrics about current power draw
  * SetLevel --level <int>             - Set the dim level range 0 (off) to 255 (on)
  * SetLightpadConfig --conf <string>  - Upload a new Lightpad config to the pad
  * SetLoadConfig  --conf <string>     - Upload a new Load config to the pad
  * SetLoadGlow  --conf <string>       - Turn on the glow ring manually
  * Subscribe  --conf <string>         - Listen for state changes from the Lightpad

Examples:
  ./plumcliraw -a GetHouses --email me@example.com --password 'friend'
  ./plumcliraw -a GetRoom --email me@example.com --password 'friend' --id dbb77fae-f027-4377-9f77-d46e0a4a7d49
  ./plumcliraw -a Subscribe --lpip 192.168.1.10 --port 8443 --hat 281babee-bb75-4a96-9de9-48c010089574
  ./plumcliraw -a SetLevel --lpip 192.168.1.10 --port 8443 --hat 281babee-bb75-4a96-9de9-48c010089574 --conf '{"level":0}' --id 8aae8c21-f60a-472d-a982-b89a7bb945e9
  ./plumcliraw -a GetLoadMetrics --lpip 192.168.1.10 --port 8443 --hat 281babee-bb75-4a96-9de9-48c010089574 --id 8aae8c21-f60a-472d-a982-b89a7bb945e9
`)
		os.Exit(0)
	}

	var conn libplumraw.WebConnection
	if options.TestMode {
		conn = makeTestConn()
	} else {
		conf := libplumraw.WebConnectionConfig{
			Email:    options.Email,
			Password: options.Password,
		}
		conn = libplumraw.NewWebConnection(conf)
	}
	switch options.Action {
	case "GetHouses":
		houses, err := conn.GetHouses()
		checkError(err)
		spew.Dump(houses)
	case "GetHouse":
		checkID("House ID", options.ID)
		house, err := conn.GetHouse(options.ID)
		checkError(err)
		spew.Dump(house)
	case "GetScenes":
		checkID("House ID", options.ID)
		scenes, err := conn.GetScenes(options.ID)
		checkError(err)
		spew.Dump(scenes)
	case "GetScene":
		checkID("Scene ID", options.ID)
		scene, err := conn.GetScene(options.ID)
		checkError(err)
		spew.Dump(scene)
	case "GetRoom":
		checkID("Room ID", options.ID)
		room, err := conn.GetRoom(options.ID)
		checkError(err)
		spew.Dump(room)
	case "GetLoad":
		checkID("Logical Load ID", options.ID)
		load, err := conn.GetLogicalLoad(options.ID)
		checkError(err)
		spew.Dump(load)
	case "GetLightpad":
		checkID("Lightpad ID", options.ID)
		pad, err := conn.GetLightpad(options.ID)
		checkError(err)
		spew.Dump(pad)
	case "GetLoadMetrics":
		checkLightpadFlags(options.LightpadIP, options.Port, options.HAT)
		ip := net.ParseIP(options.LightpadIP)
		checkIP(ip)
		lp := libplumraw.DefaultLightpad{
			LLID: options.ID,
			IP:   ip,
			Port: options.Port,
			HttpClient: &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}},
			HAT: options.HAT,
		}
		mets, err := lp.GetLogicalLoadMetrics()
		checkError(err)
		spew.Dump(mets)
	case "SetLevel":
		checkLightpadFlags(options.LightpadIP, options.Port, options.HAT)
		ip := net.ParseIP(options.LightpadIP)
		checkIP(ip)
		conf := struct{ Level int }{}
		err := json.Unmarshal([]byte(options.Conf), &conf)
		checkError(err)
		lp := libplumraw.DefaultLightpad{
			LLID: options.ID,
			IP:   ip,
			Port: options.Port,
			HttpClient: &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}},
			HAT: options.HAT,
		}
		err = lp.SetLogicalLoadLevel(conf.Level)
		checkError(err)
	case "SetLightpadConfig":
		checkLightpadFlags(options.LightpadIP, options.Port, options.HAT)
		ip := net.ParseIP(options.LightpadIP)
		checkIP(ip)
		conf := libplumraw.LightpadConfig{}
		err := json.Unmarshal([]byte(options.Conf), &conf)
		checkError(err)
		fmt.Printf("unpacked %s, %+v\n", ip, conf)
		buf, err := json.Marshal(conf)
		fmt.Printf("and remarshaled: %s\n", string(buf))
		lp := libplumraw.DefaultLightpad{
			LLID:       options.ID,
			IP:         ip,
			Port:       options.Port,
			HttpClient: &http.Client{},
		}
		err = lp.SetLightpadConfig(conf)
		checkError(err)
	case "SetLoadConfig":
		checkLightpadFlags(options.LightpadIP, options.Port, options.HAT)
		ip := net.ParseIP(options.LightpadIP)
		checkIP(ip)
		conf := libplumraw.LogicalLoadConfig{}
		err := json.Unmarshal([]byte(options.Conf), &conf)
		checkError(err)
		fmt.Printf("unpacked %s, %+v\n", ip, conf)
		buf, err := json.Marshal(conf)
		fmt.Printf("and remarshaled: %s\n", string(buf))
		lp := libplumraw.DefaultLightpad{
			LLID: options.ID,
			IP:   ip,
			Port: options.Port,
			HttpClient: &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}},
			HAT: options.HAT,
		}
		err = lp.SetLogicalLoadConfig(conf)
		checkError(err)
	case "SetLoadGlow":
		checkLightpadFlags(options.LightpadIP, options.Port, options.HAT)
		ip := net.ParseIP(options.LightpadIP)
		checkIP(ip)
		conf := libplumraw.ForceGlow{}
		err := json.Unmarshal([]byte(options.Conf), &conf)
		checkError(err)
		fmt.Printf("unpacked %s, %+v\n", ip, conf)
	case "Subscribe":
		checkLightpadFlags(options.LightpadIP, options.Port, options.HAT)
		ip := net.ParseIP(options.LightpadIP)
		checkIP(ip)
		fmt.Printf("unpacked %s\n", ip)
		lp := libplumraw.DefaultLightpad{
			LLID: options.ID,
			IP:   ip,
			Port: options.Port,
			HttpClient: &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}},
			HAT:          options.HAT,
			StateChanges: make(chan libplumraw.Event, 0),
		}
		err := lp.Subscribe(context.Background())
		checkError(err)
		for ev := range lp.StateChanges {
			switch ev := ev.(type) {
			case libplumraw.LPEDimmerChange:
				fmt.Printf("heard a %s event with value %d\n", ev.Type, ev.Level)
				// spew.Dump(ev.(libplumraw.LPEDimmerChange))
			case libplumraw.LPEPower:
				fmt.Printf("heard a %s event with value %d\n", ev.Type, ev.Watts)
				// spew.Dump(ev.(libplumraw.LPEPower))
			case libplumraw.LPEPIRSignal:
				fmt.Printf("heard a %s event with value %d\n", ev.Type, ev.Signal)
				// lp.SetLogicalLoadLevel(255) // turn the light on in response to motion
				// spew.Dump(ev.(libplumraw.LPEPower))
			case libplumraw.LPEUnknown:
				fmt.Printf("heard an unknown event with message %s\n", ev.Message)
				// spew.Dump(ev.(libplumraw.LPEPower))
			}
		}

	default:
		fmt.Printf("Action '%s' not recognized\n", options.Action)
	}

}

func checkID(name string, flag string) {
	if flag == "" {
		fmt.Printf("%s must be specified with the --id flag\n", name)
		os.Exit(1)
	}
}

func checkIP(ip net.IP) {
	if ip == nil {
		fmt.Printf("IP address failed to parse.\n", ip)
		os.Exit(1)
	}
}

func checkLightpadFlags(lpip string, port int, hat string) {
	if lpip == "" || port == 0 || hat == "" {
		fmt.Println("Lightpad IP address, port number, and House Access Token must all be specified.")
		os.Exit(1)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// type DefaultLightpad
//     func (l *DefaultLightpad) GetLogicalLoadMetrics() (LogicalLoadMetrics, error)
//     func (l *DefaultLightpad) SetLogicalLoadConfig(conf LogicalLoadConfig) error
//     func (l *DefaultLightpad) SetLogicalLoadGlow(glow ForceGlow) error
//     func (l *DefaultLightpad) SetLogicalLoadLevel(level int) error
//     func (l *DefaultLightpad) Subscribe() chan LightpadEvent
// type DefaultLightpadHeartbeat
//     func (d *DefaultLightpadHeartbeat) Listen(ctx context.Context) chan LightpadAnnouncement
// type DefaultWebConnection
//     func (c *DefaultWebConnection) GetHouse(hid string) (House, error)
//     func (c *DefaultWebConnection) GetHouses() (Houses, error)
//     func (c *DefaultWebConnection) GetLightpad(lpid string) (LightpadSpec, error)
//     func (c *DefaultWebConnection) GetLogicalLoad(llid string) (LogicalLoad, error)
//     func (c *DefaultWebConnection) GetRoom(rid string) (Room, error)

func makeTestConn() *libplumraw.TestWebConnection {
	conn := &libplumraw.TestWebConnection{
		Houses: libplumraw.Houses{"aaa", "bbb"},
		House: libplumraw.House{
			ID:      "ccc",
			RoomIDs: []string{"ddd", "eee"},
			LatLong: struct {
				Latitude  float64 `json:"latitude_degrees_north,omitempty"` // decimal degrees North
				Longitude float64 `json:"longitude_degrees_west,omitempty"` // decimal degrees West
			}{123.456, 789.012},
			AccessToken: "fff",
			Name:        "ggg",
			TimeZone:    234,
		},
		Room: libplumraw.Room{
			ID:      "hhh",
			Name:    "iii",
			HouseID: "jjj",
			LLIDs:   []string{"kkk", "lll"},
		},
		LogicalLoad: libplumraw.LogicalLoad{
			ID:     "mmm",
			Name:   "nnn",
			LPIDs:  []string{"ooo", "ppp"},
			RoomID: "qqq",
		},
		LightpadSpec: libplumraw.LightpadSpec{
			ID:             "rrr",
			LLID:           "sss",
			Config:         libplumraw.LightpadConfig{},
			IsProvisioned:  true,
			CustomGestures: 0,
			Name:           "ttt",
		},
	}
	return conn
}
