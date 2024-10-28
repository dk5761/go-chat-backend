// This is a placeholder for MongoDB migrations
db.createCollection("messages", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["sender_id", "receiver_id", "content", "created_at"],
            properties: {
                sender_id: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                receiver_id: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                content: {
                    bsonType: "string",
                    description: "must be a string and is required"
                },
                created_at: {
                    bsonType: "date",
                    description: "must be a date and is required"
                }
            }
        }
    }
});
