package main

import (
	"database/sql"
	"flag"
	"github.com/haxpax/gosms"
	"github.com/haxpax/gosms/modem"
	"log"
	"os"
	"strconv"
)

func main() {
	configFile := flag.String("config", "config.toml", "Path to the configuration file")
	flag.Parse()
	log.Println("main: ", "Initializing forms with config file: ", *configFile)

	// Load the config, abort if required config is not present
	appConfig, err := gosms.GetConfig(*configFile)
	if err != nil {
		log.Println("main: ", "Invalid config: ", err.Error(), " Aborting")
		os.Exit(1)
	}

	db, err := gosms.InitDB("sqlite3", "db.sqlite")
	if err != nil {
		log.Println("main: ", "Error initializing database: ", err, " Aborting")
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("main: ", "Error closing database: ", err)
		}
	}(db)

	serverHost := appConfig.Settings.ServerHost
	serverPort := strconv.Itoa(appConfig.Settings.ServerPort)

	serverUsername := "" // Set this to the appropriate value if needed
	serverPassword := "" // Set this to the appropriate value if needed

	numDevices := len(appConfig.Devices)
	log.Println("main: number of devices: ", numDevices)

	var modems []*modem.GSMModem
	for i := 0; i < numDevices; i++ {
		dev := appConfig.Devices[i]
		_port := dev.ComPort
		_baud := dev.BaudRate
		_devId := dev.DevID
		m := modem.New(_port, _baud, _devId)
		modems = append(modems, m)
	}

	bufferSize := appConfig.Settings.BufferSize
	bufferLow := appConfig.Settings.BufferLow
	loaderTimeout := appConfig.Settings.MsgTimeout
	loaderCountout := appConfig.Settings.MsgCountOut
	loaderTimeoutLong := appConfig.Settings.MsgTimeoutLong

	log.Println("main: Initializing worker")
	gosms.InitWorker(modems, bufferSize, bufferLow, loaderTimeout, loaderCountout, loaderTimeoutLong)

	log.Println("main: Initializing server")
	err = InitServer(serverHost, serverPort, serverUsername, serverPassword)
	if err != nil {
		log.Println("main: ", "Error starting server: ", err.Error(), " Aborting")
		os.Exit(1)
	}
}
