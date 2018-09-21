---------------------------------------------------------------------------------------------------

PRÉ REQUISITOS:
 
- Go instalado na maquina e com a variavel $GOPATH configurada corretamente, recomendo esse tutorial de instalação:
https://www.youtube.com/watch?v=hOcwQKxQsGc&t=188s

---------------------------------------------------------------------------------------------------

COMO USAR:

1 - Consulte o documento "Plano de trabalho Gateway Lora" para instalar e configurar a rede Lora localmente.

2 - Suba a rede Lora.

3 - Baixe do gitlab o projeto do Lora chamado de lora_server_repeater.

---------------------------------------------------------------------------------------------------

EMITIR MENSSAGEM PARA A REDE LORA

1 - Sua rede Lora terá configurações diferentes, por isso edite o arquivo loraEmitter.go baixado do projeto, com os paramêtros corretos que sua rede gerou ex:

LORAEMITTERCONFIG.SendMessage("Mensagem", "nwsHexKeyDaSuaRede", "appHexKeyDaSuaRede", "devHexAddrDaSuaRede")

2 - Configure o loraListen.go com os paramêtros corretos da rede para o qual você queira emittir uma menssagem ao receber alguma.

3 - Entre na pasta loraemitterconfig, edite o arquivo loraemitterconfig.go e configure a variavel gwMac com o id do seu Gateway ex:

gwMac := "1111111111111111"

4 - Entre na pasta loraemitterconfig e edite a variavel frameCount com o valor acima do que aparece em sua rede que esta no arquivo loraemitterconfig.go

5 - Abra o terminal, entre na pasta do projeto, para gerar um novo build basta usar o comando:
	
go build loraEmitter.go 

6 - Após o go ter gerado com sucesso o build e com isso um novo arquivo executavel, rode o executavel, no mesmo terminal usando o comando:

./loraEmitter 

7 - Se tudo esta certo, no mesmo terminal, irá mostrar o que o executavel está enviado para a rede Lora local e la a menssagem estara chegando em Applications / nomedaaplicacao / Devices / nomedodevice
na aba de LIVE DEVICE DATA.



---------------------------------------------------------------------------------------------------

ESCUTAR MENSSAGENS DA REDE LORA


1 - Abra o terminal, entre na pasta do projeto, para gerar um novo build basta usar o comando:
	
go build loraListen.go 

2 - Após o go ter gerado com sucesso o build e com isso um novo arquivo executavel, rode o executavel, no mesmo terminal usando o comando:

./loraListen 

3 - Se tudo esta certo, no mesmo terminal, irá mostrar o que está sendo enviado para a rede Lora local.


OBS: O loraListen também estara fazendo o papel de emittir menssagem ao se receber alguma.


---------------------------------------------------------------------------------------------------

OBSERVAÇÕES

- Perceba que o atributo data nas menssagens, estão encryptografados em Base 64, por esse motivo para saber o conteudo da menssagem é necessario descriptografa-la, aqui está um site que faz isso:
https://www.base64decode.org/

- O FrameCount se refere ao fCnt.


---------------------------------------------------------------------------------------------------

ERROS COMUNS QUE PODEM OCORRER:

- Depois de gerar o build do arquivo loraEmitter.go e rodar, se la na rede Lora Local, não estiver aparecendo as menssagens que estão sendo enviadas é provavelmente devido ao frameCont, pois ele não pode ser igual, é como um id unico que a rede vai salvando, por isso cada vez que gerar um novo build é bom certificar que o frameCount vai ser diferente editando essa variavel que está dentro do arquivo loraemitterconfig.go antes de gerar o build. 


---------------------------------------------------------------------------------------------------

