package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"libvirt.org/go/libvirt"
)

func setupUI(app *tview.Application, conn *libvirt.Connect) *tview.Flex {
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)

	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [yellow]s[-]: Start  [yellow]p[-]: Stop(Graceful)  [yellow]t[-]: Terminate(Force)  [yellow]z[-]: Suspend  [yellow]w[-]: Resume  [yellow]r[-]: Refresh  [yellow]q[-]: Quit")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(helpText, 1, 0, false)

	// Keep track of VMs to retrieve UUID when an action is selected
	var currentVMs []VMInfo

	refreshTable := func() {
		table.Clear()
		table.SetCell(0, 0, tview.NewTableCell("ID").SetTextColor(tcell.ColorYellow).SetSelectable(false))
		table.SetCell(0, 1, tview.NewTableCell("Name").SetTextColor(tcell.ColorYellow).SetSelectable(false))
		table.SetCell(0, 2, tview.NewTableCell("State").SetTextColor(tcell.ColorYellow).SetSelectable(false))
		table.SetCell(0, 3, tview.NewTableCell("CPU").SetTextColor(tcell.ColorYellow).SetSelectable(false))
		table.SetCell(0, 4, tview.NewTableCell("RAM (MB)").SetTextColor(tcell.ColorYellow).SetSelectable(false))
		table.SetCell(0, 5, tview.NewTableCell("Disk (GB)").SetTextColor(tcell.ColorYellow).SetSelectable(false))
		table.SetCell(0, 6, tview.NewTableCell("IP Address").SetTextColor(tcell.ColorYellow).SetSelectable(false))

		vms, err := FetchVMs(conn)
		if err != nil {
			table.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("Error: %v", err)).SetTextColor(tcell.ColorRed))
			return
		}
		currentVMs = vms

		for i, vm := range vms {
			idStr := strconv.Itoa(vm.ID)
			if vm.ID == -1 || vm.ID == 4294967295 { // libvirt max uint32 for inactive
				idStr = "-"
			}
			stateColor := tcell.ColorWhite
			if vm.State == "Online" {
				stateColor = tcell.ColorGreen
			} else if vm.State == "Offline" {
				stateColor = tcell.ColorGray
			} else if vm.State == "Suspended" {
				stateColor = tcell.ColorYellow
			}

			table.SetCell(i+1, 0, tview.NewTableCell(idStr))
			table.SetCell(i+1, 1, tview.NewTableCell(vm.Name).SetTextColor(tcell.ColorAqua))
			table.SetCell(i+1, 2, tview.NewTableCell(vm.State).SetTextColor(stateColor))
			table.SetCell(i+1, 3, tview.NewTableCell(strconv.Itoa(int(vm.CPU))))
			table.SetCell(i+1, 4, tview.NewTableCell(strconv.FormatUint(vm.RAM_MB, 10)))
			table.SetCell(i+1, 5, tview.NewTableCell(fmt.Sprintf("%.1f", vm.Disk_GB)))
			table.SetCell(i+1, 6, tview.NewTableCell(vm.IPAddress))
		}
	}

	refreshTable()

	showError := func(msg string) {
		modal := tview.NewModal().
			SetText(msg).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.SetRoot(layout, true)
			})
		app.SetRoot(modal, true)
	}

	confirmAction := func(action string, vmName string, uuidStr string) {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Are you sure you want to %s VM '%s'?", action, vmName)).
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					err := VMAction(conn, uuidStr, action)
					if err != nil {
						showError(fmt.Sprintf("Failed to %s: %v", action, err))
						return
					}
					refreshTable()
				}
				app.SetRoot(layout, true)
			})
		app.SetRoot(modal, true)
	}

	// Set input capture for table actions
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
			return nil
		}
		if event.Rune() == 'r' {
			refreshTable()
			return nil
		}

		row, _ := table.GetSelection()
		if row < 1 || row > len(currentVMs) {
			return event
		}

		selectedVM := currentVMs[row-1]

		switch event.Rune() {
		case 's':
			confirmAction("start", selectedVM.Name, selectedVM.UUID)
			return nil
		case 'p':
			confirmAction("stop", selectedVM.Name, selectedVM.UUID)
			return nil
		case 't':
			confirmAction("terminate", selectedVM.Name, selectedVM.UUID)
			return nil
		case 'z':
			confirmAction("suspend", selectedVM.Name, selectedVM.UUID)
			return nil
		case 'w':
			confirmAction("resume", selectedVM.Name, selectedVM.UUID)
			return nil
		}
		return event
	})

	return layout
}
