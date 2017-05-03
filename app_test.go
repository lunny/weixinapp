package weixinapp

import (
	"os"
	"testing"
)

func TestGenQR(t *testing.T) {
	// TODO:
	app := NewAPP("xxxx", "xxxx")
	os.Remove("./qr.png")
	f, err := os.Create("./qr.png")
	if err != nil {
		t.Fatal(err)
	}

	err = app.CreateQRCode("pages/index/index", 430, f)
	if err != nil {
		t.Fatal(err)
	}
}
