package filesvc

import (
	"context"
	"io"

	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

func (s *FileService) Create(
	ctx context.Context,
	user company.User,
	reader io.Reader,
	opts FileOpts,
) (*ent.File, error) {
	rec := event.Get(ctx).Sub("file/create_with_access")
	var allowed bool
	if opts.Common {
		allowed = user.Can(NewCommonCreateFileAction())
	} else {
		allowed = user.Can(NewCreateFileAction(user.ID))
	}

	if !allowed {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access-control").Set(
			"allowed", false,
			"acting_user", user)
		return nil, sesc.ErrForbidden
	}
	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", user)
	return s.create(ctx, user.ID, reader, opts)
}

func (s *FileService) Delete(ctx context.Context, user company.User, id UUID) error {
	rec := event.Get(ctx).Sub("file/delete_with_access")
	f, err := s.byID(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		return err
	}

	if !user.Can(NewDeleteFileAction(f.OwnerID)) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", user)
		return sesc.ErrForbidden
	}
	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", user)
	return s.delete(ctx, id)
}

func (s *FileService) Search(
	ctx context.Context,
	user company.User,
	opts sesc.FileSearchOptions,
) (ent.Files, int, error) {
	rec := event.Get(ctx).Sub("file/search_with_access")
	files, count, err := s.search(ctx, opts)

	if err != nil {
		rec.Add(events.Error, err)
		return nil, -1, err
	}

	for _, f := range files {
		if !user.Can(NewViewFileAction(f.OwnerID)) {
			rec.Add(events.Error, sesc.ErrForbidden)
			rec.Sub("access_control").Set(
				"allowed", false,
				"acting_user", user)
			return nil, -1, sesc.ErrForbidden
		}
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", user)

	return files, count, nil
}

func (s *FileService) ByID(ctx context.Context, user company.User, id UUID) (*ent.File, error) {
	rec := event.Get(ctx).Sub("file/by_id_with_access")

	f, err := s.byID(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", user)
		return nil, err
	}

	if !user.Can(NewViewFileAction(f.OwnerID)) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", user)
		return nil, sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", user)
	return f, nil
}

func (s *FileService) DownloadURL(ctx context.Context, user company.User, id UUID) (string, error) {
	rec := event.Get(ctx).Sub("file/download_url_with_access")

	f, err := s.byID(ctx, id)
	if err != nil {
		rec.Add(events.Error, err)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", user)
		return "", err
	}

	if !user.Can(NewViewFileAction(f.OwnerID)) {
		rec.Add(events.Error, sesc.ErrForbidden)
		rec.Sub("access_control").Set(
			"allowed", false,
			"acting_user", user)
		return "", sesc.ErrForbidden
	}

	rec.Sub("access_control").Set(
		"allowed", true,
		"acting_user", user)
	return s.downloadURL(ctx, id)
}
