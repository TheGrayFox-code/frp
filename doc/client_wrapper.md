# FRPC Wrapper (GUI + Linux TUI)

Dieses Repository enthält jetzt einen einfachen Wrapper für `frpc`, um die Nutzung auf Windows und Linux schneller einzurichten.

## Start

```bash
go run ./cmd/frpc-wrapper -mode gui
```

Oder auf Linux im Terminal:

```bash
go run ./cmd/frpc-wrapper -mode tui
```

## Features

- Profile lokal speichern (`~/.frp-wrapper/profiles.json`)
- Profile über GUI anlegen und starten/stoppen
- Linux-TUI für Server-/Terminal-Umgebungen
- Automatische Erzeugung einer temporären `frpc` TOML-Konfiguration

## Voraussetzung

Der Wrapper ruft das Binary `frpc` auf. Stelle sicher, dass `frpc` im `PATH` verfügbar ist.
