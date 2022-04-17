package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Installation struct {
	InstallLocation string
	NamespaceId     string
	ItemId          string
	ArtifactId      string
	AppVersion      string
	AppName         string
}

type Installations struct {
	InstallationList []Installation
}

type CalderaResponse struct {
	Provider string
	Jwt      string
}

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	ntdll                = windows.NewLazySystemDLL("ntdll.dll")
	NtSuspendProcess     = ntdll.NewProc("NtSuspendProcess")
	VirtualAllocEx       = kernel32.NewProc("VirtualAllocEx")
	WriteProcessMemory   = kernel32.NewProc("WriteProcessMemory")
	CreateRemoteThreadEx = kernel32.NewProc("CreateRemoteThreadEx")
	LoadLibraryA         = kernel32.NewProc("LoadLibraryA")
)

func main() {
	var exchange string

	fmt.Print("Enter a private server username: ")
	fmt.Scan(&exchange)

	regex := regexp.MustCompile(("[a-z]"))
	calderaToken, provider := fetchCalderaToken()

	antiCheatService := regex.ReplaceAllString(provider, "")
	var unusedAntiCheatService string

	switch antiCheatService {
	case "EAC":
		unusedAntiCheatService = "be"
	case "BE":
		unusedAntiCheatService = "eac"
	}

	args := []string{
		"-epicapp=Fortnite",
		"-epicenv=Prod",
		"-epicportal",
		"-epiclocale=en-us",
		"-epicsandboxid=fn",
		"-no" + unusedAntiCheatService,
		"-fromfl=" + strings.ToLower(antiCheatService),
		"-caldera=" + calderaToken,
		"-AUTH_LOGIN=" + exchange,
		"-AUTH_PASSWORD=UNUSED",
		"-AUTH_TYPE=epic",
		"-skippatchcheck",
		"-NOSSLPINNING",
	}

	installLocation := fetchInstallLocation()

	pid := launch(installLocation, antiCheatService, args)

	initMods(pid)
}

func launch(installLocation string, antiCheatService string, args []string) uint32 {
	launcherPath := installLocation + "/FortniteGame/Binaries/Win64/FortniteLauncher.exe"
	fortnitePath := installLocation + "/FortniteGame/Binaries/Win64/FortniteClient-Win64-Shipping.exe"
	antiCheatPath := installLocation + "/FortniteGame/Binaries/Win64/FortniteClient-Win64-Shipping_" + antiCheatService + ".exe"

	fmt.Println("Starting Fortnite Launcher...")
	launcherProc := exec.Command(launcherPath, args...)
	launcherStartErr := launcherProc.Start()
	suspendProcess(uint32(launcherProc.Process.Pid))

	fmt.Println("Starting Fortnite Anti-Cheat...")
	antiCheatProc := exec.Command(antiCheatPath, args...)
	antiCheatStartErr := antiCheatProc.Start()
	suspendProcess(uint32(antiCheatProc.Process.Pid))

	fmt.Println("Starting Fortnite Game...")
	cmd := exec.Command(fortnitePath, args...)
	fortniteStartErr := cmd.Start()

	if launcherStartErr != nil {
		fmt.Println("An error occured when starting Fortnite Launcher. Closing.")
		fmt.Println(launcherStartErr)
		os.Exit(0)
	} else if antiCheatStartErr != nil {
		fmt.Println("An error occured when starting Fortnite Anti-Cheat. Closing.")
		fmt.Println(antiCheatStartErr)
		os.Exit(0)
	} else if fortniteStartErr != nil {
		fmt.Println("An error occured when starting Fortnite Game. Closing.")
		fmt.Println(fortniteStartErr)
		os.Exit(0)
	}

	return uint32(cmd.Process.Pid)
}

func fetchCalderaToken() (string, string) {
	rawBody := map[string]bool{"nvidia": true}
	body, err := json.Marshal(rawBody)

	if err != nil {
		fmt.Println(err)
	}

	resp, err := http.Post("https://caldera-service-prod.ecosec.on.epicgames.com/caldera/api/v1/launcher/racp", "application/json", bytes.NewBuffer(body))

	if err != nil {
		fmt.Println(err)
	}

	var res CalderaResponse

	json.NewDecoder(resp.Body).Decode(&res)

	return res.Jwt, res.Provider
}

func inject(pid uint32, file string) {
	handle, _ := windows.OpenProcess(uint32(windows.PROCESS_VM_OPERATION|windows.PROCESS_VM_READ|windows.PROCESS_VM_WRITE|windows.PROCESS_CREATE_THREAD|windows.PROCESS_QUERY_INFORMATION), false, pid)
	defer windows.CloseHandle(handle)

	addr, _ := virtualAllocEx(handle, uintptr(4096), uintptr(windows.MEM_RESERVE|windows.MEM_COMMIT), uintptr(windows.PAGE_READWRITE))

	pathBytes, _ := windows.BytePtrFromString(file)

	writeProcessMemory(handle, addr, pathBytes, uintptr(len(file)))

	createRemoteThreadEx(handle, LoadLibraryA.Addr(), addr)

	fmt.Printf("Sucessfully injected dll file: %v to fortnite process!", file)
}

func initMods(pid uint32) {
	files, err := ioutil.ReadDir("./dlls/")

	if err != nil {
		fmt.Println("DLL folder not found. Creating...")
		os.Mkdir("dlls", os.ModePerm)
	}

	for _, file := range files {
		inject(pid, fmt.Sprintf("./dlls/%v", file.Name()))
	}
}

func fetchInstallLocation() string {
	data, err := ioutil.ReadFile("C:/ProgramData/Epic/UnrealEngineLauncher/LauncherInstalled.dat")

	if err != nil {
		fmt.Println(err)
	}

	var installations Installations

	json.Unmarshal([]byte(data), &installations)

	for _, installation := range installations.InstallationList {
		if installation.AppName == "Fortnite" {
			return installation.InstallLocation
		}
	}

	return ""
}

func suspendProcess(pid uint32) error {
	handle, err := windows.OpenProcess(windows.PROCESS_SUSPEND_RESUME, false, pid)
	if err != nil {
		return err
	}

	defer windows.CloseHandle(handle)

	if r1, _, _ := NtSuspendProcess.Call(uintptr(handle)); r1 != 0 {
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/596a1078-e883-4972-9bbc-49e60bebca55 provides more details on what the status code means
		return fmt.Errorf("NtStatus='0x%.8X'", r1)
	}

	return nil
}

func virtualAllocEx(pHandle windows.Handle, size, allocType, allocProt uintptr) (uintptr, error) {
	addr, _, err := VirtualAllocEx.Call(
		uintptr(pHandle),
		uintptr(0),
		uintptr(size),
		uintptr(allocType),
		uintptr(allocProt),
	)

	if addr == 0 {
		return 0, err
	}

	return addr, nil
}

func writeProcessMemory(pHandle windows.Handle, addr uintptr, path *byte, len uintptr) (ret uintptr, err error) {
	ret, _, err = WriteProcessMemory.Call(
		uintptr(pHandle),
		uintptr(addr),
		uintptr(unsafe.Pointer(path)),
		uintptr(len),
		uintptr(0),
	)

	if ret == 0 {
		return 0, err
	}

	return ret, nil
}

func createRemoteThreadEx(pHandle windows.Handle, remoteProcAddr, argAddr uintptr) (handle uintptr, err error) {
	handle, _, err = CreateRemoteThreadEx.Call(
		uintptr(pHandle),
		uintptr(0),
		uintptr(0),
		remoteProcAddr,
		argAddr,
		uintptr(0),
		uintptr(0),
	)

	return handle, err
}
