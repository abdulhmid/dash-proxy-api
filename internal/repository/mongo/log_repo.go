package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"api-source-proxy/internal/model"
)

type LogRepo struct {
	col *mongo.Collection
}

func NewLogRepo(db *mongo.Database) *LogRepo {
	col := db.Collection("activity_logs")

	_, _ = col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "api_source_name", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})

	return &LogRepo{col: col}
}

func (r *LogRepo) Insert(ctx context.Context, log *model.ActivityLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("insert activity log: %w", err)
	}
	return nil
}

func (r *LogRepo) List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]model.ActivityLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	f := bson.M{}
	for k, v := range filter {
		switch k {
		case "start_date":
			f["created_at"] = bson.M{"$gte": v}
		case "end_date":
			if existing, ok := f["created_at"]; ok {
				existingMap := existing.(bson.M)
				existingMap["$lte"] = v
			} else {
				f["created_at"] = bson.M{"$lte": v}
			}
		default:
			f[k] = v
		}
	}

	total, err := r.col.CountDocuments(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("count logs: %w", err)
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := r.col.Find(ctx, f, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []model.ActivityLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, fmt.Errorf("decode logs: %w", err)
	}

	if logs == nil {
		logs = []model.ActivityLog{}
	}

	return logs, total, nil
}
