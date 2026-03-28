content = """package session

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"
    "time"

    "github.com/yourname/gonest/testutil"
)

func TestSessionMiddleware_Handle(t *testing.T) {
    sm := WithMemoryStore()
    cfg := &Config{SessionManager: sm}
    cfg.ContextKey = "session"
    mw := New(cfg)

    handlerCalled := false
    handler := func(c gonest.Context) error {
        handlerCalled = true
        Put(c, "test_key", "test_value")
        Put(c, "user_id", "user123")
        return nil
    }

    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    w := httptest.NewRecorder()
    ctx := gonest.NewContext(w, req)

    err := mw.Handle(ctx, func() error { return handler(ctx) })

    if err != nil {
        t.Fatalf("Handler should not return error: %v", err)
    }
    if !handlerCalled {
        t.Fatal("Handler should have been called")
    }

    if v := GetSession(ctx); v != nil {
        if v.GetString(ctx.Context(), "test_key") != "test_value" {
            t.Errorf("Expected test_key=test_value, got %v", v.GetString(ctx.Context(), "test_key"))
        }
        if v.GetString(ctx.Context(), "user_id") != "user123" {
            t.Errorf("Expected user_id=user123, got %v", v.GetString(ctx.Context(), "user_id"))
        }
    }
}

func TestSessionMiddleware_SkipPaths(t *testing.T) {
    sm := WithMemoryStore()
    cfg := &Config{SessionManager: sm}
    cfg.ContextKey = "session"
    cfg.SkipPaths = []string{"/health", "/metrics"}
    mw := New(cfg)

    handlerCount := 0
    handler := func(c gonest.Context) error { handlerCount++; return nil }

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    w := httptest.NewRecorder()
    mw.Handle(gonest.NewContext(w, req), handler)
    if handlerCount != 1 {
        t.Errorf("Handler should be called for skipped path")
    }
    if GetSession(gonest.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/health", nil))) != nil {
        t.Error("Session should not be set for skipped path")
    }

    req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
    w = httptest.NewRecorder()
    mw.Handle(gonest.NewContext(w, req), handler)
    if handlerCount != 2 {
        t.Errorf("Handler should be called for non-skipped path")
    }
    if GetSession(gonest.NewContext(w, req)) == nil {
        t.Error("Session should be set for non-skipped path")
    }
}

func TestInMemoryStore(t *testing.T) {
    store := NewInMemoryStore()

    err := store.Commit("token", []byte("data"), time.Now().Add(time.Hour))
    if err != nil {
        t.Fatalf("Commit should not error: %v", err)
    }

    data, found, err := store.Find("token")
    if err != nil {
        t.Fatalf("Find should not error: %v", err)
    }
    if !found || string(data) != "data" {
        t.Error("Data should be found")
    }

    _, found, _ = store.Find("nonexistent")
    if found {
        t.Error("Non-existent token should not be found")
    }

    err = store.Delete("token")
    if err != nil {
        t.Fatalf("Delete should not error: %v", err)
    }

    _, found, _ = store.Find("token")
    if found {
        t.Error("Deleted token should not be found")
    }
}

func TestInMemoryStore_Concurrent(t *testing.T) {
    store := NewInMemoryStore()
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            store.Commit("token_"+string(rune('0'+id)), []byte("data"), time.Now().Add(time.Hour))
        }(i)
    }
    wg.Wait()
}

func TestInMemoryStore_Expiry(t *testing.T) {
    store := NewInMemoryStore()
    store.Commit("expired", []byte("data"), time.Now().Add(-time.Hour))
    _, found, _ := store.Find("expired")
    if found {
        t.Error("Expired data should not be found")
    }

    store.Commit("valid", []byte("data"), time.Now().Add(time.Hour))
    _, found, _ = store.Find("valid")
    if !found {
        t.Error("Valid data should be found")
    }
}

func TestSession_GetSet(t *testing.T) {
    sm := WithMemoryStore()
    cfg := &Config{SessionManager: sm}
    cfg.ContextKey = "session"
    mw := New(cfg)

    ctx, _ := testutil.NewTestContext(http.MethodGet, "/test", nil)
    mw.Handle(ctx, func() error {
        Put(ctx, "key", "value")
        if Get(ctx, "key") != "value" {
            t.Error("Get should return value")
        }
        Remove(ctx, "key")
        if Get(ctx, "key") != nil {
            t.Error("Remove should remove value")
        }
        Put(ctx, "k1", "v1")
        Clear(ctx)
        if Get(ctx, "k1") != nil {
            t.Error("Clear should clear all")
        }
        return nil
    })
}

func TestSession_Types(t *testing.T) {
    sm := WithMemoryStore()
    cfg := &Config{SessionManager: sm}
    cfg.ContextKey = "session"
    mw := New(cfg)

    ctx, _ := testutil.NewTestContext(http.MethodGet, "/test", nil)
    mw.Handle(ctx, func() error {
        Put(ctx, "str", "hello")
        Put(ctx, "int", 123)
        Put(ctx, "bool", true)
        Put(ctx, "bytes", []byte("world"))

        if GetString(ctx, "str") != "hello" {
            t.Error("GetString failed")
        }
        if GetInt(ctx, "int") != 123 {
            t.Error("GetInt failed")
        }
        if GetBool(ctx, "bool") != true {
            t.Error("GetBool failed")
        }
        if string(GetBytes(ctx, "bytes")) != "world" {
            t.Error("GetBytes failed")
        }
        return nil
    })
}

func TestSession_UserHelpers(t *testing.T) {
    sm := WithMemoryStore()
    cfg := &Config{SessionManager: sm}
    cfg.ContextKey = "session"
    mw := New(cfg)

    ctx, _ := testutil.NewTestContext(http.MethodGet, "/test", nil)
    mw.Handle(ctx, func() error {
        SetUserID(ctx, "user123")
        if GetUserID(ctx) != "user123" {
            t.Error("GetUserID failed")
        }

        type User struct{ ID, Name string }
        SetUser(ctx, User{"user123", "Test User"})
        var u User
        if !GetUser(ctx, &u) {
            t.Error("GetUser failed")
        }

        if !IsAuthenticated(ctx) {
            t.Error("Should be authenticated")
        }

        ClearUser(ctx)
        if IsAuthenticated(ctx) {
            t.Error("Should not be authenticated after ClearUser")
        }
        return nil
    })
}

func TestSession_Destroy(t *testing.T) {
    sm := WithMemoryStore()
    cfg := &Config{SessionManager: sm}
    cfg.ContextKey = "session"
    mw := New(cfg)

    ctx, _ := testutil.NewTestContext(http.MethodGet, "/test", nil)
    mw.Handle(ctx, func() error {
        Put(ctx, "key", "value")
        Destroy(ctx)
        if Get(ctx, "key") != nil {
            t.Error("Destroy should clear data")
        }
        return nil
    })
}

func TestSession_RenewToken(t *testing.T) {
    sm := WithMemoryStore()
    cfg := &Config{SessionManager: sm}
    cfg.ContextKey = "session"
    mw := New(cfg)

    ctx, _ := testutil.NewTestContext(http.MethodGet, "/test", nil)
    mw.Handle(ctx, func() error {
        oldToken := sm.Token(ctx.Context())
        if err := RenewToken(ctx); err != nil {
            t.Fatalf("RenewToken failed: %v", err)
        }
        newToken := sm.Token(ctx.Context())
        if oldToken == newToken {
            t.Error("Token should change after renewal")
        }
        return nil
    })
}

func TestNewSessionManager(t *testing.T) {
    sm := NewSessionManager(NewInMemoryStore())
    if sm == nil {
        t.Fatal("NewSessionManager should return session manager")
    }
    if sm.Cookie.Name != "session_id" {
        t.Errorf("Wrong cookie name: %s", sm.Cookie.Name)
    }
    if !sm.Cookie.HttpOnly {
        t.Error("Cookie should be HttpOnly")
    }
    if !sm.Cookie.Secure {
        t.Error("Cookie should be Secure")
    }
    if sm.Lifetime != 24*time.Hour {
        t.Errorf("Wrong lifetime: %v", sm.Lifetime)
    }
    if sm.IdleTimeout != 2*time.Hour {
        t.Errorf("Wrong idle timeout: %v", sm.IdleTimeout)
    }
}

func TestWithMemoryStore(t *testing.T) {
    sm := WithMemoryStore()
    if sm == nil {
        t
