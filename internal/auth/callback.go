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

		// Render error page
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		h.templates.ExecuteTemplate(w, "error.html", map[string]string{
			"ErrorType": errParam,
			"ErrorDesc": h.errorDesc,
		})

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
		h.templates.ExecuteTemplate(w, "error.html", map[string]string{
			"ErrorType": "missing_code",
			"ErrorDesc": "No authorization code received",
		})

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
	h.templates.ExecuteTemplate(w, "callback.html", map[string]int{"Port": h.port})

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
	fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Complete</title><style>*{margin:0;padding:0}body{font-family:Arial;background:linear-gradient(135deg,#667eea,#764ba2);min-height:100vh;display:flex;align-items:center;justify-content:center}.box{background:#fff;padding:50px;border-radius:20px;text-align:center;box-shadow:0 20px 60px rgba(0,0,0,.3)}.icon{font-size:80px;color:#2ecc71;margin:20px 0}h1{color:#2ecc71;margin:20px 0}p{color:#666;margin:15px 0}button{margin-top:20px;padding:12px 30px;background:#667eea;color:#fff;border:none;border-radius:8px;cursor:pointer;font-size:16px;font-weight:600}button:hover{background:#5568d3;transform:translateY(-2px)}</style></head><body><div class="box"><div class="icon">✓</div><h1>Authentication Complete!</h1><p>Your credentials are ready.</p><p style="font-size:14px">Return to your terminal to continue.</p><button onclick="window.close()">Close Window</button></div><script>window.close();setTimeout(()=>{window.open('','_self');window.close()},100);setTimeout(()=>{if(document.querySelector('.box')){document.querySelector('p:last-of-type').innerHTML='Press <strong>Cmd+W</strong> or <strong>Ctrl+W</strong> to close'}},500)</script></body></html>`)
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
