//go:build integration

package inprocess

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	josejwt "github.com/go-jose/go-jose/v4/jwt"
)

// mockJWKSServer holds the mock JWKS server state.
type mockJWKSServer struct {
	privKey   *rsa.PrivateKey
	issuer    string
	caCertDER []byte
}

// newMockJWKSServer creates a mock Cloudflare Access JWKS server.
// It generates RSA-2048 keys (matching real Cloudflare's RS256) and serves
// the public key over HTTPS at /cdn-cgi/access/certs.
func newMockJWKSServer(t *testing.T, audience string) *mockJWKSServer {
	t.Helper()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	// Self-signed CA
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Mock CF Access CA"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &privKey.PublicKey, privKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	caCert, _ := x509.ParseCertificate(caCertDER)

	// Server cert for localhost
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:     []string{"localhost"},
	}
	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &privKey.PublicKey, privKey)
	if err != nil {
		t.Fatalf("create server cert: %v", err)
	}

	// JWKS endpoint
	jwk := jose.JSONWebKey{Key: &privKey.PublicKey, KeyID: "test-key-1", Algorithm: "RS256", Use: "sig"}
	mux := http.NewServeMux()
	mux.HandleFunc("/cdn-cgi/access/certs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []jose.JSONWebKey{jwk},
		})
	})

	// TLS server
	serverCert, _ := x509.ParseCertificate(serverCertDER)
	tlsCert := tls.Certificate{
		Certificate: [][]byte{serverCertDER, caCertDER},
		PrivateKey:  privKey,
		Leaf:        serverCert,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	tlsListener := tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{tlsCert}})

	srv := &http.Server{Handler: mux}
	go srv.Serve(tlsListener)
	t.Cleanup(func() { srv.Close() })

	return &mockJWKSServer{
		privKey:   privKey,
		issuer:    "https://" + listener.Addr().String(),
		caCertDER: caCertDER,
	}
}

// signJWT creates a signed RS256 JWT with the given claims.
func (m *mockJWKSServer) signJWT(t *testing.T, audience, email string, duration time.Duration) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: m.privKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key-1"),
	)
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}

	claims := josejwt.Claims{
		Issuer:   m.issuer,
		Audience: josejwt.Audience{audience},
		Subject:  "user-id-123",
		IssuedAt: josejwt.NewNumericDate(time.Now().Add(-5 * time.Minute)),
		Expiry:   josejwt.NewNumericDate(time.Now().Add(duration)),
	}
	custom := map[string]interface{}{"email": email}

	token, err := josejwt.Signed(signer).Claims(claims).Claims(custom).Serialize()
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return token
}
