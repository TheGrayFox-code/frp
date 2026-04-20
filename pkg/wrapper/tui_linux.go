//go:build linux

package wrapper

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func RunTUI(store *Store, runner *Runner) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n=== FRPC Wrapper Linux TUI ===")
		fmt.Println("1) Profile anzeigen")
		fmt.Println("2) Profile speichern")
		fmt.Println("3) Profile starten")
		fmt.Println("4) Stoppen")
		fmt.Println("5) Status")
		fmt.Println("0) Beenden")
		fmt.Print("> ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			profiles, err := store.Load()
			if err != nil {
				fmt.Println("Fehler:", err)
				continue
			}
			if len(profiles) == 0 {
				fmt.Println("Keine Profile vorhanden")
				continue
			}
			for _, p := range profiles {
				fmt.Printf("- %s -> %s:%d (remote %d)\n", p.Name, p.ServerAddr, p.ServerPort, p.RemotePort)
			}
		case "2":
			profile, err := promptProfile(reader)
			if err != nil {
				fmt.Println("Fehler:", err)
				continue
			}
			if err := store.Upsert(profile); err != nil {
				fmt.Println("Fehler:", err)
				continue
			}
			fmt.Println("Profil gespeichert")
		case "3":
			fmt.Print("Profilname: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			profiles, err := store.Load()
			if err != nil {
				fmt.Println("Fehler:", err)
				continue
			}
			found := false
			for _, p := range profiles {
				if p.Name == name {
					if err := runner.Start(p); err != nil {
						fmt.Println("Fehler:", err)
					} else {
						fmt.Println("Gestartet:", name)
					}
					found = true
					break
				}
			}
			if !found {
				fmt.Println("Profil nicht gefunden")
			}
		case "4":
			if err := runner.Stop(); err != nil {
				fmt.Println("Fehler:", err)
			} else {
				fmt.Println("Gestoppt")
			}
		case "5":
			running, name, logs := runner.Status()
			if running {
				fmt.Println("Aktiv:", name)
			} else {
				fmt.Println("Nicht aktiv")
			}
			fmt.Println("--- Logs ---")
			fmt.Println(logs)
		case "0":
			return nil
		default:
			fmt.Println("Unbekannte Auswahl")
		}
	}
}

func promptProfile(reader *bufio.Reader) (Profile, error) {
	read := func(label string) string {
		fmt.Print(label)
		v, _ := reader.ReadString('\n')
		return strings.TrimSpace(v)
	}
	toInt := func(v string) (int, error) {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("invalid number %q", v)
		}
		return n, nil
	}

	serverPort, err := toInt(read("Server Port: "))
	if err != nil {
		return Profile{}, err
	}
	localPort, err := toInt(read("Local Port: "))
	if err != nil {
		return Profile{}, err
	}
	remotePort, err := toInt(read("Remote Port: "))
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		Name:       read("Profilname: "),
		ServerAddr: read("Server Addr: "),
		ServerPort: serverPort,
		AuthToken:  read("Token: "),
		ProxyName:  read("Proxy Name: "),
		ProxyType:  "tcp",
		LocalIP:    read("Local IP: "),
		LocalPort:  localPort,
		RemotePort: remotePort,
	}, nil
}
