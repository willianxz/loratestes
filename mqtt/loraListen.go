package main

import (	
	"fmt"
	"os"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	LORAEMITTERCONFIG "github.com/willianxz/loraserver-device-sim/loratestes/lora_server_repeater/loraemitterconfig"
        READDATATXT "github.com/willianxz/loraserver-device-sim/loratestes/lora_server_repeater/readdatatxt"
)


var brokerAllMessages = make(chan bool)

//Lê as informações do arquivo configData.txt e armazena em variaveis
var config, err = READDATATXT.ReadConfig(`/home/docker/go/src/github.com/willianxz/loraserver-device-sim/loratestes/lora_server_repeater/loraemitterconfig/configData`)

var nwsHexKey = config["nwsHexKey"]
var appHexKey = config["appHexKey"]
var devHexAddr = config["devHexAddr"]
var broker = config["broker"]
var username = config["username"]
var password = config["password"]


func brokerAllMessagesHandler(client MQTT.Client, msg MQTT.Message) {
	brokerAllMessages <- true	
	fmt.Printf("[%s] ", msg.Topic())
	fmt.Printf("%s\n", msg.Payload())

        //chama a função que envia as informações armazenadas em variaveis para a rede.
	LORAEMITTERCONFIG.SendMessageListener("Redirecionou a Mensagem", nwsHexKey, appHexKey, devHexAddr)
}


func main() {
       	
	//Crie a conexão com o MQTT:
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword(password)

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Connection error")
		fmt.Println(token.Error())
	}

	//#
	if token := client.Subscribe("application/1/device/#", 0, brokerAllMessagesHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}


	countMessages := 0
	maxMessages   := 100

	for i := 0; i < maxMessages; i++ { 
		select {
		case <-brokerAllMessages:
			countMessages++			
		}
	}


	fmt.Printf("Total de menssagens gerais recebidas:%3d \n", countMessages)

	client.Disconnect(250)

}
