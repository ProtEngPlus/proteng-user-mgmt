package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/protengplus/proteng-user-mgmt/configs"
	"github.com/protengplus/proteng-user-mgmt/database"
	"github.com/protengplus/proteng-user-mgmt/internal/logger" // add this
	"github.com/protengplus/proteng-user-mgmt/utils"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	logger.InitZap()
	collectionFlag := flag.String("collection", "all", "which collection to backfill: users, admins, or all")
	dryRun := flag.Bool("dry-run", true, "if true, only print planned changes without writing")
	flag.Parse()

	configs.AutomaticLoadEnv()
	if err := database.ConnectToDB(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	var collections []string
	switch *collectionFlag {
	case "users":
		collections = []string{"users"}
	case "admins":
		collections = []string{"admins"}
	case "all":
		collections = []string{"users", "admins"}
	default:
		fmt.Fprintf(os.Stderr, "invalid -collection value: %s (must be users, admins, or all)\n", *collectionFlag)
		os.Exit(1)
	}

	for _, name := range collections {
		if err := checkDuplicates(name); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	for _, name := range collections {
		if err := backfillCollection(name, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "failed to backfill %s: %v\n", name, err)
			os.Exit(1)
		}
	}
}

func checkDuplicates(collectionName string) error {
	collection := database.GetCollection(collectionName)
	pipeline := bson.A{
		bson.M{"$group": bson.M{
			"_id":   bson.M{"$toLower": "$email"},
			"count": bson.M{"$sum": 1},
			"ids":   bson.M{"$push": "$_id"},
		}},
		bson.M{"$match": bson.M{"count": bson.M{"$gt": 1}}},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return fmt.Errorf("duplicate check failed for %s: %w", collectionName, err)
	}
	defer cursor.Close(context.Background())

	var duplicates []bson.M
	if err := cursor.All(context.Background(), &duplicates); err != nil {
		return fmt.Errorf("duplicate check decode failed for %s: %w", collectionName, err)
	}

	if len(duplicates) > 0 {
		fmt.Printf("found %d duplicate email group(s) in %s (case-insensitive):\n", len(duplicates), collectionName)
		for _, d := range duplicates {
			fmt.Printf("  email=%v ids=%v\n", d["_id"], d["ids"])
		}
		return fmt.Errorf("aborting: resolve duplicates in %s manually before backfilling (merge or delete)", collectionName)
	}

	return nil
}

func backfillCollection(collectionName string, dryRun bool) error {
	collection := database.GetCollection(collectionName)

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var count int
	for cursor.Next(context.Background()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		email, _ := doc["email"].(string)
		normalized := utils.NormalizeEmail(email)
		if email == normalized {
			continue
		}

		if dryRun {
			fmt.Printf("[dry-run] %s: %v -> %q\n", collectionName, doc["_id"], normalized)
		} else {
			_, err := collection.UpdateByID(context.Background(), doc["_id"], bson.M{"$set": bson.M{"email": normalized}})
			if err != nil {
				return fmt.Errorf("failed to update %v in %s: %w", doc["_id"], collectionName, err)
			}
			fmt.Printf("%s: %v -> %q\n", collectionName, doc["_id"], normalized)
		}
		count++
	}

	verb := "would be updated (dry-run)"
	if !dryRun {
		verb = "updated"
	}
	fmt.Printf("%s: %d document(s) %s\n", collectionName, count, verb)
	return nil
}
