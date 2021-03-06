package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	cast "github.com/barnybug/go-cast"
	"github.com/barnybug/go-cast/controllers"
	"github.com/evalphobia/google-tts-go/googletts"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"context"

	"github.com/grandcat/zeroconf"
)

func find(name string, do func(*cast.Client)) {
	// Find cache first
	if expire.After(time.Now()) && len(cache) != 0 {
		for name, entry := range cache {
			if strings.Contains(entry.ServiceInstanceName(), name) {
				log.Printf("Device %s at %s:%d", entry.ServiceInstanceName(), entry.AddrIPv4[0], entry.Port)
				do(cast.NewClient(entry.AddrIPv4[0], entry.Port))
			}
		}
		return
	}

	// Clear cache
	expire = time.Now()
	cache = make(map[string]*zeroconf.ServiceEntry)

	// Discover Google Home by mDNS for 10 seconds
	duration := 10 * time.Second
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			cache[entry.ServiceInstanceName()] = entry
			if strings.Contains(entry.ServiceInstanceName(), name) {
				log.Printf("Device %s at %s:%d", entry.ServiceInstanceName(), entry.AddrIPv4[0], entry.Port)
				do(cast.NewClient(entry.AddrIPv4[0], entry.Port))
			}
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	err = resolver.Browse(ctx, "_googlecast._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	expire = time.Now().AddDate(0, 0, 7)
}

var verificationToken string
var cache map[string]*zeroconf.ServiceEntry
var expire time.Time

func main() {
	cache = make(map[string]*zeroconf.ServiceEntry)

	verificationToken = os.Getenv("GHN_TOKEN")
	port := os.Getenv("GHN_PORT")
	if port == "" {
		port = "8080"
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/", handler)
	log.Printf("Started")
	http.ListenAndServe(":"+port, r)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if verificationToken != "" && verificationToken != r.FormValue("token") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	text := r.FormValue("text")
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	lang := r.FormValue("lang")
	device := r.FormValue("device")

	log.Printf("text = %s, lang = %s, device = %s", text, lang, device)
	url, err := googletts.GetTTSURL(text, lang)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	find(device, func(client *cast.Client) {
		client.Connect(ctx)

		media, err := client.Media(ctx)
		if err != nil {
			return
		}

		item := controllers.MediaItem{
			ContentId:   url,
			StreamType:  "BUFFERED",
			ContentType: "audio/mpeg",
		}
		_, err = media.LoadMedia(ctx, item, 0, true, map[string]interface{}{})
		if err != nil {
			return
		}
	})
	w.WriteHeader(http.StatusOK)
}
