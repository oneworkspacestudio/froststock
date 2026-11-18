package froststock

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

func TestSharedFreezerAccessAndRevocation(t *testing.T) {
	data := filepath.Join(t.TempDir(), "private")
	root := filepath.Join(t.TempDir(), "Froststock")
	h, err := apphost.New(apphost.Config{ID: "froststock", Name: "Froststock", Data: data,
		Listen: "127.0.0.1:29048", Origin: "http://localhost:29048",
		Issuer: "https://accounts.example.test", Owner: "freezer-owner", ClientID: "fixture-client", ClientSecret: "fixture-secret"}, UI())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.Close() })
	if err := Register(h, root); err != nil {
		t.Fatal(err)
	}
	type person struct{ token, csrf, id string }
	login := func(subject, email, issuer string) person {
		t.Helper()
		p := person{apphost.ID() + apphost.ID(), apphost.ID(), apphost.Identity(issuer, subject)}
		_, err := h.DB.Exec(`INSERT INTO host_users(id,issuer,subject,name,email,role) VALUES(?,?,?,?,?,'member')`,
			p.id, issuer, subject, "Synthetic person", email)
		if err != nil {
			t.Fatal(err)
		}
		_, err = h.DB.Exec(`INSERT INTO host_sessions(id,user_id,csrf,expires) VALUES(?,?,?,?)`,
			apphost.Hash(p.token), p.id, p.csrf, time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano))
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	owner := login("freezer-owner", "owner@example.test", h.Config.Issuer)
	member := login("member", "friend@example.test", h.Config.Issuer)
	other := login("other", "friend@example.test", h.Config.Issuer)
	foreign := login("foreign", "friend@example.test", "https://foreign.example.test")
	unverified := login("unverified", "", h.Config.Issuer)
	call := func(p person, method, path, body, contentType string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, h.Config.Origin+path, strings.NewReader(body))
		r.Header.Set("Origin", h.Config.Origin)
		r.Header.Set("X-CSRF-Token", p.csrf)
		r.Header.Set("X-Workspace-Expected-Account", p.id)
		r.Header.Set("Content-Type", contentType)
		if p.token != "" {
			r.AddCookie(&http.Cookie{Name: "froststock_session", Value: p.token})
		}
		w := httptest.NewRecorder()
		h.Handler().ServeHTTP(w, r)
		return w
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d, got %d: %s", status, w.Code, w.Body.String())
		}
	}
	check(call(person{}, "GET", "/api/items", "", ""), 401)
	check(call(member, "GET", "/api/items", "", ""), 403)
	check(call(foreign, "GET", "/api/items", "", ""), 401)
	check(call(owner, "GET", "/api/items", "", ""), 200)
	check(call(member, "POST", "/api/sharing", `{"email":"friend@example.test"}`, "application/json"), 403)
	noCSRF := owner
	noCSRF.csrf = ""
	check(call(noCSRF, "POST", "/api/sharing", `{"email":"friend@example.test"}`, "application/json"), 403)
	check(call(owner, "POST", "/api/sharing", `{"email":"not-an-email"}`, "application/json"), 400)
	check(call(owner, "POST", "/api/sharing", `{"email":" Friend@Example.test "}`, "application/json"), 200)
	check(call(unverified, "GET", "/api/items", "", ""), 403)
	check(call(member, "GET", "/api/access", "", ""), 200)
	check(call(other, "GET", "/api/items", "", ""), 403)
	check(call(member, "GET", "/api/sharing", "", ""), 403)
	check(call(member, "PUT", "/api/recognition/settings", `{"key":"test-key"}`, "application/json"), 403)
	check(call(member, "GET", "/api/recognition/settings", "", ""), 200)
	check(call(member, "POST", "/api/recognition/read", "", ""), 204)

	item := entry{ID: "shared-food-001", Job: "shared-job-0001", Created: 1, Quantity: 1, State: "ready", Name: "Peas"}
	var photo, body bytes.Buffer
	if err := jpeg.Encode(&photo, image.NewGray(image.Rect(0, 0, 40, 40)), nil); err != nil {
		t.Fatal(err)
	}
	form := multipart.NewWriter(&body)
	record, _ := json.Marshal(item)
	_ = form.WriteField("record", string(record))
	_ = form.WriteField("item", photo.String())
	_ = form.WriteField("expiry", photo.String())
	_ = form.Close()
	check(call(member, "POST", "/api/items/"+item.ID, body.String(), form.FormDataContentType()), 201)
	ownerList := call(owner, "GET", "/api/items", "", "")
	memberList := call(member, "GET", "/api/items", "", "")
	if !bytes.Equal(ownerList.Body.Bytes(), memberList.Body.Bytes()) || !bytes.Contains(ownerList.Body.Bytes(), []byte("Peas")) {
		t.Fatal("owner and shared member did not see the same freezer")
	}
	check(call(member, "GET", "/api/items/"+item.ID+"/item.jpg", "", ""), 200)
	item.Revision, item.Name = 1, "Garden peas"
	record, _ = json.Marshal(item)
	check(call(member, "PUT", "/api/items/"+item.ID, string(record), "application/json"), 200)
	check(call(owner, "PUT", "/api/items/"+item.ID, string(record), "application/json"), 409)
	check(call(owner, "POST", "/api/sharing", `{"email":"friend@example.test"}`, "application/json"), 200)
	check(call(other, "GET", "/api/items", "", ""), 403)
	if _, err := h.DB.Exec("UPDATE host_users SET email=? WHERE id=?", "renamed@example.test", member.id); err != nil {
		t.Fatal(err)
	}
	check(call(member, "GET", "/api/items", "", ""), 200)
	check(call(owner, "DELETE", "/api/sharing/friend@example.test", "", ""), 200)
	for _, endpoint := range []string{"/api/items", "/api/items/" + item.ID + "/item.jpg", "/api/recognition/settings"} {
		check(call(member, "GET", endpoint, "", ""), 403)
	}
	check(call(member, "PUT", "/api/items/"+item.ID, string(record), "application/json"), 403)
	check(call(member, "POST", "/api/recognition/read", "", ""), 403)
	check(call(owner, "GET", "/api/items", "", ""), 200)
	if _, err := os.Stat(filepath.Join(root, item.ID, "item.jpg")); err != nil {
		t.Fatal("revocation removed a shared photo", err)
	}
}
