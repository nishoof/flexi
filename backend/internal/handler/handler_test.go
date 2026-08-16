package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/nishoof/flexi/backend/internal/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testJWTTokenCookie *http.Cookie
var testUserId int64

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	container, err := startTestDatabase(ctx)
	if err != nil {
		fmt.Printf("Failed to start test database: %v\n", err)
		fmt.Println("Docker must be running to execute api tests.")
		return 1
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Printf("Failed to terminate postgres testcontainer: %v\n", err)
		}
	}()

	// Test-only signing key (base64 of "testke"); same value as util/jwt_test.go.
	if err := os.Setenv("JWT_KEY", "dGVzdGtl"); err != nil {
		fmt.Printf("Failed to set JWT_KEY: %v\n", err)
		return 1
	}

	if err := setupTestUser(); err != nil {
		fmt.Printf("Failed to set up test user: %v\n", err)
		return 1
	}

	token, err := generateJWT(testUserId)
	if err != nil {
		fmt.Printf("Failed to generate JWT: %v\n", err)
		return 1
	}
	testJWTTokenCookie = &http.Cookie{
		Name:  "auth_token",
		Value: token,
	}

	exitCode := m.Run()

	if err := cleanupTestUser(); err != nil {
		fmt.Printf("Failed to clean up test user: %v\n", err)
		return 1
	}

	return exitCode
}

// startTestDatabase starts Postgres via testcontainers, applies supabase
// migrations, and sets DATABASE_URL for the test process.
func startTestDatabase(ctx context.Context) (*postgres.PostgresContainer, error) {
	scripts, err := getMigrationScripts()
	if err != nil {
		return nil, fmt.Errorf("failed to locate migrations: %w", err)
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithOrderedInitScripts(scripts...),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(container)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}
	if err := os.Setenv("DATABASE_URL", connStr); err != nil {
		_ = testcontainers.TerminateContainer(container)
		return nil, fmt.Errorf("failed to set DATABASE_URL: %w", err)
	}

	return container, nil
}

// getMigrationScripts returns supabase migration .sql paths in filename order
// (timestamp prefixes already match apply order).
func getMigrationScripts() ([]string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get caller info")
	}
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "supabase", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir: %w", err)
	}

	var scripts []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		scripts = append(scripts, filepath.Join(dir, e.Name()))
	}
	if len(scripts) == 0 {
		return nil, fmt.Errorf("failed to find .sql migrations in %s", dir)
	}
	sort.Strings(scripts)
	return scripts, nil
}

func setupTestUser() error {
	userId, err := getOrCreateUser(context.Background(), "testuser@nishilanand.com")
	if err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}
	if userId == noUserId {
		return fmt.Errorf("failed to create test user: no user ID returned")
	}
	testUserId = userId

	_, err = getOrCreateTerm(context.Background(), testUserId)
	if err != nil {
		return fmt.Errorf("failed to create test term: %w", err)
	}

	fmt.Println("Test user and term created")
	return nil
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
	if err := queries.DeleteTermsByUser(ctx, testUserId); err != nil {
		return fmt.Errorf("failed to delete test terms: %w", err)
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
