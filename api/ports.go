package api

import (
	"context"

	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	SESC interface {
		CreateDepartment(ctx context.Context, name, description string) (sesc.Department, error)
		CreateTeacher(ctx context.Context, opt sesc.UserOptions, department sesc.Department) (sesc.User, error)
		CreateUser(ctx context.Context, opt sesc.UserOptions, role sesc.Role) (sesc.User, error)
		User(ctx context.Context, id sesc.UUID) (sesc.User, error)
		SetRole(ctx context.Context, user sesc.User, role sesc.Role) (sesc.User, error)
		SetDepartment(ctx context.Context, user sesc.User, department sesc.Department) (sesc.User, error)
		SetUserInfo(ctx context.Context, user sesc.User, opt sesc.UserOptions) (sesc.User, error)
		SetProfilePic(ctx context.Context, user sesc.User, pictureURL string) (sesc.User, error)
		Roles(ctx context.Context) ([]sesc.Role, error)
		Departments(ctx context.Context) ([]sesc.Department, error)
		DepartmentByID(ctx context.Context, id sesc.UUID) (sesc.Department, error)
		GrantExtraPermissions(ctx context.Context, user sesc.User, permissions ...sesc.Permission) (sesc.User, error)
		RevokeExtraPermissions(ctx context.Context, user sesc.User, permissions ...sesc.Permission) (sesc.User, error)
	}
)
