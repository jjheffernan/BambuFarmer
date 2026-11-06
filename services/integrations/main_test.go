package main

import "testing"

func TestStatusURL(t *testing.T) {
	off := statusURL("")
	if off["configured"] != false {
		t.Fatalf("%v", off)
	}
	on := statusURL("http://spoolman:7912")
	if on["configured"] != true || on["url"] != "http://spoolman:7912" {
		t.Fatalf("%v", on)
	}
}
