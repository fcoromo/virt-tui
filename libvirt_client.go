package main

import (
	"fmt"
	"strings"

	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

// VMInfo holds data about a virtual machine
type VMInfo struct {
	ID        int
	Name      string
	UUID      string
	State     string
	CPU       uint
	RAM_MB    uint64
	Disk_GB   float64
	IPAddress string
}

// ConnectLibvirt connects to the local qemu system daemon
func ConnectLibvirt() (*libvirt.Connect, error) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %v", err)
	}
	return conn, nil
}

// FetchVMs retrieves a list of all domains and their info
func FetchVMs(conn *libvirt.Connect) ([]VMInfo, error) {
	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %v", err)
	}

	var vmList []VMInfo

	for _, dom := range domains {
		name, err := dom.GetName()
		if err != nil {
			name = "<unknown>"
		}

		id, _ := dom.GetID()
		uuidStr, _ := dom.GetUUIDString()
		
		stateStr := "Unknown"
		state, _, err := dom.GetState()
		if err == nil {
			switch state {
			case libvirt.DOMAIN_NOSTATE:
				stateStr = "No State"
			case libvirt.DOMAIN_RUNNING:
				stateStr = "Online"
			case libvirt.DOMAIN_BLOCKED:
				stateStr = "Blocked"
			case libvirt.DOMAIN_PAUSED:
				stateStr = "Suspended"
			case libvirt.DOMAIN_SHUTDOWN:
				stateStr = "Shutting Down"
			case libvirt.DOMAIN_SHUTOFF:
				stateStr = "Offline"
			case libvirt.DOMAIN_CRASHED:
				stateStr = "Crashed"
			case libvirt.DOMAIN_PMSUSPENDED:
				stateStr = "PM Suspended"
			}
		}

		var cpuCount uint
		var ramMB uint64
		info, err := dom.GetInfo()
		if err == nil {
			cpuCount = info.NrVirtCpu
			ramMB = info.MaxMem / 1024
		}

		var totalDiskGB float64
		xmlDesc, err := dom.GetXMLDesc(0)
		if err == nil {
			domXML := &libvirtxml.Domain{}
			err = domXML.Unmarshal(xmlDesc)
			if err == nil && domXML.Devices != nil {
				for _, disk := range domXML.Devices.Disks {
					if disk.Target != nil && disk.Target.Dev != "" {
						blockInfo, err := dom.GetBlockInfo(disk.Target.Dev, 0)
						if err == nil {
							totalDiskGB += float64(blockInfo.Capacity) / (1024 * 1024 * 1024)
						}
					}
				}
			}
		}

		var ips []string
		// Try to fetch IP address if the domain is running
		if state == libvirt.DOMAIN_RUNNING {
			// DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE = 0
			// DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT = 1
			ifaces, err := dom.ListAllInterfaceAddresses(0)
			if err != nil || len(ifaces) == 0 {
				ifaces, _ = dom.ListAllInterfaceAddresses(1)
			}
			
			if len(ifaces) > 0 {
				for _, iface := range ifaces {
					for _, addr := range iface.Addrs {
						// 0 is usually IPV4 in libvirt if constant is missing, but let's check length of string
						// or check for "." to safely filter IPv4 since type constants can be tricky.
						if strings.Contains(addr.Addr, ".") && addr.Addr != "127.0.0.1" {
							ips = append(ips, addr.Addr)
						}
					}
				}
			}
		}

		ipAddress := "N/A"
		if len(ips) > 0 {
			ipAddress = strings.Join(ips, ", ")
		}

		vmList = append(vmList, VMInfo{
			ID:        int(id),
			Name:      name,
			UUID:      uuidStr,
			State:     stateStr,
			CPU:       cpuCount,
			RAM_MB:    ramMB,
			Disk_GB:   totalDiskGB,
			IPAddress: ipAddress,
		})

		dom.Free()
	}

	return vmList, nil
}

// VMAction performs a lifecycle action on a domain
func VMAction(conn *libvirt.Connect, uuidStr string, action string) error {
	dom, err := conn.LookupDomainByUUIDString(uuidStr)
	if err != nil {
		return fmt.Errorf("failed to find domain: %v", err)
	}
	defer dom.Free()

	switch action {
	case "start":
		return dom.Create()
	case "stop":
		return dom.Shutdown() // Graceful shutdown
	case "terminate":
		return dom.Destroy()  // Force stop
	case "suspend":
		return dom.Suspend()
	case "resume":
		return dom.Resume()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
