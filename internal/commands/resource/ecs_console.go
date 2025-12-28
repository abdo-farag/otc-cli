package resource

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "os/exec"
  "runtime"
  "time"

  "github.com/abdo-farag/otc-cli/internal/config"
  "github.com/abdo-farag/otc-cli/internal/otc"
  "github.com/fatih/color"
)

// GetConsoleECS gets VNC console URL for an ECS instance and opens it in browser
func GetConsoleECS(cfg *config.Config, client *otc.Client, unscopedToken, projectID, serverNameOrID string, noBrowser bool) {
  projectID, projectToken, err := GetProjectToken(cfg, client, unscopedToken, projectID, false)
  if err != nil {
    color.Red("✗ %v", err)
    return
  }

  color.Yellow("⏳ Finding server...")

  // Get server details to find ID if name provided
  computeURL := fmt.Sprintf("https://ecs.%s.otc.t-systems.com/v1/%s/cloudservers/detail", cfg.Region, projectID)

  body, statusCode, err := MakeRequest(computeURL, projectToken)
  if err != nil {
    color.Red("✗ Request failed: %v", err)
    return
  }

  if statusCode != 200 {
    color.Red("✗ API error (status %d): %s", statusCode, string(body))
    return
  }

  var result struct {
    Servers []struct {
      ID     string `json:"id"`
      Name   string `json:"name"`
      Status string `json:"status"`
    } `json:"servers"`
  }

  if err := json.Unmarshal(body, &result); err != nil {
    color.Red("✗ Failed to parse response: %v", err)
    return
  }

  // Find server by name or ID
  var serverID string
  var serverName string
  var serverStatus string

  for _, s := range result.Servers {
    if s.ID == serverNameOrID || s.Name == serverNameOrID {
      serverID = s.ID
      serverName = s.Name
      serverStatus = s.Status
      break
    }
  }

  if serverID == "" {
    color.Red("✗ Server not found: %s", serverNameOrID)
    return
  }

  color.Cyan("✓ Found server: %s (%s)", serverName, serverID)

  if serverStatus != "ACTIVE" {
    color.Yellow("⚠ Warning: Server status is '%s', console may not be available", serverStatus)
  }

  color.Yellow("⏳ Getting VNC console URL...")

  // Use Nova API v2.6+ endpoint for remote console
  consoleURL := fmt.Sprintf("https://ecs.%s.otc.t-systems.com/v2.1/%s/servers/%s/remote-consoles",
    cfg.Region, projectID, serverID)

  // Request body for VNC console (microversion 2.6+)
  requestBody := map[string]interface{}{
    "remote_console": map[string]string{
      "protocol": "vnc",
      "type":     "novnc",
    },
  }

  jsonBody, _ := json.Marshal(requestBody)

  req, _ := http.NewRequest("POST", consoleURL, bytes.NewBuffer(jsonBody))
  req.Header.Set("X-Auth-Token", projectToken)
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("X-OpenStack-Nova-API-Version", "2.6")

  httpClient := &http.Client{Timeout: 30 * time.Second}
  resp, err := httpClient.Do(req)
  if err != nil {
    color.Red("✗ Request failed: %v", err)
    return
  }
  defer resp.Body.Close()

  respBody, _ := io.ReadAll(resp.Body)

  if resp.StatusCode != 200 {
    color.Red("✗ API error (status %d): %s", resp.StatusCode, string(respBody))
    return
  }

  var consoleResp struct {
    RemoteConsole struct {
      Type     string `json:"type"`
      Protocol string `json:"protocol"`
      URL      string `json:"url"`
    } `json:"remote_console"`
  }

  if err := json.Unmarshal(respBody, &consoleResp); err != nil {
    color.Red("✗ Failed to parse response: %v", err)
    return
  }

  vncURL := consoleResp.RemoteConsole.URL

  // Display console info
  color.Green("✓ VNC Console URL obtained successfully")
  fmt.Printf("\n")
  color.Cyan("Server: %s", serverName)
  color.Cyan("Type: %s", consoleResp.RemoteConsole.Type)
  color.Cyan("Protocol: %s", consoleResp.RemoteConsole.Protocol)
  fmt.Printf("\n")
  color.Green("Console URL:")
  fmt.Printf("%s\n", vncURL)

  // Open in browser unless disabled
  if !noBrowser {
    color.Yellow("\n⏳ Opening console in browser...")
    if err := openBrowser(vncURL); err != nil {
      color.Red("✗ Failed to open browser: %v", err)
      color.Yellow("Please open the URL manually in your browser")
    } else {
      color.Green("✓ Console opened in browser")
    }
  }
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
  var cmd string
  var args []string

  switch runtime.GOOS {
  case "windows":
    cmd = "rundll32"
    args = []string{"url.dll,FileProtocolHandler", url}
  case "darwin":
    cmd = "open"
    args = []string{url}
  default: // linux, freebsd, openbsd, netbsd
    cmd = "xdg-open"
    args = []string{url}
  }

  return exec.Command(cmd, args...).Start()
}
