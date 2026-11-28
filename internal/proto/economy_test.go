package proto

import (
	"bytes"
	"testing"
)

func TestEconomyQuoteRoundTrip(t *testing.T) {
	quote := EconomySettings{Revision: 3, BuyPercent: 150, SellPercent: 80, RerollGold: 75}
	got, err := DecodeEconomySettings(EconomySettingsPacket(quote))
	if err != nil || got != quote {
		t.Fatalf("quote round trip: %+v, %v", got, err)
	}
	action := ShopBuyPacket(40, 3, 2)
	rev, unit, body, err := DecodeEconomyAction(EconomyActionPacket(quote.Revision, 150, action))
	if err != nil || rev != quote.Revision || unit != 150 || !bytes.Equal(body, action) {
		t.Fatalf("action round trip: revision %d, unit %d, body %x, error %v", rev, unit, body, err)
	}
	if price := ScaleVendorUnitPrice(99, 150); price != 148 {
		t.Fatalf("scaled unit price = %d, want 148", price)
	}
}
