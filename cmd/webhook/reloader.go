package main

import (
	"crypto/tls"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type certReloader struct {
	certFile string
	keyFile  string
	mu       sync.RWMutex
	cert     *tls.Certificate
	lastMod  time.Time
}

func newCertReloader(certFile, keyFile string, initialCert tls.Certificate) *certReloader {
	r := &certReloader{
		certFile: certFile,
		keyFile:  keyFile,
		cert:     &initialCert,
		lastMod:  time.Now(),
	}
	return r
}

func (r *certReloader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.checkAndReload()
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cert, nil
}

func (r *certReloader) checkAndReload() {
	fi, err := os.Stat(r.certFile)
	if err != nil {
		return
	}
	if !fi.ModTime().After(r.lastMod) {
		return
	}

	newCert, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)
	if err != nil {
		log.Printf("[reloader] warning: failed to reload cert: %v", err)
		return
	}

	r.mu.Lock()
	r.cert = &newCert
	r.lastMod = fi.ModTime()
	r.mu.Unlock()
	log.Printf("[reloader] reloaded TLS certificate from %s", r.certFile)
}

func (r *certReloader) watch() {
	// Background periodic stat checker for symlinks / mounted secrets
	dir := filepath.Dir(r.certFile)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if fi, err := os.Stat(dir); err == nil {
				if fi.ModTime().After(r.lastMod) {
					r.checkAndReload()
				}
			}
			r.checkAndReload()
		}
	}()
}
