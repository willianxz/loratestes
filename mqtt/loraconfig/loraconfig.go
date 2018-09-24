package loraconfig

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	

	"github.com/brocaar/lorawan"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)


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
func (d *Device) Uplink(client MQTT.Client, mType lorawan.MType, fPort uint8, rxInfo *RxInfo, menssagem string) error {

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
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(menssagem)}},
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

//publish publishes a message to the broker. | função que todos os emitters usão para enviar para a rede lora.
func publish(client MQTT.Client, topic string, v interface{}) error {

	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	
	//Ocultei todas as informações que ele envia para a rede, para que nos Emitters só mostre que está enviado a menssagem solicitada.
	//fmt.Println("Publishing:")
	//fmt.Println(string(bytes)) 	


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

