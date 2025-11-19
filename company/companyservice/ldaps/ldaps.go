package ldapds

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
)

type Config struct {
	Address  string // e.g., "localhost:389"
	BaseDN   string // e.g., "DC=lyceum,DC=usu,DC=ru"
	BindUser string // e.g., "Administrator@lyceum.usu.ru"
	BindPass string
}

type DataSource struct {
	cfg Config

	mu          sync.RWMutex
	users       map[string]company.User       // key: user ID
	departments map[string]company.Department // key: department ID

	authCacheMu sync.RWMutex
	authCache   map[string]authCacheEntry // key: hash of username+password+salt

	salt      string
	stopCh    chan struct{}
	stoppedCh chan struct{}
}

type authCacheEntry struct {
	validUntil time.Time
}

func New(cfg Config) (*DataSource, error) {
	ds := &DataSource{
		cfg:         cfg,
		users:       make(map[string]company.User),
		departments: make(map[string]company.Department),
		authCache:   make(map[string]authCacheEntry),
		salt:        generateSalt(),
		stopCh:      make(chan struct{}),
		stoppedCh:   make(chan struct{}),
	}

	// Initial sync
	if err := ds.ForceUpdate(context.Background()); err != nil {
		return nil, fmt.Errorf("initial LDAP sync failed: %w", err)
	}

	// Start background updater
	go ds.periodicUpdater()

	return ds, nil
}

func (ds *DataSource) Close() error {
	close(ds.stopCh)
	<-ds.stoppedCh
	return nil
}

// periodicUpdater updates the cache every 10 minutes with panic recovery
func (ds *DataSource) periodicUpdater() {
	defer close(ds.stoppedCh)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("LDAP updater panic recovered: %v\n", r)
			// Restart the updater
			go ds.periodicUpdater()
		}
	}()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ds.stopCh:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			if err := ds.ForceUpdate(ctx); err != nil {
				fmt.Printf("LDAP periodic update failed: %v\n", err)
			}
			cancel()
		}
	}
}

// ForceUpdate forces an immediate update of the in-memory cache from LDAP
func (ds *DataSource) ForceUpdate(ctx context.Context) error {
	conn, err := ds.connect()
	if err != nil {
		return fmt.Errorf("LDAP connection failed: %w", err)
	}
	defer conn.Close()

	// Fetch departments
	departments, err := ds.fetchDepartments(conn)
	if err != nil {
		return fmt.Errorf("failed to fetch departments: %w", err)
	}

	// Fetch users
	users, err := ds.fetchUsers(conn, departments)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	// Update cache atomically
	ds.mu.Lock()
	ds.departments = departments
	ds.users = users
	ds.mu.Unlock()

	return nil
}

func (ds *DataSource) connect() (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", ds.cfg.Address)
	if err != nil {
		return nil, err
	}

	err = conn.Bind(ds.cfg.BindUser, ds.cfg.BindPass)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func (ds *DataSource) fetchDepartments(conn *ldap.Conn) (map[string]company.Department, error) {
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("OU=employees,%s", ds.cfg.BaseDN),
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=organizationalUnit)",
		[]string{"ou"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	departments := make(map[string]company.Department)
	for _, entry := range sr.Entries {
		ou := entry.GetAttributeValue("ou")
		if ou == "" || ou == "employees" {
			continue
		}

		// Only include departments starting with kaf_ or otd_
		if !strings.HasPrefix(ou, "kaf_") && !strings.HasPrefix(ou, "otd_") {
			continue
		}

		departments[ou] = company.Department{
			ID:   ou,
			Name: ou, // For now, name equals ID
		}
	}

	return departments, nil
}

func (ds *DataSource) fetchUsers(
	conn *ldap.Conn,
	departments map[string]company.Department,
) (map[string]company.User, error) {
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("OU=employees,%s", ds.cfg.BaseDN),
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(&(objectClass=user)(sAMAccountName=*))",
		[]string{"sAMAccountName", "givenName", "sn", "displayName", "memberOf"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	users := make(map[string]company.User)
	for _, entry := range sr.Entries {
		username := entry.GetAttributeValue("sAMAccountName")
		if username == "" {
			continue
		}

		displayName := entry.GetAttributeValue("displayName")

		memberOf := entry.GetAttributeValues("memberOf")

		var departmentID string

		var roles []company.Role

		for _, dn := range memberOf {
			cn := extractCN(dn)

			if _, isDept := departments[cn]; isDept {
				departmentID = cn
			}

			if strings.HasPrefix(cn, "stim_") {
				role, err := company.AsRole(strings.TrimPrefix(cn, "stim_"))
				if err != nil {
					roles = append(roles, role)
				}

			}
		}

		users[username] = company.User{
			ID:           username,
			FullName:     displayName,
			DepartmentID: departmentID,
			Roles:        roles,
		}
	}

	return users, nil
}

// User implements company.DataSource
func (ds *DataSource) User(ctx context.Context, q companyquery.User) (company.User, error) {
	if q.ID == "" {
		return company.User{}, company.ErrUserNotFound
	}

	// If password is provided, authenticate
	if q.Password != "" {
		if err := ds.authenticateUser(q.ID, q.Password); err != nil {
			return company.User{}, company.ErrUserNotFound
		}
	}

	ds.mu.RLock()
	user, exists := ds.users[q.ID]
	ds.mu.RUnlock()

	if !exists {
		return company.User{}, company.ErrUserNotFound
	}

	return user, nil
}

// Users implements company.DataSource
func (ds *DataSource) Users(ctx context.Context, q companyquery.Users) ([]company.User, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []company.User

	for _, user := range ds.users {
		if !matchesUsersQuery(user, q) {
			continue
		}
		result = append(result, user)
	}

	return result, nil
}

// Department implements company.DataSource
func (ds *DataSource) Department(ctx context.Context, q companyquery.Department) (company.Department, error) {
	if q.ID == "" {
		return company.Department{}, company.ErrDepartmentNotFound
	}

	ds.mu.RLock()
	dept, exists := ds.departments[q.ID]
	ds.mu.RUnlock()

	if !exists {
		return company.Department{}, company.ErrDepartmentNotFound
	}

	return dept, nil
}

// Departments implements company.DataSource
func (ds *DataSource) Departments(ctx context.Context, q companyquery.Departments) ([]company.Department, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []company.Department

	for _, dept := range ds.departments {
		if q.Name != "" {
			if !strings.Contains(strings.ToLower(dept.Name), strings.ToLower(q.Name)) {
				continue
			}
		}
		result = append(result, dept)
	}

	return result, nil
}

// UsersWithIDs implements company.DataSource
func (ds *DataSource) UsersWithIDs(ctx context.Context, ids []string) ([]company.User, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make([]company.User, len(ids))
	for i, id := range ids {
		if user, exists := ds.users[id]; exists {
			result[i] = user
		} else {
			result[i] = company.ExEmployee(id)
		}
	}

	return result, nil
}

// authenticateUser attempts to bind to LDAP with user credentials
// Results are cached for 3 minutes
func (ds *DataSource) authenticateUser(username, password string) error {
	cacheKey := ds.hashCredentials(username, password)

	// Check cache first
	ds.authCacheMu.RLock()
	entry, exists := ds.authCache[cacheKey]
	ds.authCacheMu.RUnlock()

	if exists && time.Now().Before(entry.validUntil) {
		return nil // Cached authentication valid
	}

	// Attempt LDAP bind
	conn, err := ldap.Dial("tcp", ds.cfg.Address)
	if err != nil {
		return fmt.Errorf("LDAP connection failed: %w", err)
	}
	defer conn.Close()

	userDN := fmt.Sprintf("%s@%s", username, strings.TrimPrefix(ds.cfg.BaseDN, "DC="))
	userDN = strings.ReplaceAll(userDN, ",DC=", ".")

	err = conn.Bind(userDN, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Cache successful authentication
	ds.authCacheMu.Lock()
	ds.authCache[cacheKey] = authCacheEntry{
		validUntil: time.Now().Add(3 * time.Minute),
	}
	ds.authCacheMu.Unlock()

	// Clean old cache entries periodically
	go ds.cleanAuthCache()

	return nil
}

func (ds *DataSource) hashCredentials(username, password string) string {
	data := fmt.Sprintf("%s:%s:%s", username, password, ds.salt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (ds *DataSource) cleanAuthCache() {
	ds.authCacheMu.Lock()
	defer ds.authCacheMu.Unlock()

	now := time.Now()
	for key, entry := range ds.authCache {
		if now.After(entry.validUntil) {
			delete(ds.authCache, key)
		}
	}
}

// Helper functions

func matchesUsersQuery(user company.User, q companyquery.Users) bool {
	// Exact matches for ID fields
	if q.DepartmentID != "" && user.DepartmentID != q.DepartmentID {
		return false
	}
	if q.RoleID != "" && user.HasRole(company.Role(q.RoleID)) {
		return false
	}

	// Substring matches (case-insensitive)
	if q.Department != "" {
		if !strings.Contains(strings.ToLower(user.DepartmentID), strings.ToLower(q.Department)) {
			return false
		}
	}
	if q.FullName != "" {
		if !strings.Contains(strings.ToLower(user.FullName), strings.ToLower(q.FullName)) {
			return false
		}
	}

	return true
}

func extractCN(dn string) string {
	parts := strings.Split(dn, ",")
	if len(parts) == 0 {
		return ""
	}
	cnPart := parts[0]
	if strings.HasPrefix(cnPart, "CN=") {
		return strings.TrimPrefix(cnPart, "CN=")
	}
	return ""
}

func generateSalt() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
