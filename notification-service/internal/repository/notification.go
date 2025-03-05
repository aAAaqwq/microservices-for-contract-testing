package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"notification-service/internal/model"
)

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, notification)
	if err != nil {
		return err
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*model.Notification, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var notification model.Notification
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uint) ([]model.Notification, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []model.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == "sent" {
		update["$set"].(bson.M)["sent_at"] = time.Now()
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *NotificationRepository) GetFailedNotifications(ctx context.Context) ([]model.Notification, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": "failed"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []model.Notification
	if err = cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}
