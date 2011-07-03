package main

import (
	"netgrowl"
)

func main() {
	notifications := []string{
		"NetGrowl example notification",
	}

	growl := netgrowl.NewNetGrowl(netgrowl.DefaultAddress, "netgrowl", notifications, "password")
	if err := growl.Register(); err != nil {
		panic(err)
	}
	if err := growl.Notify("NetGrowl example notification", "Hello", "...world!", netgrowl.PriorityNormal, false); err != nil {
		panic(err)
	}
}
