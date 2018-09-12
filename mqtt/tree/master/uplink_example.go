package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"
	

	"github.com/brocaar/lorawan"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)


var frameCont = 2229



//Message holds physical payload and rx info.
type Message struct {
	PhyPayload string  `json:"phyPayload"`
	RxInfo     *RxInfo `json:"rxInfo"`
}

//RxInfo holds all relevant information of a lora message.
type RxInfo struct {
	Mac       string    `json:"mac"`
	Time      string    `json:"time"`
	Timestamp int32     `json:"timestamp"`
	Frequency int       `json:"frequency"`
	Channel   int       `json:"channel"`
	RfChain   int       `json:"rfChain"`	
	CrcStatus int       `json:"crcStatus"`
	CodeRate  string    `json:"codeRate"`
	Rssi      int       `json:"rssi"`
	LoRaSNR   float32   `json:"loRaSNR"`  	
	Size      int       `json:"size"`
	Datr 	  string
	DataRate  struct {
			Modulation   string `json:"modulation"`
			SpreadFactor int    `json:"spreadFactor"`
			Bandwidth    int    `json:"bandwidth"`	
			BitRate     int		

	} `json:"dataRate"`		
	Board      int    
	Antenna    int    
}

//DataRate holds relevant info for data rate.
type DataRate struct {	
	Modulation   string `json:"modulation"`	
	SpreadFactor int    `json:"spreadFactor"`
	Bandwidth    int    `json:"bandwidth"`
	BitRate     int
}

//Device holds device keys, addr, eui and fcnt.
type Device struct {
	DevEUI  lorawan.EUI64
	DevAddr lorawan.DevAddr
	NwkSKey lorawan.AES128Key
	AppSKey lorawan.AES128Key
	AppKey  [16]byte
	AppEUI  lorawan.EUI64
	UlFcnt  uint32
	DlFcnt  uint32
}

//Join sends a join request for a given device (OTAA) and rxInfo.
func (d *Device) Join(client MQTT.Client, gwMac string, rxInfo RxInfo) error {

	joinPhy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			JoinEUI:  d.AppEUI,
			DevEUI:   d.DevEUI,
			DevNonce: lorawan.DevNonce(uint16(65535)),
		},
	}

	if err := joinPhy.SetUplinkJoinMIC(d.AppKey); err != nil {
		return err
	}

	joinStr, err := joinPhy.MarshalText()
	if err != nil {
		return err
	}

	message := &Message{
		PhyPayload: string(joinStr),
		RxInfo:     &rxInfo,
	}

	pErr := publish(client, "gateway/"+rxInfo.Mac+"/rx", message)

	return pErr

}

//Uplink sends an uplink message for a given device, mType (UnconfirmedDataUp, ConfirmedDataUp), rxInfo and payload (unencrypted).
func (d *Device) Uplink(client MQTT.Client, mType lorawan.MType, fPort uint8, rxInfo *RxInfo, payload []byte) error {

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: mType,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: d.DevAddr,
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt:  d.UlFcnt,
				FOpts: []lorawan.Payload{}, // you can leave this out when there is no MAC command to send
			},
			FPort:      &fPort,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte("HelloISI")}},
		},
	}

	if err := phy.EncryptFRMPayload(d.AppSKey); err != nil {
		fmt.Printf("encrypt frm payload: %s", err)
		return err
	}

	if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, d.NwkSKey, d.AppSKey); err != nil {
		fmt.Printf("set uplink mic error: %s", err)
		return err
	}

	upDataStr, err := phy.MarshalText()
	if err != nil {
		fmt.Printf("marshal up data error: %s", err)
		return err
	}

	message := &Message{
		PhyPayload: string(upDataStr),
		RxInfo:     rxInfo,
	}

	pErr := publish(client, "gateway/"+rxInfo.Mac+"/rx", message)


	//Increase uplink fcnt if unconfirmed
	if pErr != nil && mType == lorawan.UnconfirmedDataUp {
		d.UlFcnt++
	}

	return pErr

}



/////////////////////////
// Auxiliary functions //
/////////////////////////

//HexToDevAddress converts a string hex representation of a device address to a [4]byte.
func HexToDevAddress(hexAddress string) ([4]byte, error) {
	var devAddr ([4]byte)
	da, err := hex.DecodeString(hexAddress)
	if err != nil {
		return devAddr, err
	}
	copy(devAddr[:], da[:])
	return devAddr, nil
}

//HexToKey converts a string hex representation of an AES128Key to a [16]byte.
func HexToKey(hexKey string) ([16]byte, error) {
	var key ([16]byte)
	k, err := hex.DecodeString(hexKey)
	if err != nil {
		return key, err
	}
	copy(key[:], k[:])
	return key, nil
}

func testMIC(appKey [16]byte, appEUI, devEUI [8]byte) error {
	joinPhy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			JoinEUI:  appEUI,
			DevEUI:   devEUI,
			DevNonce: lorawan.DevNonce(uint16(65535)),
		},
	}

	if err := joinPhy.SetUplinkJoinMIC(appKey); err != nil {
		fmt.Printf("set uplink join mic error: %s", err)
		return err
	}

	fmt.Println("Printing MIC")
	fmt.Println(hex.EncodeToString(joinPhy.MIC[:]))

	joinStr, err := joinPhy.MarshalText()
	if err != nil {
		fmt.Printf("join marshal error: %s", err)
		return err
	}
	fmt.Println(joinStr)

	return nil
}

//publish publishes a message to the broker.
func publish(client MQTT.Client, topic string, v interface{}) error {

	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	

	fmt.Println("Publishing:")
	fmt.Println(string(bytes))
	fmt.Println("FRAME COUNT:");
	fmt.Println(frameCont);


	if token := client.Publish(topic, 0, false, bytes); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}

	return nil
}

/////////////////////////
// Custom data helpers //
/////////////////////////

func generateRisk(r int8) []byte {
	risk := uint8(r)
	bRep := make([]byte, 1)
	bRep[0] = risk
	return bRep
}

func generateTemp1byte(t int8) []byte {
	temp := uint8(t)
	bRep := make([]byte, 1)
	bRep[0] = temp
	return bRep
}

func generateTemp2byte(t int16) []byte {

	temp := uint16(float32(t/127.0) * float32(math.Pow(2, 15)))
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, temp)
	return bRep
}

func generateLight(l int16) []byte {

	light := uint16(l)
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, light)
	return bRep
}

func generateAltitude(a float32) []byte {

	alt := uint16(float32(a/1200) * float32(math.Pow(2, 15)))
	bRep := make([]byte, 2)
	binary.BigEndian.PutUint16(bRep, alt)
	return bRep
}

func generateLat(l float32) []byte {
	lat := uint32((l / 90.0) * float32(math.Pow(2, 31)))
	bRep := make([]byte, 4)
	binary.BigEndian.PutUint32(bRep, lat)
	return bRep
}

func generateLng(l float32) []byte {
	lng := uint32((l / 180.0) * float32(math.Pow(2, 31)))
	bRep := make([]byte, 4)
	binary.BigEndian.PutUint32(bRep, lng)
	return bRep
}

func main() {

	//Connect to the broker
	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://localhost:1884")
	opts.SetUsername("")
	opts.SetPassword("")

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Connection error")
		fmt.Println(token.Error())
	}

	fmt.Println("Connection established.")

	//Build your node with known keys (ABP).
	nwsHexKey := "bdfb8e5935883ea37e1b8fe135d29eba"
	appHexKey := "bd7f18f6c1a6dcba915f096627068d1c"
	devHexAddr := "01936fdc"
	devAddr, err := HexToDevAddress(devHexAddr)
	if err != nil {
		fmt.Printf("dev addr error: %s", err)
	}

	nwkSKey, err := HexToKey(nwsHexKey)
	if err != nil {
		fmt.Printf("nwkskey error: %s", err)
	}

	appSKey, err := HexToKey(appHexKey)
	if err != nil {
		fmt.Printf("appskey error: %s", err)
	}

	appKey := [16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	appEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 2}
	devEUI := [8]byte{0, 0, 0, 0, 0, 0, 0, 2}
	

	device := &Device{
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

		/*
			*	Make up some random values.
			*
			*	These should be decoded at lora-app-server with a proper function.
			* 	For this example, the object should look like this:
				obj : {
					"temperature": {
						"value":((bytes[0]*256+bytes[1])/100),"unit":"°C"
					},
					"pressure": {
						"value":((bytes[2]*16*16*16*16+bytes[3]*256+bytes[4])/100),"unit":"hPa"
					},
					"humidity": {
						"value":((bytes[5]*256+bytes[6])/1024),"unit":"%"
					}
				}
			*
		*/

		rand.Seed(time.Now().UnixNano() / 10000)
		temp := [2]byte{uint8(rand.Intn(25)), uint8(rand.Intn(100))}
		pressure := [3]byte{uint8(rand.Intn(2)), uint8(rand.Intn(20)), uint8(rand.Intn(100))}
		humidity := [2]byte{uint8(rand.Intn(100)), uint8(rand.Intn(100))}

		//Create the payload, data rate and rx info.
		//payload := []byte{temp[0], temp[1], pressure[0], pressure[1], pressure[2], humidity[0], humidity[1]}

		payload := []byte{temp[0], temp[1], pressure[0], pressure[1], pressure[2], humidity[0], humidity[1]}

		//Change to your gateway MAC to build RxInfo.
		gwMac := "0001020304050607"

		//Construct DataRate RxInfo with proper values according to your band (example is for US 915).

		
		dataRate := DataRate{"LORA", 7, 125, 0}		

		rxInfo := &RxInfo{
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

		//Now send an uplink
		err = device.Uplink(client, lorawan.UnconfirmedDataUp, 1, rxInfo, payload)
		frameCont++
		device.UlFcnt = uint32(frameCont)
		device.DlFcnt = uint32(frameCont)

		
		if err != nil {
			fmt.Printf("couldn't send uplink: %s\n", err)
		}

		time.Sleep(3 * time.Second)

	}

}
