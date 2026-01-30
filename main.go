package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var defaultVoices = []string{
	"af_heart", "af_bella", "af_nicole", "af_sarah", "af_sky",
	"am_adam", "am_michael",
	"bf_emma", "bf_isabella",
	"bm_george", "bm_lewis",
}

func main() {
	voice := flag.String("v", "af_heart", "voice to use")
	speed := flag.Float64("s", 1.0, "speech speed (0.5-2.0)")
	output := flag.String("o", "", "output file (instead of playing)")
	listVoices := flag.Bool("voices", false, "list available voices")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: kokoro-say [options] [text]\n\n")
		fmt.Fprintf(os.Stderr, "Convert text to speech using Kokoro TTS.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  kokoro-say \"Hello, world!\"\n")
		fmt.Fprintf(os.Stderr, "  echo \"Hello\" | kokoro-say\n")
		fmt.Fprintf(os.Stderr, "  kokoro-say -v bf_emma \"British accent\"\n")
		fmt.Fprintf(os.Stderr, "  kokoro-say -o output.mp3 \"Save to file\"\n")
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

	serverURL := os.Getenv("KOKORO_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8880"
	}

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
		// ffplay supports stdin, afplay does not
		if _, err := exec.LookPath("ffplay"); err == nil {
			cmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", "-")
		} else {
			// Fallback: write to temp file for afplay
			return playWithTempFile(audio, "afplay")
		}
	case "linux":
		if _, err := exec.LookPath("ffplay"); err == nil {
			cmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", "-")
		} else if _, err := exec.LookPath("mpv"); err == nil {
			cmd = exec.Command("mpv", "--no-video", "-")
		} else if _, err := exec.LookPath("paplay"); err == nil {
			return playWithTempFile(audio, "paplay")
		} else {
			return playWithTempFile(audio, "aplay")
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	cmd.Stdin = audio
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func playWithTempFile(audio io.Reader, player string) error {
	f, err := os.CreateTemp("", "kokoro-*.mp3")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err := io.Copy(f, audio); err != nil {
		return err
	}
	f.Close()

	cmd := exec.Command(player, f.Name())
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
