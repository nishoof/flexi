package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/repository"
)

var testJWTTokenCookie *http.Cookie
var testUserId int64
var testTermId int64

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("Error loading .env file\n", err)
		os.Exit(1)
	}

	setupTestUser()
	setupTestTerm()

	token, err := generateJWT(testUserId)
	if err != nil {
		fmt.Printf("Failed to generate JWT: %v\n", err)
		os.Exit(1)
	}
	testJWTTokenCookie = &http.Cookie{
		Name:  "auth_token",
		Value: token,
	}

	exitCode := m.Run()

	if err := cleanupTestUser(); err != nil {
		fmt.Printf("Failed to clean up test user: %v\n", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}

func setupTestUser() {
	userId, err := getOrCreateUser(context.Background(), "testuser@nishilanand.com")
	if err != nil {
		fmt.Printf("Failed to create test user: %v\n", err)
		os.Exit(1)
	}
	if userId == noUserId {
		fmt.Println("Failed to create test user: no user ID returned")
		os.Exit(1)
	}
	testUserId = userId
	fmt.Println("Test user created with ID:", testUserId)
}

func setupTestTerm() {
	queries, err := database.Queries(context.Background())
	if err != nil {
		fmt.Printf("Failed to get database queries: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	term, err := queries.GetActiveTerm(ctx, testUserId)
	if err == nil {
		testTermId = term.ID
		fmt.Println("Test term already exists with ID:", testTermId)
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		fmt.Printf("Failed to get test term: %v\n", err)
		os.Exit(1)
	}

	endDate, err := time.Parse("2006-01-02", "2026-05-23")
	if err != nil {
		fmt.Printf("Failed to parse test term end date: %v\n", err)
		os.Exit(1)
	}

	termID, err := queries.CreateActiveTerm(ctx, repository.CreateActiveTermParams{
		UserID: testUserId,
		Name:   "Spring 2026",
		EndDate: pgtype.Date{
			Time:  endDate,
			Valid: true,
		},
	})
	if err != nil {
		fmt.Printf("Failed to create test term: %v\n", err)
		os.Exit(1)
	}
	testTermId = termID
	fmt.Println("Test term created with ID:", testTermId)
}

func cleanupTestUser() error {
	queries, err := database.Queries(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get database queries: %w", err)
	}

	ctx := context.Background()
	if err := queries.DeleteEntriesByUser(ctx, testUserId); err != nil {
		return fmt.Errorf("failed to delete test entries: %w", err)
	}
	if err := queries.DeleteActiveTermByUser(ctx, testUserId); err != nil {
		return fmt.Errorf("failed to delete test term: %w", err)
	}
	if err := queries.DeleteUser(ctx, testUserId); err != nil {
		return fmt.Errorf("failed to delete test user: %w", err)
	}
	return nil
}

func sendRequest(method string, url string, body io.Reader, auth *http.Cookie, handler http.HandlerFunc) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, body)
	if auth != nil {
		req.AddCookie(auth)
	}

	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}

func assertStatusAndBody(t testing.TB, expected, actual int, body *bytes.Buffer) {
	if actual != expected {
		t.Fatalf("expected status %d, got %d, body: %s", expected, actual, body.String())
	}
}
