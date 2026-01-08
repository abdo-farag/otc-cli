package auth

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"
)

type CallbackHandler struct {
	code, state, errorType, errorDesc, validationStatus, validationMessage string
	mu                                                                     sync.RWMutex
	codeChan                                                               chan string
	errorChan                                                              chan string
	server                                                                 *http.Server
	port                                                                   int
	templates                                                              *template.Template
}

func NewCallbackHandler(port int, templatesFS embed.FS) *CallbackHandler {
	return &CallbackHandler{
		codeChan:  make(chan string, 1),
		errorChan: make(chan string, 1),
		port:      port,
		templates: template.Must(template.ParseFS(templatesFS, "templates/*.html")),
	}
}

func (h *CallbackHandler) StartServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/oidc/auth", h.handleCallback)
	mux.HandleFunc("/status", h.handleStatus)
	mux.HandleFunc("/close", h.handleClose)

	h.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go h.server.ListenAndServe()
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (h *CallbackHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Check for OAuth errors first
	if errParam := q.Get("error"); errParam != "" {
		h.mu.Lock()
		h.errorType = errParam
		h.errorDesc = q.Get("error_description")
		h.mu.Unlock()

		// Return inline error HTML (no template needed for OAuth errors)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		errorHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OAuth Error - OTC SSO</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box}
        body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#e74c3c,#c0392b);min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
        .box{background:#fff;padding:50px;border-radius:20px;text-align:center;max-width:500px;box-shadow:0 20px 60px rgba(0,0,0,0.3);animation:slideUp 0.4s ease-out}
        @keyframes slideUp{from{opacity:0;transform:translateY(20px)}to{opacity:1;transform:translateY(0)}}
        .icon{font-size:70px;color:#e74c3c;margin-bottom:20px}
        h1{color:#e74c3c;margin-bottom:20px;font-size:28px;font-weight:600}
        .error-code{background:#ffebee;color:#c62828;padding:8px 16px;border-radius:6px;display:inline-block;margin:10px 0;font-family:monospace;font-size:14px}
        p{color:#666;margin:15px 0;line-height:1.6}
        .close-hint{background:#f0f0f0;padding:15px;border-radius:8px;margin-top:30px;border:2px dashed #999}
        kbd{background:#f4f4f4;border:1px solid #ccc;padding:3px 7px;border-radius:3px;font-family:monospace;font-size:12px}
    </style>
</head>
<body>
    <div class="box">
        <div class="icon">✗</div>
        <h1>OAuth Authentication Error</h1>
        <div class="error-code">%s</div>
        <p>%s</p>
        <div class="close-hint">
            <p style="margin:0;color:#666;font-size:14px">Close this tab: <kbd>Ctrl+W</kbd> or <kbd>Cmd+W</kbd></p>
        </div>
    </div>
    <script>
        // Attempt to close window
        window.addEventListener('load', function() {
            window.close();
        });
    </script>
</body>
</html>`, errParam, h.errorDesc)

		fmt.Fprint(w, errorHTML)

		// Send error to channel
		select {
		case h.errorChan <- errParam:
		default:
		}
		return
	}

	// Get authorization code
	code := q.Get("code")
	if code == "" {
		// Missing code error
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		errorHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OAuth Error - OTC SSO</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box}
        body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#e74c3c,#c0392b);min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
        .box{background:#fff;padding:50px;border-radius:20px;text-align:center;max-width:500px;box-shadow:0 20px 60px rgba(0,0,0,0.3);animation:slideUp 0.4s ease-out}
        @keyframes slideUp{from{opacity:0;transform:translateY(20px)}to{opacity:1;transform:translateY(0)}}
        .icon{font-size:70px;color:#e74c3c;margin-bottom:20px}
        h1{color:#e74c3c;margin-bottom:20px;font-size:28px;font-weight:600}
        .error-code{background:#ffebee;color:#c62828;padding:8px 16px;border-radius:6px;display:inline-block;margin:10px 0;font-family:monospace;font-size:14px}
        p{color:#666;margin:15px 0;line-height:1.6}
        .close-hint{background:#f0f0f0;padding:15px;border-radius:8px;margin-top:30px;border:2px dashed #999}
        kbd{background:#f4f4f4;border:1px solid #ccc;padding:3px 7px;border-radius:3px;font-family:monospace;font-size:12px}
    </style>
</head>
<body>
    <div class="box">
        <div class="icon">✗</div>
        <h1>OAuth Authentication Error</h1>
        <div class="error-code">missing_code</div>
        <p>No authorization code received from the identity provider</p>
        <div class="close-hint">
            <p style="margin:0;color:#666;font-size:14px">Close this tab: <kbd>Ctrl+W</kbd> or <kbd>Cmd+W</kbd></p>
        </div>
    </div>
    <script>
        window.addEventListener('load', function() {
            window.close();
        });
    </script>
</body>
</html>`

		fmt.Fprint(w, errorHTML)

		select {
		case h.errorChan <- "missing_code":
		default:
		}
		return
	}

	// Success - store code and render callback page
	h.mu.Lock()
	h.code = code
	h.state = q.Get("state")
	h.mu.Unlock()

	// Render success callback page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.templates.ExecuteTemplate(w, "callback.html", map[string]interface{}{"Port": h.port})

	// Send code to channel
	select {
	case h.codeChan <- code:
	default:
	}
}

func (h *CallbackHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	status := h.validationStatus
	message := h.validationMessage
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	fmt.Fprintf(w, `{"status":"%s","message":"%s"}`, status, message)
}

func (h *CallbackHandler) handleClose(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authentication Complete - OTC SSO</title>
    <style>
        *{margin:0;padding:0;box-sizing:border-box}
        body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea,#764ba2);min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
        .box{background:#fff;padding:50px;border-radius:20px;text-align:center;box-shadow:0 20px 60px rgba(0,0,0,0.3);max-width:500px;animation:slideUp 0.4s ease-out}
        @keyframes slideUp{from{opacity:0;transform:translateY(20px)}to{opacity:1;transform:translateY(0)}}
        .icon{font-size:80px;color:#2ecc71;margin:20px 0;animation:iconPop 0.5s ease-out}
        @keyframes iconPop{0%{transform:scale(0)}50%{transform:scale(1.1)}100%{transform:scale(1)}}
        h1{color:#2ecc71;margin:20px 0;font-size:28px;font-weight:600}
        p{color:#666;margin:15px 0;line-height:1.6}
        .close-hint{background:#f8f9fa;padding:20px;border-radius:10px;margin-top:20px;border:2px dashed #999}
        kbd{background:#f4f4f4;border:1px solid #ccc;border-radius:3px;padding:3px 7px;font-family:monospace;font-size:12px;box-shadow:0 1px 2px rgba(0,0,0,0.1)}
    </style>
</head>
<body>
    <div class="box">
        <div class="icon">✓</div>
        <h1>Authentication Complete!</h1>
        <p>Your credentials are ready.</p>
        <p style="font-size:14px;color:#999">Return to your terminal to continue.</p>
        <div class="close-hint">
            <p style="margin:0;color:#666;font-size:14px">Close this tab: <kbd>Ctrl+W</kbd> or <kbd>Cmd+W</kbd></p>
        </div>
    </div>
    <script>
        window.addEventListener('load', function() {
            window.close();
            setTimeout(function() {
                window.open('', '_self');
                window.close();
            }, 100);
        });
    </script>
</body>
</html>`)
}

// SetValidationStatus updates the validation status shown in the browser
// This should be called from the CLI after token exchange succeeds/fails
func (h *CallbackHandler) SetValidationStatus(status, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.validationStatus = status
	h.validationMessage = message
}

// WaitForCode waits for the authorization code from the callback
func (h *CallbackHandler) WaitForCode(timeout time.Duration) (string, error) {
	select {
	case code := <-h.codeChan:
		return code, nil
	case err := <-h.errorChan:
		return "", fmt.Errorf("OAuth error: %s", err)
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for callback")
	}
}

// Close gracefully shuts down the callback server
func (h *CallbackHandler) Close() error {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return h.server.Shutdown(ctx)
	}
	return nil
}
