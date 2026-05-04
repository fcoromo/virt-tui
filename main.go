package main

import (
	"log"

	"github.com/rivo/tview"
)

func main() {
	// Connect to libvirt
	conn, err := ConnectLibvirt()
	if err != nil {
		log.Fatalf("Error connecting to libvirt: %v\nAre you running on a system with QEMU/KVM and libvirt daemon active?", err)
	}
	// Close connection on exit
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Printf("Error closing libvirt connection: %v\n", err)
		}
	}()

	// Initialize TUI
	app := tview.NewApplication()
	
	// Setup layout and logic
	layout := setupUI(app, conn)

	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error running TUI: %v\n", err)
	}
}
