package qualification_test

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/everFinance/goether"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGraphQLErrorBodyIsLiteral(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		limit  bool
	}{
		{"percent error", 400, "query 100% invalid: %s", false},
		{"ordinary error", 400, "query invalid", false},
		{"rate limit", 429, "slow down", true},
		{"successful query", 200, `{"data":{"synthetic":true}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/graphql" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()
			data, err := goar.NewClient(server.URL).GraphQL(context.Background(), "{ synthetic }")
			if tc.status == 200 {
				if err != nil || string(data) != `{"synthetic":true}` {
					t.Fatalf("successful query = %s, %v", data, err)
				}
				return
			}
			if tc.limit {
				if !errors.Is(err, goar.ErrRequestLimit) {
					t.Fatalf("rate limit = %v", err)
				}
				return
			}
			if err == nil || err.Error() != tc.body {
				t.Fatalf("error = %v, want literal %q", err, tc.body)
			}
		})
	}
}

func TestSyntheticEthereumBundleRoundTrip(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	signer, err := goether.NewSigner(hex.EncodeToString(crypto.FromECDSA(key)))
	if err != nil {
		t.Fatal(err)
	}
	itemSigner, err := goar.NewItemSigner(signer)
	if err != nil {
		t.Fatal(err)
	}
	item, err := itemSigner.CreateAndSignItem([]byte("synthetic decoder qualification"), "", "", []types.Tag{{Name: "Synthetic", Value: "true"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := utils.VerifyBundleItem(item); err != nil {
		t.Fatalf("fresh signature rejected: %v", err)
	}
	addr, err := utils.ItemSignerAddr(item)
	if err != nil || addr != signer.Address.String() {
		t.Fatalf("signer = %q, %v; want %q", addr, err, signer.Address.String())
	}
	decoded, err := utils.DecodeBundleItem(item.ItemBinary)
	if err != nil {
		t.Fatal(err)
	}
	if err := utils.VerifyBundleItem(*decoded); err != nil {
		t.Fatalf("decoded signature rejected: %v", err)
	}
	item.Data = utils.Base64Encode([]byte("tampered"))
	if err := utils.VerifyBundleItem(item); err == nil {
		t.Fatal("modified data accepted under original signature")
	}
}
