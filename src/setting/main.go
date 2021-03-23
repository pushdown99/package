package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	express "github.com/DronRathore/goexpress"
	"github.com/joho/godotenv"
	"golang.org/x/sys/windows/registry"
)

type Tmpl struct {
	HttpHost  string
	WsHost    string
	RcnNum    string
	Printer   string
	Port1     string
	Port2     string
	HeartBeat string
	Ports     string
	PubAddr   string
	IpAddr    string
	MacAddr   string
}

type Config struct {
	HttpHost  string `json:"server"`
	WsHost    string `json:"ws"`
	RcnNum    string `json:"rcn"`
	Printer   string `json:"printer"`
	Port1     string `json:"port1"`
	Port2     string `json:"port2"`
	HeartBeat string `json:"heartbeat"`
}

type Port struct {
	Port1    string `json:"port1"`
	Port2    string `json:"port2"`
}

func createVirtualSerial (Port1 string, Port2 string) {
  os.Chdir("c:\\hancom")
  o, err := exec.Command("c:\\hancom\\com0com.exe",  "remove",  "1").Output()
  if err != nil {
    log.Printf("[agent] com0com execution error: %s", err)
  }
  log.Printf("%s", o)

  o, err = exec.Command("c:\\hancom\\com0com.exe",  "install",  "1", Port1, Port2).Output()
  if err != nil {
    log.Printf("[agent] com0com execution error: %s", err)
  }
  log.Printf("%s", o)
}

func GetPublicIP () string {
	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}	
	return string(ip)
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil { log.Fatal(err) }
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP
}
 
 func GetOutboundMac(currentIP string) string {
	var currentNetworkHardwareName string
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
	   if addrs, err := interf.Addrs(); err == nil {
		  for _, addr := range addrs {
			 if strings.Contains(addr.String(), currentIP) { currentNetworkHardwareName = interf.Name }
		  }
	   }
	}
	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)
	if err != nil { log.Printf("[agent] Interface error: %s", err) }
	return netInterface.HardwareAddr.String()
}

func findCOM () []string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\\DEVICEMAP\\SERIALCOMM`, registry.QUERY_VALUE)
	if err != nil {
	  log.Fatal(err)
	}
	defer k.Close()
	ki, err := k.Stat() // ki.SubKeyCount, ki.ValueCount
  
	if err != nil {
		log.Fatal(err)
	}
  
	s, err := k.ReadValueNames(int(ki.ValueCount))
	if err != nil {
		log.Fatal(err)
	}
	ports := make([]string, ki.ValueCount)
  
	for i, name := range s {
		q, _, err := k.GetStringValue(name)
		if err != nil {
			log.Fatal(err)
		}
		ports[i] = q
	}
	return ports
}

func getConfig () Tmpl {
	err := godotenv.Load("c:\\hancom\\.env")
	if err != nil { log.Fatal("Error loading .env file") }
	HttpHost  := os.Getenv("SERVER")
	WsHost    := os.Getenv("WS")
	RcnNum    := os.Getenv("RCN")
	Printer   := os.Getenv("PRINTER")
	Port1     := os.Getenv("PORT1")
	Port2     := os.Getenv("PORT2")
	HeartBeat := os.Getenv("HEARTBEAT")
	Ports     := strings.Join(findCOM (), ", ")
	IpAddr    := GetOutboundIP().String()
	PubAddr   := GetPublicIP()
	MacAddr   := GetOutboundMac(IpAddr)

	log.Printf("[agent] HttpHost: %s",  HttpHost )
	log.Printf("[agent] WsHost: %s",    WsHost   )
	log.Printf("[agent] RcnNum: %s",    RcnNum   )
	log.Printf("[agent] Printer: %s",   Printer  )
	log.Printf("[agent] Port1: %s",     Port1    )
	log.Printf("[agent] Port2: %s",     Port2    )
	log.Printf("[agent] HeartBeat: %s", HeartBeat)
	log.Printf("[agent] Ports: %s",     Ports    )
	log.Printf("[agent] IpAddr: %s",    IpAddr   )
	log.Printf("[agent] MacAddr: %s",   MacAddr  )

	return Tmpl {HttpHost, WsHost, RcnNum, Printer, Port1, Port2, HeartBeat, Ports, PubAddr, IpAddr, MacAddr}
}

func putConfig (conf Config) {
	f, _ := os.Create("c:\\hancom\\.env")
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintf(w, "SERVER=%v\n",    conf.HttpHost )
	fmt.Fprintf(w, "WS=%v\n",        conf.WsHost   )
	fmt.Fprintf(w, "RCN=%v\n",       conf.RcnNum   )
	fmt.Fprintf(w, "PRINTER=%v\n",   conf.Printer  )
	fmt.Fprintf(w, "PORT1=%v\n",     conf.Port1    )
	fmt.Fprintf(w, "PORT2=%v\n",     conf.Port2    )
	fmt.Fprintf(w, "HEARTBEAT=%v\n", conf.HeartBeat)
	w.Flush()
}

func main () { 
  var app = express.Express()

  app.Get("/", func(req express.Request, res express.Response) {
	data := getConfig () 
    res.Render("index.html", &data)
  })

  app.Post("/json/update", func(req express.Request, res express.Response) {
	var conf Config
    err := req.JSON().Decode(&conf)
    if err != nil {
      res.Error(400, "Invalid JSON")
    } else {
	  putConfig (conf)
	  res.JSON(conf)
    }
  })

  app.Post("/json/vport", func(req express.Request, res express.Response) {
	var port Port
    err := req.JSON().Decode(&port)
    if err != nil {
      res.Error(400, "Invalid JSON")
    } else {
	  data := getConfig ()
	  putConfig (Config{data.HttpHost, data.WsHost, data.RcnNum, data.Printer, port.Port1, port.Port2, data.HeartBeat})
      createVirtualSerial (port.Port1, port.Port2)
      res.JSON(port)
    }
  })

  dir, err := os.Getwd()
  if err != nil {
	log.Printf("Current working directory error: %s", err)
  }
  log.Printf("Current working directory: %s",dir)

  app.Start("8080")
}