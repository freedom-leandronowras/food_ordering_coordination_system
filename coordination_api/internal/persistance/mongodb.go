package persistence

import (
	"context"
	"fmt"
	"strings"
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
	usersCollection   = "users"
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

	_, err = r.db.Collection(usersCollection).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "member_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "email_domain", Value: 1}, {Key: "email", Value: 1}},
		},
	})
	if err != nil {
		return fmt.Errorf("create users indexes: %w", err)
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

func (r *MongoRepository) OrdersByMember(memberID uuid.UUID) ([]domain.FoodOrder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.db.Collection(ordersCollection).Find(
		ctx,
		bson.M{"member_id": memberID.String()},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("find orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []domain.FoodOrder
	for cursor.Next(ctx) {
		var doc orderDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode order: %w", err)
		}
		id, _ := uuid.Parse(doc.ID)
		memID, _ := uuid.Parse(doc.MemberID)

		items := make([]domain.FoodItem, 0, len(doc.Items))
		for _, di := range doc.Items {
			itemID, _ := uuid.Parse(di.ID)
			items = append(items, domain.FoodItem{
				ID:       itemID,
				Name:     di.Name,
				Quantity: di.Quantity,
				Price:    di.Price,
			})
		}
		orders = append(orders, domain.FoodOrder{
			ID:            id,
			MemberID:      memID,
			Items:         items,
			Status:        domain.OrderStatus(doc.Status),
			TotalPrice:    doc.TotalPrice,
			DeliveryNotes: doc.DeliveryNotes,
		})
	}
	return orders, cursor.Err()
}

type orderDocument struct {
	ID            string              `bson:"id"`
	MemberID      string              `bson:"member_id"`
	Status        string              `bson:"status"`
	TotalPrice    float64             `bson:"total_price"`
	DeliveryNotes string              `bson:"delivery_notes"`
	Items         []orderItemDocument `bson:"items"`
}

type orderItemDocument struct {
	ID       string  `bson:"id"`
	Name     string  `bson:"name"`
	Quantity int     `bson:"quantity"`
	Price    float64 `bson:"price"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func emailDomain(email string) string {
	at := strings.LastIndex(email, "@")
	if at == -1 || at == len(email)-1 {
		return ""
	}
	return email[at+1:]
}

func (r *MongoRepository) CreateUser(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	email := normalizeEmail(user.Email)
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	_, err := r.db.Collection(usersCollection).InsertOne(ctx, bson.M{
		"id":            user.UserID.String(),
		"member_id":     user.MemberID.String(),
		"email":         email,
		"email_domain":  emailDomain(email),
		"full_name":     strings.TrimSpace(user.FullName),
		"role":          strings.ToUpper(strings.TrimSpace(user.Role)),
		"password_hash": user.PasswordHash,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	})
	return err
}

func (r *MongoRepository) FindUserByEmail(email string) (User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc userDocument
	err := r.db.Collection(usersCollection).
		FindOne(ctx, bson.M{"email": normalizeEmail(email)}).
		Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	user, err := doc.toUser()
	if err != nil {
		return User{}, false, err
	}
	return user, true, nil
}

func (r *MongoRepository) FindUserByMemberID(memberID uuid.UUID) (User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc userDocument
	err := r.db.Collection(usersCollection).
		FindOne(ctx, bson.M{"member_id": memberID.String()}).
		Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	user, err := doc.toUser()
	if err != nil {
		return User{}, false, err
	}
	return user, true, nil
}

func (r *MongoRepository) ListUsersByEmailDomain(domain string) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.db.Collection(usersCollection).Find(
		ctx,
		bson.M{"email_domain": strings.ToLower(strings.TrimSpace(domain))},
		options.Find().SetSort(bson.D{{Key: "email", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	users := []User{}
	for cursor.Next(ctx) {
		var doc userDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		user, err := doc.toUser()
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

type userDocument struct {
	ID           string    `bson:"id"`
	MemberID     string    `bson:"member_id"`
	Email        string    `bson:"email"`
	FullName     string    `bson:"full_name"`
	Role         string    `bson:"role"`
	PasswordHash string    `bson:"password_hash"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func (d userDocument) toUser() (User, error) {
	userID, err := uuid.Parse(d.ID)
	if err != nil {
		return User{}, fmt.Errorf("parse user id: %w", err)
	}
	memberID, err := uuid.Parse(d.MemberID)
	if err != nil {
		return User{}, fmt.Errorf("parse member id: %w", err)
	}
	return User{
		UserID:       userID,
		MemberID:     memberID,
		Email:        d.Email,
		FullName:     d.FullName,
		Role:         d.Role,
		PasswordHash: d.PasswordHash,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}, nil
}

// EventsByAggregate returns every event for the given aggregate, sorted by
// occurred_at ascending.  This is the fundamental read path for event
// sourcing: replay this slice to rebuild the aggregate's current state.
func (r *MongoRepository) EventsByAggregate(ctx context.Context, aggregateID uuid.UUID) ([]domain.Event, error) {
	cursor, err := r.db.Collection(eventsCollection).Find(
		ctx,
		bson.M{"aggregate_id": aggregateID.String()},
		options.Find().SetSort(bson.D{{Key: "occurred_at", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []domain.Event
	for cursor.Next(ctx) {
		var doc eventDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode event: %w", err)
		}

		id, _ := uuid.Parse(doc.ID)
		aggID, _ := uuid.Parse(doc.AggregateID)

		var correlationID, causationID uuid.UUID
		if doc.CorrelationID != "" {
			correlationID, _ = uuid.Parse(doc.CorrelationID)
		}
		if doc.CausationID != "" {
			causationID, _ = uuid.Parse(doc.CausationID)
		}

		events = append(events, domain.Event{
			ID:            id,
			Type:          domain.EventType(doc.Type),
			AggregateID:   aggID,
			CorrelationID: correlationID,
			CausationID:   causationID,
			OccurredAt:    doc.OccurredAt,
			Payload:       map[string]any(doc.Payload),
		})
	}
	return events, cursor.Err()
}

type eventDocument struct {
	ID            string    `bson:"id"`
	Type          string    `bson:"type"`
	AggregateID   string    `bson:"aggregate_id"`
	CorrelationID string    `bson:"correlation_id"`
	CausationID   string    `bson:"causation_id"`
	OccurredAt    time.Time `bson:"occurred_at"`
	Payload       bson.M    `bson:"payload"`
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
