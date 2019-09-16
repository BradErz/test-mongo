package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Id        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Nickname  string    `json:"nickname" bson:"nickname"`
	Email     string    `json:"email" bson:"email"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

func NewUser(i int) User {
	return User{
		Id:        uuid.New().String(),
		Name:      fmt.Sprintf("test-name-%d", i),
		Nickname:  fmt.Sprintf("test-nickname-%d", i),
		Email:     fmt.Sprintf("test-email-%d@gmail.com", i),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

type Database struct {
	users *mongo.Collection
	db    *mongo.Database
}

// NewDatabase connects to the database with a given URI and attempts to ping when connecting
func NewDatabase(uri string) (*Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("error creating db: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}
	return &Database{
		users: client.Database("test-mongo").Collection("users"),
		db:    client.Database("test-mongo"),
	}, nil
}

func (d Database) GetUser(ctx context.Context, id string) (User, error) {
	query := bson.M{"_id": id}
	var user User
	if err := d.users.FindOne(ctx, query).Decode(&user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (d Database) AddUser(ctx context.Context, user User) error {
	if _, err := d.users.InsertOne(ctx, user); err != nil {
		return err
	}
	return nil
}

func (d Database) Drop(ctx context.Context) error {
	return d.db.Drop(ctx)
}
