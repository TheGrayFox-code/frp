package wrapper

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"runtime"
)

type GUI struct {
	store  *Store
	runner *Runner
}

func NewGUI(store *Store, runner *Runner) *GUI {
	return &GUI{store: store, runner: runner}
}

func (g *GUI) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", g.handleIndex)
	mux.HandleFunc("/api/profiles", g.handleProfiles)
	mux.HandleFunc("/api/start", g.handleStart)
	mux.HandleFunc("/api/stop", g.handleStop)
	mux.HandleFunc("/api/status", g.handleStatus)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	url := "http://" + ln.Addr().String()
	_ = openBrowser(url)
	fmt.Printf("FRP Wrapper GUI läuft unter %s\n", url)
	return http.Serve(ln, mux)
}

func (g *GUI) handleIndex(w http.ResponseWriter, _ *http.Request) {
	t := template.Must(template.New("index").Parse(indexHTML))
	_ = t.Execute(w, nil)
}

func (g *GUI) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		profiles, err := g.store.Load()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, profiles)
	case http.MethodPost:
		var p Profile
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := g.store.Upsert(p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (g *GUI) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Profile string `json:"profile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	profiles, err := g.store.Load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, p := range profiles {
		if p.Name == req.Profile {
			if err := g.runner.Start(p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "profile not found", http.StatusNotFound)
}

func (g *GUI) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := g.runner.Stop(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (g *GUI) handleStatus(w http.ResponseWriter, _ *http.Request) {
	running, profile, logs := g.runner.Status()
	writeJSON(w, map[string]any{
		"running": running,
		"profile": profile,
		"logs":    logs,
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

const indexHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>FRP Client Wrapper</title>
  <style>
    body { font-family: sans-serif; margin: 2rem; }
    label { display:block; margin-top: .5rem; }
    input { width: 360px; padding: .4rem; }
    button { margin-top: .8rem; padding: .45rem .7rem; }
    .row { margin-top: 1rem; }
    pre { background: #111; color: #8f8; padding: 1rem; height: 200px; overflow:auto; }
  </style>
</head>
<body>
<h2>FRP Client Wrapper</h2>
<p>Einfaches Profil-Management für frpc auf Windows/Linux.</p>
<div>
  <label>Profilname <input id="name" /></label>
  <label>Server Adresse <input id="serverAddr" value="example.com" /></label>
  <label>Server Port <input id="serverPort" type="number" value="7000" /></label>
  <label>Token <input id="authToken" /></label>
  <label>Proxy Name <input id="proxyName" value="rdp" /></label>
  <label>Lokale IP <input id="localIP" value="127.0.0.1" /></label>
  <label>Lokaler Port <input id="localPort" type="number" value="3389" /></label>
  <label>Remote Port <input id="remotePort" type="number" value="6000" /></label>
  <button onclick="saveProfile()">Profil speichern</button>
</div>
<div class="row">
  <select id="profiles"></select>
  <button onclick="startProfile()">Starten</button>
  <button onclick="stopProfile()">Stoppen</button>
  <button onclick="refresh()">Aktualisieren</button>
</div>
<div class="row" id="status"></div>
<pre id="logs"></pre>
<script>
async function api(path, method='GET', body=null) {
  const res = await fetch(path, {method, headers:{'content-type':'application/json'}, body: body?JSON.stringify(body):null});
  if (!res.ok) throw new Error(await res.text());
  if (res.status === 204) return null;
  return await res.json();
}
async function loadProfiles() {
  const profiles = await api('/api/profiles');
  const sel = document.getElementById('profiles');
  sel.innerHTML = '';
  for (const p of profiles) {
    const o = document.createElement('option');
    o.value = p.name; o.textContent = p.name;
    sel.appendChild(o);
  }
}
async function saveProfile() {
  const p = {
    name: name.value,
    serverAddr: serverAddr.value,
    serverPort: Number(serverPort.value),
    authToken: authToken.value,
    proxyName: proxyName.value,
    proxyType: 'tcp',
    localIP: localIP.value,
    localPort: Number(localPort.value),
    remotePort: Number(remotePort.value),
  };
  await api('/api/profiles', 'POST', p);
  await loadProfiles();
}
async function startProfile() {
  await api('/api/start', 'POST', {profile: profiles.value});
  await refresh();
}
async function stopProfile() {
  await api('/api/stop', 'POST');
  await refresh();
}
async function refresh() {
  await loadProfiles();
  const st = await api('/api/status');
  status.textContent = st.running ? ('Läuft: ' + st.profile) : 'Nicht aktiv';
  logs.textContent = st.logs || '';
}
setInterval(refresh, 2000);
refresh();
</script>
</body></html>`
