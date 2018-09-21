package main

import (
	LORAEMITTERCONFIG "github.com/willianxz/loraserver-device-sim/loratestes/lora_server_repeater/loraemitterconfig"
)



func main(){
	LORAEMITTERCONFIG.SendMessage("Mensagem", "bdfb8e5935883ea37e1b8fe135d29eba", "bd7f18f6c1a6dcba915f096627068d1c", "01936fdc")
}

