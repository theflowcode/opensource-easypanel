package http

import (
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type listActionsInput struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Limit       int    `json:"limit"`
}

type actionIDInput struct {
	ID string `json:"id"`
}

type createNotificationInput struct {
	Name   string                    `json:"name"`
	Target domain.NotificationTarget `json:"target"`
	Events domain.NotificationEvents `json:"events"`
}

type updateNotificationInput struct {
	ID     string                    `json:"id"`
	Name   string                    `json:"name"`
	Target domain.NotificationTarget `json:"target"`
	Events domain.NotificationEvents `json:"events"`
}

type channelIDInput struct {
	ID string `json:"id"`
}

func toNotificationChannelDTO(ch *domain.NotificationChannel) *domain.NotificationChannel {
	if ch == nil {
		return nil
	}
	clone := *ch
	if clone.Target.Password != "" {
		clone.Target.Password = "********"
	}
	if clone.Target.AccessToken != "" {
		clone.Target.AccessToken = "********"
	}
	if clone.Target.Secret != "" {
		clone.Target.Secret = "********"
	}
	return &clone
}

// registerActionsAndExtraRoutes binds background tasks, audit logging, and notifications to the oRPC dispatcher.
func registerActionsAndExtraRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	d.Register("actions/listActions", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[listActionsInput](c)
		if err != nil {
			return nil, err
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 50
		}
		actions, err := db.ListActions(c.Context, in.ProjectName, in.ServiceName, limit, 0)
		if err != nil {
			return nil, err
		}
		if actions == nil {
			return []any{}, nil
		}

		if in.Status != "" {
			filtered := make([]*domain.Action, 0)
			for _, a := range actions {
				if a.Status == in.Status {
					filtered = append(filtered, a)
				}
			}
			return filtered, nil
		}

		return actions, nil
	})

	d.Register("actions/getAction", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[actionIDInput](c)
		if err != nil {
			return nil, err
		}
		return db.GetAction(c.Context, in.ID)
	})

	d.Register("actions/killAction", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	// Notification Channels
	d.Register("notifications/listNotificationChannels", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		channels, err := db.ListNotificationChannels(c.Context)
		if err != nil {
			return nil, err
		}
		if channels == nil {
			return []any{}, nil
		}
		dtos := make([]*domain.NotificationChannel, 0, len(channels))
		for _, ch := range channels {
			dtos = append(dtos, toNotificationChannelDTO(ch))
		}
		return dtos, nil
	})

	d.Register("notifications/createNotificationChannel", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[createNotificationInput](c)
		if err != nil {
			return nil, err
		}
		ch := &domain.NotificationChannel{
			ID:        domain.NewID(),
			Name:      strings.TrimSpace(in.Name),
			Target:    in.Target,
			Events:    in.Events,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := db.CreateNotificationChannel(c.Context, ch); err != nil {
			return nil, err
		}
		return toNotificationChannelDTO(ch), nil
	})

	d.Register("notifications/updateNotificationChannel", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[updateNotificationInput](c)
		if err != nil {
			return nil, err
		}
		if in.ID == "" {
			return nil, orpc.NewBadRequest("channel id is required")
		}
		ch, err := db.GetNotificationChannel(c.Context, in.ID)
		if err != nil {
			return nil, domain.ErrNotFound
		}
		if in.Name != "" {
			ch.Name = strings.TrimSpace(in.Name)
		}
		ch.Target = in.Target
		ch.Events = in.Events
		ch.UpdatedAt = time.Now().UTC()
		if err := db.UpdateNotificationChannel(c.Context, ch); err != nil {
			return nil, err
		}
		return toNotificationChannelDTO(ch), nil
	})

	d.Register("notifications/destroyNotificationChannel", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[channelIDInput](c)
		if err != nil {
			return nil, err
		}
		if err := db.DeleteNotificationChannel(c.Context, in.ID); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	testNotificationHandler := func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	}
	d.Register("notifications/testNotificationChannel", testNotificationHandler)
	d.Register("notifications/sendTestNotification", testNotificationHandler)
}
