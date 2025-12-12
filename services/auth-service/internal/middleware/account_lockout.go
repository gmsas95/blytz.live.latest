package middleware

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// AccountLockout tracks failed login attempts and locks accounts temporarily
type AccountLockout struct {
	mu              sync.Mutex
	attempts        map[string]*lockoutState
	maxAttempts     int           // Max failed attempts before lockout
	lockoutDuration time.Duration // How long to lock the account
	attemptWindow   time.Duration // Window for counting failed attempts
	logger          *zap.Logger
}

type lockoutState struct {
	failedAttempts int
	firstAttempt   time.Time
	lockedUntil    time.Time
}

// NewAccountLockout creates a new account lockout tracker
// maxAttempts: number of failed attempts before lockout
// lockoutDuration: how long to lock the account
// attemptWindow: time window for counting attempts (resets if no attempts in this window)
func NewAccountLockout(maxAttempts int, lockoutDuration, attemptWindow time.Duration, logger *zap.Logger) *AccountLockout {
	al := &AccountLockout{
		attempts:        make(map[string]*lockoutState),
		maxAttempts:     maxAttempts,
		lockoutDuration: lockoutDuration,
		attemptWindow:   attemptWindow,
		logger:          logger,
	}

	// Start cleanup goroutine
	go al.cleanup()

	return al
}

// cleanup removes old lockout entries periodically
func (al *AccountLockout) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		al.mu.Lock()
		now := time.Now()
		for email, state := range al.attempts {
			// Remove entries that are:
			// 1. Not locked and haven't had attempts in the window
			// 2. Locked but lockout has expired
			if state.lockedUntil.Before(now) && state.firstAttempt.Add(al.attemptWindow).Before(now) {
				delete(al.attempts, email)
			}
		}
		al.mu.Unlock()
	}
}

// IsLocked checks if an account is currently locked
func (al *AccountLockout) IsLocked(email string) bool {
	al.mu.Lock()
	defer al.mu.Unlock()

	state, exists := al.attempts[email]
	if !exists {
		return false
	}

	// Check if lockout has expired
	if time.Now().After(state.lockedUntil) && !state.lockedUntil.IsZero() {
		// Lockout expired, reset state
		state.failedAttempts = 0
		state.lockedUntil = time.Time{}
		return false
	}

	return !state.lockedUntil.IsZero() && time.Now().Before(state.lockedUntil)
}

// RemainingLockoutTime returns how long until the account is unlocked
func (al *AccountLockout) RemainingLockoutTime(email string) time.Duration {
	al.mu.Lock()
	defer al.mu.Unlock()

	state, exists := al.attempts[email]
	if !exists {
		return 0
	}

	remaining := time.Until(state.lockedUntil)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RecordFailedAttempt records a failed login attempt
// Returns true if the account is now locked
func (al *AccountLockout) RecordFailedAttempt(email string) bool {
	al.mu.Lock()
	defer al.mu.Unlock()

	now := time.Now()
	state, exists := al.attempts[email]

	if !exists {
		al.attempts[email] = &lockoutState{
			failedAttempts: 1,
			firstAttempt:   now,
		}
		return false
	}

	// Check if attempt window has passed, reset if so
	if now.After(state.firstAttempt.Add(al.attemptWindow)) {
		state.failedAttempts = 1
		state.firstAttempt = now
		state.lockedUntil = time.Time{}
		return false
	}

	// Increment failed attempts
	state.failedAttempts++

	// Check if we should lock the account
	if state.failedAttempts >= al.maxAttempts {
		state.lockedUntil = now.Add(al.lockoutDuration)
		if al.logger != nil {
			al.logger.Warn("Account locked due to failed login attempts",
				zap.String("email", email),
				zap.Int("failed_attempts", state.failedAttempts),
				zap.Duration("lockout_duration", al.lockoutDuration),
			)
		}
		return true
	}

	return false
}

// RecordSuccessfulLogin resets the failed attempt counter for an account
func (al *AccountLockout) RecordSuccessfulLogin(email string) {
	al.mu.Lock()
	defer al.mu.Unlock()

	delete(al.attempts, email)
}

// GetFailedAttempts returns the current number of failed attempts for an account
func (al *AccountLockout) GetFailedAttempts(email string) int {
	al.mu.Lock()
	defer al.mu.Unlock()

	state, exists := al.attempts[email]
	if !exists {
		return 0
	}
	return state.failedAttempts
}

// DefaultAccountLockout creates an AccountLockout with sensible defaults:
// - 5 failed attempts before lockout
// - 15 minute lockout duration
// - 30 minute attempt window
func DefaultAccountLockout(logger *zap.Logger) *AccountLockout {
	return NewAccountLockout(5, 15*time.Minute, 30*time.Minute, logger)
}
