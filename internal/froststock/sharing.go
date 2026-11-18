package froststock

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

type sharing struct {
	host *apphost.Host
}

func (s *sharing) register() error {
	if err := s.host.Migrate(1, `CREATE TABLE froststock_shares (
		email TEXT PRIMARY KEY,
		user_id TEXT NOT NULL DEFAULT ''
	);`); err != nil {
		return err
	}
	s.host.Mux.HandleFunc("GET /api/access", s.host.Require(s.status))
	s.host.Mux.HandleFunc("GET /api/sharing", s.owner(s.list))
	s.host.Mux.HandleFunc("POST /api/sharing", s.owner(s.add))
	s.host.Mux.HandleFunc("DELETE /api/sharing/{email}", s.owner(s.remove))
	return nil
}

func (s *sharing) isOwner(r *http.Request) bool {
	u := apphost.Current(r).User
	return u.Issuer == s.host.Config.Issuer && apphost.Equal(u.Subject, s.host.Config.Owner)
}

func (s *sharing) allowed(r *http.Request) (bool, error) {
	if s.isOwner(r) {
		return true, nil
	}
	u := apphost.Current(r).User
	if u.Issuer != s.host.Config.Issuer {
		return false, nil
	}
	// MyAccounts supplies verified email claims. Bind an invitation to the
	// first account that accepts it so a reassigned email cannot inherit access.
	_, err := s.host.DB.ExecContext(r.Context(), `UPDATE froststock_shares SET user_id=?
		WHERE email=? AND user_id=''`, u.ID, strings.ToLower(u.Email))
	if err != nil {
		return false, err
	}
	var exists bool
	err = s.host.DB.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM froststock_shares WHERE user_id=?)`, u.ID).Scan(&exists)
	return exists, err
}

func (s *sharing) require(handler http.HandlerFunc) http.HandlerFunc {
	return s.host.Require(func(w http.ResponseWriter, r *http.Request) {
		allowed, err := s.allowed(r)
		if err != nil {
			apphost.Problem(w, 503, "Freezer sharing is unavailable. Try again.")
			return
		}
		if !allowed {
			apphost.Problem(w, 403, "Ask the freezer owner to share Froststock with your OneAccounts email address.")
			return
		}
		handler(w, r)
	})
}

func (s *sharing) owner(handler http.HandlerFunc) http.HandlerFunc {
	return s.host.Owner(func(w http.ResponseWriter, r *http.Request) {
		if !s.isOwner(r) {
			apphost.Problem(w, 403, "Only the freezer owner can change these settings.")
			return
		}
		handler(w, r)
	})
}

func (s *sharing) status(w http.ResponseWriter, r *http.Request) {
	allowed, err := s.allowed(r)
	if err != nil {
		apphost.Problem(w, 503, "Freezer sharing is unavailable. Try again.")
		return
	}
	apphost.JSON(w, 200, map[string]bool{"owner": s.isOwner(r), "allowed": allowed})
}

func (s *sharing) list(w http.ResponseWriter, r *http.Request) {
	rows, err := s.host.DB.QueryContext(r.Context(), "SELECT email FROM froststock_shares ORDER BY email")
	if err != nil {
		apphost.Problem(w, 503, "The sharing list could not be opened.")
		return
	}
	defer rows.Close()
	emails := []string{}
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			apphost.Problem(w, 503, "The sharing list could not be opened.")
			return
		}
		emails = append(emails, email)
	}
	if rows.Err() != nil {
		apphost.Problem(w, 503, "The sharing list could not be opened.")
		return
	}
	apphost.JSON(w, 200, map[string]any{"emails": emails})
}

func (s *sharing) add(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	if !apphost.Decode(w, r, &input) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 {
		apphost.Problem(w, 400, "Enter the person's OneAccounts email address.")
		return
	}
	if _, err := s.host.DB.ExecContext(r.Context(),
		"INSERT INTO froststock_shares(email) VALUES(?) ON CONFLICT(email) DO NOTHING", email); err != nil {
		apphost.Problem(w, 503, "The freezer could not be shared. Try again.")
		return
	}
	s.list(w, r)
}

func (s *sharing) remove(w http.ResponseWriter, r *http.Request) {
	if _, err := s.host.DB.ExecContext(r.Context(), "DELETE FROM froststock_shares WHERE email=?", r.PathValue("email")); err != nil {
		apphost.Problem(w, 503, "Access could not be removed. Try again.")
		return
	}
	s.list(w, r)
}
