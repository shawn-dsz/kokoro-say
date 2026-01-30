package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

//go:embed web/index.html
var webContent embed.FS

var defaultVoices = []string{
	"af_heart", "af_bella", "af_nicole", "af_sarah", "af_sky",
	"am_adam", "am_michael",
	"bf_emma", "bf_isabella",
	"bm_george", "bm_lewis",
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "web" {
		runWebServer()
		return
	}

	runCLI()
}

func runCLI() {
	voice := flag.String("v", "af_heart", "voice to use")
	speed := flag.Float64("s", 1.0, "speech speed (0.5-2.0)")
	output := flag.String("o", "", "output file (instead of playing)")
	listVoices := flag.Bool("voices", false, "list available voices")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: kokoro-say [options] [text]\n")
		fmt.Fprintf(os.Stderr, "       kokoro-say web [--port PORT]\n\n")
		fmt.Fprintf(os.Stderr, "Convert text to speech using Kokoro TTS.\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  web         Start web interface\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  kokoro-say \"Hello, world!\"\n")
		fmt.Fprintf(os.Stderr, "  echo \"Hello\" | kokoro-say\n")
		fmt.Fprintf(os.Stderr, "  kokoro-say -v bf_emma \"British accent\"\n")
		fmt.Fprintf(os.Stderr, "  kokoro-say -o output.mp3 \"Save to file\"\n")
		fmt.Fprintf(os.Stderr, "  kokoro-say web --port 3000\n")
	}

	flag.Parse()

	if *listVoices {
		fmt.Println("Available voices:")
		for _, v := range defaultVoices {
			fmt.Printf("  %s\n", v)
		}
		return
	}

	text := getText(flag.Args())
	if text == "" {
		fmt.Fprintln(os.Stderr, "No text provided. Pass text as argument or pipe to stdin.")
		os.Exit(1)
	}

	serverURL := getKokoroURL()

	audio, err := synthesize(serverURL, text, *voice, *speed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer audio.Close()

	if *output != "" {
		if err := saveToFile(*output, audio); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving file: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := playAudio(audio); err != nil {
		fmt.Fprintf(os.Stderr, "Error playing audio: %v\n", err)
		os.Exit(1)
	}
}

func runWebServer() {
	webFlags := flag.NewFlagSet("web", flag.ExitOnError)
	port := webFlags.String("port", "3456", "port to listen on")
	noBrowser := webFlags.Bool("no-browser", false, "don't open browser automatically")
	webFlags.Parse(os.Args[2:])

	kokoroURL := getKokoroURL()

	// Serve the embedded HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, _ := webContent.ReadFile("web/index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
	})

	// Proxy TTS requests to Kokoro server (avoids CORS issues)
	http.HandleFunc("/api/speech", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Forward request to Kokoro
		resp, err := http.Post(kokoroURL+"/v1/audio/speech", "application/json", r.Body)
		if err != nil {
			http.Error(w, "Kokoro server unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "audio/mpeg")
		io.Copy(w, resp.Body)
	})

	// Serve voices list
	http.HandleFunc("/api/voices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(defaultVoices)
	})

	addr := ":" + *port
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Printf("Starting web interface at %s\n", url)
	fmt.Println("Press Ctrl+C to stop")

	// Open browser after short delay
	if !*noBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(url)
		}()
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func getKokoroURL() string {
	url := os.Getenv("KOKORO_URL")
	if url == "" {
		url = "http://localhost:8880"
	}
	return url
}

func getText(args []string) string {
	if len(args) > 0 {
		return strings.Join(args, " ")
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}

func synthesize(serverURL, text, voice string, speed float64) (io.ReadCloser, error) {
	payload := map[string]interface{}{
		"input":           text,
		"voice":           voice,
		"speed":           speed,
		"response_format": "mp3",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(serverURL+"/v1/audio/speech", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Kokoro server not found at %s. Is it running?", serverURL)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func saveToFile(path string, audio io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, audio)
	return err
}

func playAudio(audio io.Reader) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("mpv"); err == nil {
			cmd = exec.Command("mpv", "--no-video", "--really-quiet", "-")
		} else if _, err := exec.LookPath("ffplay"); err == nil {
			cmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", "-")
		} else {
			return playWithTempFile(audio, "afplay")
		}
	case "linux":
		if _, err := exec.LookPath("mpv"); err == nil {
			cmd = exec.Command("mpv", "--no-video", "--really-quiet", "-")
		} else if _, err := exec.LookPath("ffplay"); err == nil {
			cmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", "-")
		} else if _, err := exec.LookPath("paplay"); err == nil {
			return playWithTempFile(audio, "paplay")
		} else {
			return playWithTempFile(audio, "aplay")
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return runWithCleanup(cmd, audio)
}

func runWithCleanup(cmd *exec.Cmd, stdin io.Reader) error {
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-sigChan:
		cmd.Process.Kill()
		return nil
	case err := <-done:
		return err
	}
}

func playWithTempFile(audio io.Reader, player string) error {
	f, err := os.CreateTemp("", "kokoro-*.mp3")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := io.Copy(f, audio); err != nil {
		f.Close()
		return err
	}
	f.Close()

	cmd := exec.Command(player, f.Name())
	return runWithCleanup(cmd, nil)
}
