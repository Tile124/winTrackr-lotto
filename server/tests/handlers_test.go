package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tile124/winTrackr/auth"
	"github.com/Tile124/winTrackr/db"
)

// loginAndGetCookie logs a seeded user in and returns their session cookie,
// for use by tests that need to call cookie-authenticated handlers directly.
func loginAndGetCookie(t *testing.T, email, password string) *http.Cookie {
	t.Helper()
	body := []byte(fmt.Sprintf(`{"Email": %q, "Password": %q}`, email, password))
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(auth.LoginHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("login returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set a session cookie")
	}
	return cookies[0]
}

// seedTestUser inserts a test user with a known password for test use.
func seedTestUser(email, password string) {
	_, err := db.InsertUser(email, password)
	if err != nil {
		log.Fatalf("failed to seed test user %s: %v", email, err)
	}
}

// TestMain ensures the DB is initialized and cleaned up.
func TestMain(m *testing.M) {
	// Remove any leftover test DB from a previous run.
	os.Remove("./users.db")
	db.InitDbs()
	code := m.Run()
	os.Remove("./users.db")
	os.Exit(code)
}

// Test validation of a user's credentials
func TestValidateUser(t *testing.T) {
	seedTestUser("testuser", "secret123")
	usr, err := auth.ValidateUser(&auth.UserCredentials{
		Email:    "testuser",
		Password: "secret123",
	})

	if err != nil {
		t.Errorf("error in validating user: %v", err)
	}
	if usr.Email != "testuser" {
		t.Errorf("got wrong user: got %q want %q", usr.Email, "testuser")
	}
}

// Test successful login
func TestLoginSuccess(t *testing.T) {
	seedTestUser("logintest", "mypassword")
	body := []byte(`{"Email": "logintest", "Password": "mypassword"}`)
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// Test malformed login request
func TestLoginBadRequest(t *testing.T) {
	body := []byte(`{: "admin": "Password"}`)
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

// Test incorrect password login
func TestLoginIncorrectPassword(t *testing.T) {
	seedTestUser("wrongpwduser", "correctpwd")
	body := []byte(`{"Email": "wrongpwduser", "Password": "wrongpwd"}`)
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

// Test login with non-existent user
func TestLoginNonExistentUser(t *testing.T) {
	body := []byte(`{"Email": "nobody", "Password": "anything"}`)
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

// Test fetching user scratchoffs after login
func TestGetUserScratchoffs(t *testing.T) {
	seedTestUser("scratchuser", "scratchpass")
	body := []byte(`{"Email": "scratchuser", "Password": "scratchpass"}`)
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("login handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	req, err = http.NewRequest("GET", "/api/get-scratchoffs", nil)
	if err != nil {
		t.Fatal(err)
	}
}

// addScratchoffForTest adds a scratchoff as the owner of the given cookie and
// returns its assigned Id by re-fetching the user's scratchoffs.
func addScratchoffForTest(t *testing.T, cookie *http.Cookie, name string) int64 {
	t.Helper()

	body := []byte(fmt.Sprintf(`{"Name": %q, "GameId": "G1", "Prize": "5", "Date": "2024-01-01"}`, name))
	req, err := http.NewRequest("POST", "/api/add-scratchoff", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	http.HandlerFunc(auth.AddScratchoffHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("add-scratchoff returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	req, err = http.NewRequest("GET", "/api/get-scratchoffs", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)

	rr = httptest.NewRecorder()
	http.HandlerFunc(auth.GetScratchoffsHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("get-scratchoffs returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var scratchoffs []db.Scratchoff
	if err := json.NewDecoder(rr.Body).Decode(&scratchoffs); err != nil {
		t.Fatal(err)
	}
	for _, s := range scratchoffs {
		if s.Name == name {
			return s.Id
		}
	}
	t.Fatalf("could not find newly added scratchoff %q", name)
	return 0
}

// Test that a user can update their own scratchoff entry
func TestUpdateScratchoffSuccess(t *testing.T) {
	seedTestUser("updateuser", "updatepass")
	cookie := loginAndGetCookie(t, "updateuser", "updatepass")
	id := addScratchoffForTest(t, cookie, "Original Name")

	body := []byte(fmt.Sprintf(`{"Id": %d, "Name": "Updated Name", "GameId": "G2", "Prize": "10", "Date": "2024-02-02"}`, id))
	req, err := http.NewRequest("POST", "/api/update-scratchoff", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	http.HandlerFunc(auth.UpdateScratchoffHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("update-scratchoff returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	req, err = http.NewRequest("GET", "/api/get-scratchoffs", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)

	rr = httptest.NewRecorder()
	http.HandlerFunc(auth.GetScratchoffsHandler).ServeHTTP(rr, req)

	var scratchoffs []db.Scratchoff
	if err := json.NewDecoder(rr.Body).Decode(&scratchoffs); err != nil {
		t.Fatal(err)
	}
	if len(scratchoffs) != 1 || scratchoffs[0].Name != "Updated Name" || scratchoffs[0].Prize != 10 {
		t.Errorf("scratchoff was not updated as expected: got %+v", scratchoffs)
	}
}

// Test that a user cannot update another user's scratchoff entry
func TestUpdateScratchoffWrongOwner(t *testing.T) {
	seedTestUser("updateowner", "ownerpass")
	ownerCookie := loginAndGetCookie(t, "updateowner", "ownerpass")
	id := addScratchoffForTest(t, ownerCookie, "Owner's Entry")

	seedTestUser("updateintruder", "intruderpass")
	intruderCookie := loginAndGetCookie(t, "updateintruder", "intruderpass")

	body := []byte(fmt.Sprintf(`{"Id": %d, "Name": "Hijacked", "GameId": "G3", "Prize": "99", "Date": "2024-03-03"}`, id))
	req, err := http.NewRequest("POST", "/api/update-scratchoff", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(intruderCookie)

	rr := httptest.NewRecorder()
	http.HandlerFunc(auth.UpdateScratchoffHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("update-scratchoff by non-owner returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// Test that a user can delete their own scratchoff entry
func TestDeleteScratchoffSuccess(t *testing.T) {
	seedTestUser("deleteuser", "deletepass")
	cookie := loginAndGetCookie(t, "deleteuser", "deletepass")
	id := addScratchoffForTest(t, cookie, "To Be Deleted")

	body := []byte(fmt.Sprintf(`{"Id": %d}`, id))
	req, err := http.NewRequest("POST", "/api/delete-scratchoff", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()
	http.HandlerFunc(auth.DeleteScratchoffHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("delete-scratchoff returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	req, err = http.NewRequest("GET", "/api/get-scratchoffs", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookie)

	rr = httptest.NewRecorder()
	http.HandlerFunc(auth.GetScratchoffsHandler).ServeHTTP(rr, req)

	var scratchoffs []db.Scratchoff
	if err := json.NewDecoder(rr.Body).Decode(&scratchoffs); err != nil {
		t.Fatal(err)
	}
	if len(scratchoffs) != 0 {
		t.Errorf("expected scratchoff to be deleted, still found: %+v", scratchoffs)
	}
}

// Test that a user cannot delete another user's scratchoff entry
func TestDeleteScratchoffWrongOwner(t *testing.T) {
	seedTestUser("deleteowner", "ownerpass")
	ownerCookie := loginAndGetCookie(t, "deleteowner", "ownerpass")
	id := addScratchoffForTest(t, ownerCookie, "Owner's Other Entry")

	seedTestUser("deleteintruder", "intruderpass")
	intruderCookie := loginAndGetCookie(t, "deleteintruder", "intruderpass")

	body := []byte(fmt.Sprintf(`{"Id": %d}`, id))
	req, err := http.NewRequest("POST", "/api/delete-scratchoff", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(intruderCookie)

	rr := httptest.NewRecorder()
	http.HandlerFunc(auth.DeleteScratchoffHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("delete-scratchoff by non-owner returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}
