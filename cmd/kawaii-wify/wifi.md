# FortiGate Captive Portal Discovery & Command History Log

## 1. Environment & Target Specifications

* **Firewall Target:** `[https://fw.bits-pilani.ac.in:8090/](https://fw.bits-pilani.ac.in:8090/)`
* **WAN Probe Endpoint:** `[http://connectivitycheck.gstatic.com/generate_204](http://connectivitycheck.gstatic.com/generate_204)`
* **Subnet Identified:** `172.17.x.x` (e.g., interface `wlan0` leasing `172.17.71.59` and `172.17.71.141`)
* **Target SSID Profile UUID:** `36ccb0fb-ca0d-42f4-996e-785b3bd64342` (`BITS-STAFF` / `BITS-PILANI`)

---

## 2. Reverse Engineering Journey & Command Ledger

### Step 1: Forcing a Clean Captive Lease via MAC Randomization

* **Objective:** Invalidate the authenticated MAC session on the firewall to drop back into the captive trap without waiting for session timeouts.
* **Commands Run:**
```fish
nmcli connection modify '36ccb0fb-ca0d-42f4-996e-785b3bd64342' 802-11-wireless.cloned-mac-address random
nmcli connection down '36ccb0fb-ca0d-42f4-996e-785b3bd64342'; and nmcli connection up '36ccb0fb-ca0d-42f4-996e-785b3bd64342'

```


* **Observation:** Connection successfully cycled, local IPv4 changed from `172.17.71.59` to `172.17.71.141`.

---

### Step 2: Extracting the Challenge Token ($T$)

* **Attempt 1 (HTTP HEAD Request):**
```fish
set TOKEN (curl -s -I http://connectivitycheck.gstatic.com/generate_204 | grep -i "location:" | string match -r 'fgtauth\?([a-f0-9]+)' | tail -n 1)

```


* **Result:** `$TOKEN` was empty.
* **Diagnosis via Verbose Output:** Running `curl -v -o /dev/null [http://connectivitycheck.gstatic.com/generate_204](http://connectivitycheck.gstatic.com/generate_204)` revealed:
* The server returned `HTTP/1.1 200 OK` (Content-Length: 142) instead of a standard `302 Found`.
* FortiOS does not inject an HTTP `Location` redirect header; it intercepts the TCP port 80 stream and injects an HTML body containing JavaScript redirection.




* **Attempt 2 (Raw HTML Body Inspection):**
```fish
curl -s http://connectivitycheck.gstatic.com/generate_204

```


* **Payload Captured:**
```html
<html><body><script language="JavaScript">window.location="https://fw.bits-pilani.ac.in:8090/fgtauth?161cbec86122b347";</script></body></html>

```




* **Resolution (Body Regex Parser):**
```fish
set TOKEN (curl -s http://connectivitycheck.gstatic.com/generate_204 | string match -rg 'fgtauth\?([a-f0-9]+)')

```


* **Result:** Successfully captured the dynamic 16-character hexadecimal token (e.g., `1613bbc47f823993`).



---

### Step 3: Resolving Connection Resets (`curl: (52) Empty reply from server`)

* **Attempt 1 (Direct POST with Token):**
```fish
curl -k -i -X POST "https://fw.bits-pilani.ac.in:8090/" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data "4Tredir=https%3A%2F%2Fconnectivitycheck.gstatic.com%2Fgenerate_204" \
  --data "magic=$TOKEN" \
  --data "username=F20230814" \
  --data-urlencode "password=F20237057%23"

```


* **Result:** `curl: (52) Empty reply from server`.


* **Root Cause Analysis:**
1. **ALPN / HTTP/2 Reset:** Modern `curl` negotiates HTTP/2 by default over TLS ALPN. The embedded micro-server on FortiGate's port 8090 resets TLS sockets when HTTP/2 is offered.
2. **Uninitialized Session State (Missing Prime Step):** FortiOS generates the challenge token during the probe intercept, but marks the session pending. If a client attempts to `POST` without first fetching `GET /fgtauth?$TOKEN`, the firewall drops the connection.
3. **Header Validation:** The firewall verifies the `Referer` header matches `[https://fw.bits-pilani.ac.in:8090/fgtauth?$TOKEN](https://fw.bits-pilani.ac.in:8090/fgtauth?$TOKEN)` and expects a standard browser `User-Agent`.
4. **URL Encoding Defect:** Passing `%23` inside `--data-urlencode` caused `curl` to double-encode the `#` symbol to `%2523`. Raw characters must be passed when using `--data-urlencode`.
5. **Token Race Condition:** Background OS connectivity daemons hitting `generate_204` regenerate and invalidate single-use challenge tokens. Token fetch, session priming, and auth submission must occur in an uninterrupted sequence.



---

## 3. Verified Working Protocol Scripts

### Working Atomic Authentication Sequence

```fish
begin
  # 1. Fetch single-use challenge token from HTML body
  set -l T (curl -s http://connectivitycheck.gstatic.com/generate_204 | string match -rg 'fgtauth\?([a-f0-9]+)')
  echo "Acquired Token: $T"

  # 2. Prime the challenge session via HTTP/1.1 (registers state in FortiOS)
  curl -k --http1.1 -s -A "Mozilla/5.0" "https://fw.bits-pilani.ac.in:8090/fgtauth?$T" > /dev/null

  # 3. Post credentials with mandatory Referer and HTTP/1.1 enforcement
  curl -k --http1.1 -i -X POST "https://fw.bits-pilani.ac.in:8090/" \
    -A "Mozilla/5.0" \
    -e "https://fw.bits-pilani.ac.in:8090/fgtauth?$T" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    --data "4Tredir=http%3A%2F%2Fconnectivitycheck.gstatic.com%2Fgenerate_204" \
    --data "magic=$T" \
    --data "username=F20230814" \
    --data-urlencode "password=F20237057#"
end

```

### Working Session Teardown (Logout)

```fish
curl -k --http1.1 -s -A "Mozilla/5.0" "https://fw.bits-pilani.ac.in:8090/logout?$T" > /dev/null

```

### WAN Routing Verification

```fish
curl -s -o /dev/null -w "%{http_code}\n" http://connectivitycheck.gstatic.com/generate_204

```

* **Observed Output:** `204` (confirmed unblocked routing).

---

## 4. Technical Rules for the Go Engine Implementation

| Behavioral Requirement | Low-Level Go Mechanism |
| --- | --- |
| **Enforce HTTP/1.1** | Configure `http.Transport` with `TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper)` to suppress ALPN negotiation. |
| **Bypass Self-Signed Cert Safely** | Use `crypto/tls` with `InsecureSkipVerify: true` coupled to a `VerifyConnection` callback that matches `cs.ServerName == "fw.bits-pilani.ac.in"`. |
| **Token Parsing** | Read response stream into memory and run `regexp.MustCompile(`fgtauth?([a-f0-9]+)`)`. |
| **Priming Handshake** | Issue `http.NewRequestWithContext` using method `GET` against `/fgtauth?<token>` with header `User-Agent: Mozilla/5.0` prior to login. |
| **Payload Encoding** | Use standard `net/url.Values{}` to handle character escaping for usernames and passwords automatically. |
| **Referer Pinning** | Explicitly set `req.Header.Set("Referer", "[https://fw.bits-pilani.ac.in:8090/fgtauth](https://fw.bits-pilani.ac.in:8090/fgtauth)?"+token)` on the `POST` request. |
| **Session Teardown** | Persist active session token ($T$) so termination handlers (`SIGINT`/`SIGTERM`) can execute `GET /logout?<token>`. |

---

