package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	"strings"

	"github.com/perlin-network/noise/crypto/ed25519"
	"github.com/perlin-network/noise/examples/chat/messages"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/network"
	"github.com/perlin-network/noise/network/discovery"
	"github.com/perlin-network/noise/types/opcode"
)

type ChatPlugin struct{ *network.Plugin }

func (state *ChatPlugin) Receive(ctx *network.PluginContext) error {
	switch msg := ctx.Message().(type) {
	case *messages.ChatMessage:
		log.Info().Msgf("<%s> %s", ctx.Client().ID.Address, msg.Message)
	}

	return nil
}

func main() {
	// process other flags
	portFlag := flag.Int("port", 3000, "port to listen to")
	hostFlag := flag.String("host", "localhost", "host to listen to")
	protocolFlag := flag.String("protocol", "tcp", "protocol to use (kcp/tcp)")
	peersFlag := flag.String("peers", "", "peers to connect to")
	flag.Parse()

	port := uint16(*portFlag)
	host := *hostFlag
	protocol := *protocolFlag
	peers := strings.Split(*peersFlag, ",")

	keys := ed25519.RandomKeyPair()

	log.Info().Msgf("Private Key: %s", keys.PrivateKeyHex())
	log.Info().Msgf("Public Key: %s", keys.PublicKeyHex())

	opcode.RegisterMessageType(opcode.Opcode(1000), &messages.ChatMessage{})
	builder := network.NewBuilder()
	builder.SetKeys(keys)
	builder.SetAddress(network.FormatAddress(protocol, host, port))

	// Register peer discovery plugin.
	builder.AddPlugin(new(discovery.Plugin))

	// Add custom chat plugin.
	builder.AddPlugin(new(ChatPlugin))

	net, err := builder.Build()
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	go net.Listen()

	if len(peers) > 0 {
		net.Bootstrap(peers...)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')

		// skip blank lines
		if len(strings.TrimSpace(input)) == 0 {
			continue
		}

		ss := strings.Fields(input)

		myRecipient := ss[0]
		myMsg := ss[1]

		ctx := network.WithSignMessage(context.Background(), true)

		if client, err := net.Client("tcp://127.0.01:"+myRecipient); err == nil{
			client.Tell(ctx, &messages.ChatMessage{Message: myMsg})
			log.Info().Msgf("<%s> %s", net.Address, myMsg)
		}
	}

}
