package blockchain

import (
	"fmt"
	"strings"
	"testing"
)

const testNodeId = "test"

type TestSetup struct {
	client       *Client
	creationAddr string
}

func setup() TestSetup {
	client := NewClient(testNodeId)
	addr := client.CreateWallet()
	client.CreateBlockchain(addr)

	setup := TestSetup{client, addr}
	return setup
}

func TestListAddresses(t *testing.T) {
	setup := setup()

	actual := setup.client.ListAddresses()
	expected := []string{setup.creationAddr}
	if compareAddresses(actual, expected) == false {
		t.Errorf("List addresses failed: %s", getAddressErrorMessage(actual, expected))
	}

	newAddr := setup.client.CreateWallet()
	actual = setup.client.ListAddresses()
	expected = []string{setup.creationAddr, newAddr}
	if compareAddresses(actual, expected) == false {
		t.Errorf("List addresses failed: %s", getAddressErrorMessage(actual, expected))
	}

	tearDown()
}

func compareAddresses(actual []string, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	counts := make(map[string]int, len(expected))
	for _, addr := range expected {
		counts[addr]++
	}

	for _, addr := range actual {
		counts[addr]--
		if counts[addr] < 0 {
			return false
		}
	}

	return true
}

func getAddressErrorMessage(actual []string, expected []string) string {
	var b strings.Builder

	b.WriteString("Actual:\n")
	for _, a := range actual {
		fmt.Fprintf(&b, "  - %s\n", a)
	}

	b.WriteString("Expected:\n")
	for _, e := range expected {
		fmt.Fprintf(&b, "  - %s\n", e)
	}

	return b.String()
}

func TestSendSuccess(t *testing.T) {
	setup := setup()

	addr := setup.client.CreateWallet()
	assertBalance(t, setup.client, setup.creationAddr, 10)
	assertBalance(t, setup.client, addr, 0)

	setup.client.Send(setup.creationAddr, addr, 6)
	assertBalance(t, setup.client, setup.creationAddr, 10-6+10) // 14
	assertBalance(t, setup.client, addr, 6)

	secondAddr := setup.client.CreateWallet()
	setup.client.Send(addr, setup.creationAddr, 5)
	setup.client.Send(setup.creationAddr, secondAddr, 13)
	assertBalance(t, setup.client, setup.creationAddr, 14+5-13+10) // 16
	assertBalance(t, setup.client, addr, 6-5+10)                   // 11
	assertBalance(t, setup.client, secondAddr, 13)

	tearDown()
}

func TestSendFailed(t *testing.T) {
	setup := setup()

	addr := setup.client.CreateWallet()
	setup.client.Send(setup.creationAddr, addr, 11)
	assertBalance(t, setup.client, setup.creationAddr, 10)
	assertBalance(t, setup.client, addr, 0)

	tearDown()
}

func assertBalance(t *testing.T, client *Client, address string, expected int) {
	actual, _ := client.GetBalance(address)
	if actual != expected {
		t.Errorf("Wrong balance on '%s': Expected %d, Got %d", address, expected, actual)
	}
}

func tearDown() {
	Destroy(testNodeId)
	walletRepo := NewWalletRepo(testNodeId)
	walletRepo.Destroy()
}
