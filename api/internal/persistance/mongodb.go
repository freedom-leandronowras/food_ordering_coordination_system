package persistence

import (
	"context"
	"fmt"
	"time"

	"food_ordering_coordination_system/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	creditsCollection = "credits"
	ordersCollection  = "orders"
	eventsCollection  = "events"
)

type MongoRepository struct {
	db *mongo.Database
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

func (r *MongoRepository) EnsureSchema(ctx context.Context) error {
	_, err := r.db.Collection(creditsCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "member_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create credits index: %w", err)
	}

	_, err = r.db.Collection(ordersCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create orders index: %w", err)
	}

	_, err = r.db.Collection(eventsCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create events index: %w", err)
	}

	return nil
}

func (r *MongoRepository) Get(memberID uuid.UUID) (float64, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var row struct {
		Amount float64 `bson:"amount"`
	}
	err := r.db.Collection(creditsCollection).
		FindOne(ctx, bson.M{"member_id": memberID.String()}).
		Decode(&row)
	if err == mongo.ErrNoDocuments {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return row.Amount, true, nil
}

func (r *MongoRepository) Set(memberID uuid.UUID, amount float64) error {
	if amount > domain.MaxMemberCredits {
		return fmt.Errorf("credits exceed maximum allowed (%v)", domain.MaxMemberCredits)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.db.Collection(creditsCollection).UpdateOne(
		ctx,
		bson.M{"member_id": memberID.String()},
		bson.M{"$set": bson.M{"member_id": memberID.String(), "amount": amount}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *MongoRepository) Save(order domain.FoodOrder) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	items := make([]bson.M, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, bson.M{
			"id":       item.ID.String(),
			"name":     item.Name,
			"quantity": item.Quantity,
			"price":    item.Price,
		})
	}

	_, err := r.db.Collection(ordersCollection).InsertOne(ctx, bson.M{
		"id":             order.ID.String(),
		"member_id":      order.MemberID.String(),
		"status":         string(order.Status),
		"total_price":    order.TotalPrice,
		"delivery_notes": order.DeliveryNotes,
		"items":          items,
		"created_at":     time.Now().UTC(),
	})
	return err
}

func (r *MongoRepository) Append(event domain.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload, err := bson.Marshal(event.Payload)
	if err != nil {
		return err
	}

	var payloadValue bson.M
	if err := bson.Unmarshal(payload, &payloadValue); err != nil {
		return err
	}

	doc := bson.M{
		"id":           event.ID.String(),
		"type":         string(event.Type),
		"aggregate_id": event.AggregateID.String(),
		"occurred_at":  event.OccurredAt,
		"payload":      payloadValue,
	}
	if event.CorrelationID != uuid.Nil {
		doc["correlation_id"] = event.CorrelationID.String()
	}
	if event.CausationID != uuid.Nil {
		doc["causation_id"] = event.CausationID.String()
	}

	_, err = r.db.Collection(eventsCollection).InsertOne(ctx, doc)
	return err
}
