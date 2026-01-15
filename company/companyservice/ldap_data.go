package companyservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

var _ DataSource = (*ldapDS)(nil)

type LDAPConfig struct {
	URL          string
	BindDN       string
	BindPassword string
	BaseDN       string
	SyncInterval time.Duration
}

type ldapDS struct {
	config    LDAPConfig
	mu        sync.RWMutex
	storage   *storage
	eventSink EventSink
	ctx       context.Context
	cancel    context.CancelFunc
}

type EventSink interface {
	RecordEvent(ctx context.Context, rec *event.Record)
}

func NewLDAP(ctx context.Context, config LDAPConfig, eventSink EventSink) (DataSource, error) {
	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(ctx)

	ds := &ldapDS{
		config:    config,
		storage:   newStorage(nil, nil),
		eventSink: eventSink,
		ctx:       ctx,
		cancel:    cancel,
	}

	if err := ds.sync(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("initial LDAP sync failed: %w", err)
	}

	go ds.backgroundSync()

	return ds, nil
}

func (l *ldapDS) Close() {
	l.cancel()
}

func (l *ldapDS) backgroundSync() {
	ticker := time.NewTicker(l.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						ctx, rec := event.NewRecord(l.ctx, "ldap_sync_panic")
						rec.Set(
							"panic", r,
							"panic_message", fmt.Sprintf("%v", r),
							events.Critical, true,
						)
						l.eventSink.RecordEvent(ctx, rec)
					}
				}()

				if err := l.sync(l.ctx); err != nil {
					ctx, rec := event.NewRecord(l.ctx, "ldap_sync_error")
					rec.Set(
						events.Error, err,
						"error_message", err.Error(),
					)
					l.eventSink.RecordEvent(ctx, rec)
				}
			}()
		}
	}
}

func (l *ldapDS) sync(ctx context.Context) error {
	ctx, rec := event.NewRecord(ctx, "ldap_sync")
	defer l.eventSink.RecordEvent(ctx, rec)

	startTime := time.Now()
	rec.Set("start_time", startTime)

	conn, err := ldap.DialURL(l.config.URL)
	if err != nil {
		rec.Add(events.Error, err)
		return fmt.Errorf("failed to dial LDAP: %w", err)
	}
	defer conn.Close()

	if err := conn.Bind(l.config.BindDN, l.config.BindPassword); err != nil {
		rec.Add(events.Error, err)
		return fmt.Errorf("failed to bind to LDAP: %w", err)
	}

	rec.Set("ldap_connected", true)

	departments, err := l.fetchDepartments(conn)
	if err != nil {
		rec.Add(events.Error, err)
		return fmt.Errorf("failed to fetch departments: %w", err)
	}

	rec.Set("departments_fetched", len(departments))

	users, err := l.fetchUsers(conn)
	if err != nil {
		rec.Add(events.Error, err)
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	rec.Set("users_fetched", len(users))

	newStorage := newStorage(users, departments)
	l.mu.Lock()
	l.storage = newStorage
	l.mu.Unlock()

	duration := time.Since(startTime)
	rec.Set(
		"sync_duration", duration,
		"sync_success", true,
	)

	return nil
}

func (l *ldapDS) fetchDepartments(conn *ldap.Conn) ([]company.Department, error) {
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("ou=Catalogues,%s", l.config.BaseDN),
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=organizationalUnit)",
		[]string{"ou", "description"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("department search failed: %w", err)
	}

	var departments []company.Department
	for _, entry := range sr.Entries {
		ou := entry.GetAttributeValue("ou")
		if ou == "Catalogues" {
			continue
		}
		if !strings.HasPrefix(ou, "kaf_") && !strings.HasPrefix(ou, "otd_") {
			continue
		}

		departments = append(departments, company.Department{
			ID:          ou,
			Name:        entry.GetAttributeValue("description"),
			Description: entry.GetAttributeValue("description"),
		})
	}

	return departments, nil
}

func (l *ldapDS) fetchUsers(conn *ldap.Conn) ([]company.User, error) {
	searchRequest := ldap.NewSearchRequest(
		l.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=inetOrgPerson)",
		[]string{"cn", "uid", "displayName", "memberOf", "ou"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	var users []company.User
	for _, entry := range sr.Entries {
		uid := entry.GetAttributeValue("uid")
		if uid == "" {
			continue
		}

		displayName := entry.GetAttributeValue("displayName")
		if displayName == "" {
			displayName = entry.GetAttributeValue("cn")
		}

		dn := entry.DN
		departmentID := l.extractDepartmentFromDN(dn)
		roles := l.extractRolesFromMemberOf(entry.GetAttributeValues("memberOf"))

		users = append(users, company.User{
			ID:           uid,
			FullName:     displayName,
			DepartmentID: departmentID,
			Roles:        roles,
		})
	}

	return users, nil
}

func (l *ldapDS) extractDepartmentFromDN(dn string) string {
	parts := strings.Split(dn, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "ou=") {
			ou := strings.TrimPrefix(part, "ou=")
			if strings.HasPrefix(ou, "kaf_") || strings.HasPrefix(ou, "otd_") {
				return ou
			}
		}
	}
	return ""
}

func (l *ldapDS) extractRolesFromMemberOf(memberOf []string) []company.Role {
	var roles []company.Role
	for _, group := range memberOf {
		parts := strings.Split(group, ",")
		if len(parts) == 0 {
			continue
		}
		cnPart := strings.TrimSpace(parts[0])
		if !strings.HasPrefix(cnPart, "cn=") {
			continue
		}
		cn := strings.TrimPrefix(cnPart, "cn=")

		if !strings.HasPrefix(cn, "stim_") {
			continue
		}
		roleStr := strings.TrimPrefix(cn, "stim_")

		role, err := company.AsRole(roleStr)
		if err == nil {
			roles = append(roles, role)
		}
	}
	return roles
}

func (l *ldapDS) User(ctx context.Context, q companyquery.User) (company.User, error) {
	l.mu.RLock()
	s := l.storage
	l.mu.RUnlock()

	verifyPassword := func(userID, password string) error {
		return l.verifyPassword(ctx, userID, password)
	}

	return s.queryUser(ctx, q, verifyPassword)
}

func (l *ldapDS) verifyPassword(_ context.Context, userID, password string) error {
	conn, err := ldap.DialURL(l.config.URL)
	if err != nil {
		return fmt.Errorf("failed to dial LDAP: %w", err)
	}
	defer conn.Close()

	if err := conn.Bind(l.config.BindDN, l.config.BindPassword); err != nil {
		return fmt.Errorf("failed to bind for search: %w", err)
	}

	searchRequest := ldap.NewSearchRequest(
		l.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1, 0, false,
		fmt.Sprintf("(uid=%s)", ldap.EscapeFilter(userID)),
		[]string{"dn"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("user search failed: %w", err)
	}

	if len(sr.Entries) == 0 {
		return errors.New("user not found")
	}

	userDN := sr.Entries[0].DN

	if err := conn.Bind(userDN, password); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}

func (l *ldapDS) Users(ctx context.Context, q companyquery.Users) ([]company.User, error) {
	l.mu.RLock()
	s := l.storage
	l.mu.RUnlock()

	return s.queryUsers(ctx, q)
}

func (l *ldapDS) UsersWithIDs(ctx context.Context, ids []string) ([]company.User, error) {
	l.mu.RLock()
	s := l.storage
	l.mu.RUnlock()

	return s.usersWithIDs(ctx, ids)
}

func (l *ldapDS) Department(ctx context.Context, q companyquery.Department) (company.Department, error) {
	l.mu.RLock()
	s := l.storage
	l.mu.RUnlock()

	return s.queryDepartment(ctx, q)
}

func (l *ldapDS) Departments(ctx context.Context, q companyquery.Departments) ([]company.Department, error) {
	l.mu.RLock()
	s := l.storage
	l.mu.RUnlock()

	return s.queryDepartments(ctx, q)
}
