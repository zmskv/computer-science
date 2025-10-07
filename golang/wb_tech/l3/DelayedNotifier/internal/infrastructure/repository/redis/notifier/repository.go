package notifier

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wb-go/wbf/redis"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/infrastructure/repository/redis/dto"
	"go.uber.org/zap"
)

type RedisRepository struct {
	client *redis.Client
	log    *zap.Logger
}

func NewRedisClient(addr, password string, db int, logger *zap.Logger) interfaces.NotifierRepository {
	client := redis.New(addr, password, db)
	return &RedisRepository{client: client, log: logger}
}

func (r *RedisRepository) CreateNote(ctx context.Context, note entity.Note) (string, error) {
	key := "notifier:task:" + note.ID

	noteDTO := dto.Note(note)
	data, err := json.Marshal(noteDTO)
	if err != nil {
		r.log.Error("failed to marshal note DTO", zap.Error(err))
		return "", err
	}

	if err := r.client.Set(ctx, key, data); err != nil {
		r.log.Error("failed to save note to redis", zap.Error(err))
		return "", err
	}

	return note.ID, nil
}

func (r *RedisRepository) GetNote(ctx context.Context, id string) (entity.Note, error) {
	key := "notifier:task:" + id
	val, err := r.client.Get(ctx, key)
	if err != nil {
		return entity.Note{}, err
	}

	var noteDTO dto.Note
	if err := json.Unmarshal([]byte(val), &noteDTO); err != nil {
		r.log.Error("failed to unmarshal note DTO", zap.Error(err))
		return entity.Note{}, err
	}

	return entity.Note(noteDTO), nil
}

func (r *RedisRepository) DeleteNote(ctx context.Context, id string) error {
	note, err := r.GetNote(ctx, id)
	if err != nil {
		return err
	}
	note.Status = "cancelled"

	return r.saveNote(ctx, note)
}

func (r *RedisRepository) UpdateNoteStatus(ctx context.Context, id, status string) error {
	note, err := r.GetNote(ctx, id)
	if err != nil {
		return err
	}
	note.Status = status

	return r.saveNote(ctx, note)
}

func (r *RedisRepository) UpdateNoteRetries(ctx context.Context, id string, retries int) error {
	note, err := r.GetNote(ctx, id)
	if err != nil {
		return err
	}
	note.Retries = retries

	return r.saveNote(ctx, note)
}

func (r *RedisRepository) saveNote(ctx context.Context, note entity.Note) error {
	key := "notifier:task:" + note.ID
	noteDTO := dto.Note(note)

	data, err := json.Marshal(noteDTO)
	if err != nil {
		r.log.Error("failed to marshal note DTO", zap.Error(err))
		return err
	}

	if err := r.client.Set(ctx, key, data); err != nil {
		r.log.Error("failed to save note to redis", zap.Error(err))
		return err
	}
	return nil
}

func (r *RedisRepository) GetDueNotificationIDs(ctx context.Context) ([]string, error) {
	var dueIDs []string

	keys, err := r.client.Keys(ctx, "notifier:task:*").Result()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	for _, key := range keys {
		val, err := r.client.Get(ctx, key)
		if err != nil {
			r.log.Warn("Failed to get note during scan", zap.String("key", key), zap.Error(err))
			continue
		}

		var note dto.Note
		if err := json.Unmarshal([]byte(val), &note); err != nil {
			r.log.Warn("Failed to unmarshal note during scan", zap.String("key", key), zap.Error(err))
			continue
		}
		if note.Status == "pending" && now.After(note.ExpirationTime.UTC()) {
			dueIDs = append(dueIDs, note.ID)
		}
	}

	return dueIDs, nil
}

func (r *RedisRepository) RemoveFromSchedule(ctx context.Context, ids ...string) error {
	for _, id := range ids {
		if err := r.client.Del(ctx, "notifier:task:"+id).Err(); err != nil {
			r.log.Error("Failed to delete processed task", zap.String("id", id), zap.Error(err))
		}
	}
	return nil
}
