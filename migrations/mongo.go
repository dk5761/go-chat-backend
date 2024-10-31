package migrations

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RunMigrations runs MongoDB migrations, such as collection creation and schema validation
func RunMigrations(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the collection already exists
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": "messages"})
	if err != nil {
		log.Printf("Failed to list collections: %v", err)
		return err
	}
	// If the "messages" collection already exists, skip the migration
	for _, collection := range collections {
		if collection == "messages" {
			log.Println("Collection 'messages' already exists, skipping migration.")
			return nil
		}
	}

	// Define schema for the messages collection
	schema := bson.M{
		"bsonType": "object",
		"required": []string{"sender_id", "receiver_id", "content", "created_at"},
		"properties": bson.M{
			"sender_id": bson.M{
				"bsonType":    "string",
				"description": "must be a string and is required",
			},
			"receiver_id": bson.M{
				"bsonType":    "string",
				"description": "must be a string and is required",
			},
			"content": bson.M{
				"bsonType":    "string",
				"description": "must be a string and is required",
			},
			"created_at": bson.M{
				"bsonType":    "date",
				"description": "must be a date and is required",
			},
		},
	}

	// Create collection with validation
	err = createCollectionWithValidation(ctx, db, "messages", schema)
	if err != nil {
		log.Printf("Failed to create messages collection: %v", err)
		return err
	}

	log.Println("Migration ran successfully: messages collection created with schema validation.")
	return nil
}

func createCollectionWithValidation(ctx context.Context, db *mongo.Database, collectionName string, schema bson.M) error {
	opts := options.CreateCollection().SetValidator(bson.M{"$jsonSchema": schema})

	err := db.CreateCollection(ctx, collectionName, opts)
	if err != nil {
		return err
	}

	log.Printf("Collection %s created successfully with schema validation", collectionName)
	return nil
}
