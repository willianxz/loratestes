package loraemitterconfig

import (	
	"fmt"	
	"time"	

	"github.com/brocaar/lorawan"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	LORACONFIG "github.com/willianxz/loraserver-device-sim/loratestes/lora_server_repeater/loraconfig"	
)

type Options struct {	
	Broker      string    `json:"modulation"`	
	Username    string    `json:"spreadFactor"`
	Password    string    `json:"bandwidth"`	
}


type DeviceSetup struct{
     NwsHexKey  string
     AppHexKey  string
     DevHexAddr string
}



func SendMessage(menssagem string, nwsHexKey string, appHexKey string, devHexAddr string) {
	 var frameCont = 2694

	//Crie a conexão com o MQTT:
        optionsConfig := Options{"tcp://localhost:1884","", ""}
        opts := MQTT.NewClientOptions()
        opts.AddBroker(optionsConfig.Broker)
	opts.SetUsername(optionsConfig.Username)
	opts.SetPassword(optionsConfig.Password)	
	client := MQTT.NewClient(opts)


	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Connection error")
		fmt.Println(token.Error())
	}


	deviceSetup := DeviceSetup{nwsHexKey, //Network session encryption key
        appHexKey, //Application session key
        devHexAddr, //Device address 
        }

	fmt.Println("Connection established.")

	
	devAddr, err := LORACONFIG.HexToDevAddress(deviceSetup.DevHexAddr)
	if err != nil {
		fmt.Printf("dev addr error: %s", err)
	}

	nwkSKey, err := LORACONFIG.HexToKey(deviceSetup.NwsHexKey)
	if err != nil {
		fmt.Printf("nwkskey error: %s", err)
	}

	appSKey, err := LORACONFIG.HexToKey(deviceSetup.AppHexKey)
	if err != nil {
		fmt.Printf("appskey error: %s", err)
	}

	appKey := [16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	appEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 2}
	devEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 2}
	

	device := &LORACONFIG.Device{
		DevEUI:  devEUI,
		DevAddr: devAddr,
		NwkSKey: nwkSKey,
		AppSKey: appSKey,
		AppKey:  appKey,
		AppEUI:  appEUI,
		UlFcnt:  uint32(frameCont),
		DlFcnt:  uint32(frameCont),
	}
	

	for {
	
		gwMac := "1111111111111111"  //Gateway ID

		dataRate := LORACONFIG.DataRate{"LORA", 7, 125, 0}

		rxInfo := &LORACONFIG.RxInfo{
			Mac:  gwMac,
			Time:      time.Now().Format(time.RFC3339),
			Timestamp: int32(time.Now().UnixNano() / 1000000000),
			Frequency: 866349812,
			Channel:   2,
			RfChain:   0,
			CrcStatus: 1,
			CodeRate:  "4/6",
			Rssi:      -35,
			LoRaSNR:   5.1,
			Size:      21,
			Datr: 	   "SF7BW125",
			DataRate:  dataRate,	
			Board: 0,
			Antenna: 0,
		}

		//Mande para a rede a nossa menssagem.
		err = device.Uplink(client, lorawan.UnconfirmedDataUp, 1, rxInfo, menssagem)

		fmt.Println("FRAME COUNT:");
		fmt.Println(frameCont);
		frameCont++

		

		device.UlFcnt = uint32(frameCont)
		device.DlFcnt = uint32(frameCont)

		
		if err != nil {
			fmt.Printf("couldn't send uplink: %s\n", err)
		}

		time.Sleep(3 * time.Second)

	}

}




func SendMessageListener(menssagem string, nwsHexKey string, appHexKey string, devHexAddr string) {
	 var frameCont = 2600 //Aqui o frameCount sempre será o mesmo.

	//Crie a conexão com o MQTT:
        optionsConfig := Options{"tcp://localhost:1884","", ""}
        opts := MQTT.NewClientOptions()
        opts.AddBroker(optionsConfig.Broker)
	opts.SetUsername(optionsConfig.Username)
	opts.SetPassword(optionsConfig.Password)	
	client := MQTT.NewClient(opts)


	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Connection error")
		fmt.Println(token.Error())
	}


	deviceSetup := DeviceSetup{nwsHexKey, //Network session encryption key
        appHexKey, //Application session key
        devHexAddr, //Device address 
        }

	fmt.Println("Connection established.")

	
	devAddr, err := LORACONFIG.HexToDevAddress(deviceSetup.DevHexAddr)
	if err != nil {
		fmt.Printf("dev addr error: %s", err)
	}

	nwkSKey, err := LORACONFIG.HexToKey(deviceSetup.NwsHexKey)
	if err != nil {
		fmt.Printf("nwkskey error: %s", err)
	}

	appSKey, err := LORACONFIG.HexToKey(deviceSetup.AppHexKey)
	if err != nil {
		fmt.Printf("appskey error: %s", err)
	}

	appKey := [16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	appEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 2}
	devEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 2}
	

	device := &LORACONFIG.Device{
		DevEUI:  devEUI,
		DevAddr: devAddr,
		NwkSKey: nwkSKey,
		AppSKey: appSKey,
		AppKey:  appKey,
		AppEUI:  appEUI,
		UlFcnt:  uint32(frameCont),
		DlFcnt:  uint32(frameCont),
	}
	

		gwMac := "1111111111111111"  //Gateway ID

		dataRate := LORACONFIG.DataRate{"LORA", 7, 125, 0}

		rxInfo := &LORACONFIG.RxInfo{
			Mac:  gwMac,
			Time:      time.Now().Format(time.RFC3339),
			Timestamp: int32(time.Now().UnixNano() / 1000000000),
			Frequency: 866349812,
			Channel:   2,
			RfChain:   0,
			CrcStatus: 1,
			CodeRate:  "4/6",
			Rssi:      -35,
			LoRaSNR:   5.1,
			Size:      21,
			Datr: 	   "SF7BW125",
			DataRate:  dataRate,	
			Board: 0,
			Antenna: 0,
		}

		//Mande para a rede a nossa menssagem.
		err = device.Uplink(client, lorawan.UnconfirmedDataUp, 1, rxInfo, menssagem)

		

		device.UlFcnt = uint32(frameCont)
		device.DlFcnt = uint32(frameCont)

		
		if err != nil {
			fmt.Printf("couldn't send uplink: %s\n", err)
		}

		time.Sleep(3 * time.Second)
}






