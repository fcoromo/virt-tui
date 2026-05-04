# virt-tui

`virt-tui` is a high-performance Linux terminal user interface (TUI) application designed to manage local QEMU/KVM virtual machines. 

Built with Go, `tview`, and the official `libvirt` bindings, it connects to your local hypervisor (`qemu:///system`) and provides an easy-to-use, interactive dashboard to view and control your VMs without needing a heavy graphical interface like Virt-Manager.

## Features
* **Live Dashboard**: Displays all active and inactive domains.
* **Hardware Overview**: See the provisioned CPU count, RAM (MB), and total Disk capacity (GB) for each VM.
* **Network Details**: Automatically fetches and lists assigned IPv4 addresses (excluding localhost).
* **Lifecycle Management**: Start, stop (graceful shutdown), suspend, resume, and terminate (force stop) VMs directly from your keyboard.

---

## Prerequisites

Because `virt-tui` relies on CGo to interact with the native `libvirt` C API, you must compile it on a Linux system with the appropriate development headers installed.

1. **Go Compiler**: You need Go installed on your system.
2. **C Compiler**: Such as `gcc`.
3. **Libvirt Headers**: The development packages for libvirt.

**On Ubuntu / Debian:**
```bash
sudo apt-get update
sudo apt-get install golang gcc libvirt-dev
```

**On RHEL / Fedora / CentOS:**
```bash
sudo dnf install golang gcc libvirt-devel
```

---

## Compilation

1. Clone or copy the repository to your Linux machine.
2. Navigate to the project directory:
   ```bash
   cd virt-tui
   ```
3. Download the dependencies and tidy the module:
   ```bash
   go mod tidy
   ```
4. Build the binary:
   ```bash
   go build -o virt-tui .
   ```

---

## Usage

Running the application usually requires `root` privileges or membership in the `libvirt` user group in order to successfully connect to the `qemu:///system` daemon.

Start the application:
```bash
sudo ./virt-tui
```

### Keyboard Controls

Once the TUI is running, you can navigate the table using your **Up/Down arrow keys** or your **Mouse**. When a VM is selected, use the following shortcuts:

* **`s`**: Start the selected VM
* **`p`**: Stop the VM (triggers an ACPI Graceful Shutdown)
* **`t`**: Terminate the VM (Force Destroys the VM immediately)
* **`z`**: Suspend the VM (pauses execution)
* **`w`**: Resume a suspended VM
* **`r`**: Manually refresh the VM list and metrics
* **`q`**: Quit the application
