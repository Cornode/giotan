/*
MIT License

Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/alecthomas/kingpin"
	"github.com/cornodeledger/gcornode"
)

func main() {
	var (
		app = kingpin.New("gcornoden", "gcornode CLI Tool")

		send      = app.Command("send", "Send token")
		recipient = send.Flag("recipient", "recipient address").Required().String()
		sender    = send.Flag("sender", "sender addresses, separated with comma").String()
		amount    = send.Flag("amount", "amount to send").Required().Int64()
		tag       = send.Flag("tag", "tag to send").Default("PRETTYGcornodeN").String()
		mwm       = send.Flag("mwm", "MinWeightMagnituce").Default("18").Int64()

		addresses = app.Command("addresses", "List used/unused addresses")

		newseed = app.Command("new", "create a new seed")
	)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case send.FullCommand():
		fmt.Print("input your seed:")
		seed, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			panic(err)
		}
		Send(string(seed), *recipient, *sender, *amount, *mwm, *tag)
	case addresses.FullCommand():
		fmt.Print("input your seed:")
		seedA, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			panic(err)
		}
		handleAddresses(string(seedA))
	case newseed.FullCommand():
		seed := gcornode.NewSeed()
		fmt.Println("New seed: ", seed)
		fmt.Printf("To display addresses, run\n\t%s addresses\n", os.Args[0])
		fmt.Println("and input the seed above.")
	}
}

func handleAddresses(seed string) {
	server := gcornode.RandomNode()
	fmt.Printf("using IRI server: %s\n", server)
	seedT, err := gcornode.ToTrytes(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "You must specify valid seed")
		os.Exit(-1)
	}
	api := gcornode.NewAPI(server, nil)
	adr, adrs, err := gcornode.GetUsedAddress(api, seedT, 2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get addresses: %s\n", err.Error())
		os.Exit(-1)
	}
	var resp *gcornode.GetBalancesResponse
	if len(adrs) > 0 {
		resp, err = api.GetBalances(adrs, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot get balance: %s\n", err.Error())
			os.Exit(-1)
		}
	}
	fmt.Println("address info:")
	fmt.Println("used:")
	for i, a := range adrs {
		fmt.Printf("\t%s (balance=%d)\n", a, resp.Balances[i])
	}
	fmt.Println("\nunused:")
	fmt.Printf("\t%s\n", adr)
}

func check(seed, recipient, sender string, amount int64) (gcornode.Trytes, gcornode.Address, []gcornode.Address) {
	if amount <= 0 {
		fmt.Fprintln(os.Stderr, "You must specify the amount with positive value.")
		os.Exit(-1)
	}
	seedT, err := gcornode.ToTrytes(seed)
	if err != nil {
		fmt.Fprintln(os.Stderr, "You must specify valid seed")
		os.Exit(-1)
	}
	recipientT, err := gcornode.ToAddress(recipient)
	if err != nil {
		fmt.Fprintln(os.Stderr, "You must specify valid recipient")
		os.Exit(-1)
	}
	var senderT []gcornode.Address
	if sender != "" {
		senders := strings.Split(sender, ",")
		senderT = make([]gcornode.Address, len(senders))
		for i, s := range senders {
			senderT[i], err = gcornode.ToAddress(s)
			if err != nil {
				fmt.Fprintln(os.Stderr, "You must specify valid sender")
				os.Exit(-1)
			}
		}
	}
	return seedT, recipientT, senderT
}

func sendToSender(api *gcornode.API, trs []gcornode.Transfer, sender []gcornode.Address, seedT gcornode.Trytes, mwm int64) (gcornode.Bundle, error) {
	_, adrs, err := gcornode.GetUsedAddress(api, seedT, 2)
	if err != nil {
		return nil, err
	}
	adrinfo := make([]gcornode.AddressInfo, len(sender))
	for i, s := range sender {
		for j, a := range adrs {
			if s == a {
				adrinfo[i] = gcornode.AddressInfo{
					Seed:     seedT,
					Index:    j,
					Security: 2,
				}
				break
			}
		}
		return nil, fmt.Errorf("cannot found address %s from seed", s)
	}
	bdl, err := gcornode.PrepareTransfers(api, seedT, trs, adrinfo, "", 2)
	if err != nil {
		return nil, err
	}
	name, pow := gcornode.GetBestPoW()
	fmt.Fprintf(os.Stderr, "using PoW:%s\n", name)
	err = gcornode.SendTrytes(api, gcornode.Depth, []gcornode.Transaction(bdl), mwm, pow)
	return bdl, err
}

//Send handles send command.
func Send(seed, recipient, sender string, amount int64, mwm int64, tag string) {
	seedT, recipientT, senderT := check(seed, recipient, sender, amount)
	ttag, err := gcornode.ToTrytes(tag)
	if err != nil {
		panic(err)
	}
	trs := []gcornode.Transfer{
		gcornode.Transfer{
			Address: recipientT,
			Value:   amount,
			Tag:     ttag,
		},
	}

	var bdl gcornode.Bundle
	server := gcornode.RandomNode()
	fmt.Printf("using IRI server: %s\n", server)

	api := gcornode.NewAPI(server, nil)
	name, pow := gcornode.GetBestPoW()
	fmt.Fprintf(os.Stderr, "using PoW:%s\n", name)
	if senderT == nil {
		bdl, err = gcornode.PrepareTransfers(api, seedT, trs, nil, "", 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		err = gcornode.SendTrytes(api, gcornode.Depth, []gcornode.Transaction(bdl), mwm, pow)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	} else {
		bdl, err = sendToSender(api, trs, senderT, seedT, mwm)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot send: %s\n", err.Error())
		os.Exit(-1)
	}

	fmt.Println("bundle info:")
	fmt.Println("bundle hash: ", bdl.Hash())
	for i, tx := range bdl {
		fmt.Printf(`
		No: %d/%d
		Hash : %s
		Address:%s
		Value:%d
		Timestamp:%s
`,
			i, len(bdl), tx.Hash(), tx.Address, tx.Value, tx.Timestamp)
	}
}
