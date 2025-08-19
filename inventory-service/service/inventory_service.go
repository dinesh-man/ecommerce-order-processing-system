package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/models"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/mongodb"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

type InventoryService struct {
	collectionName string
}

func NewInventoryService(collectionName string) *InventoryService {
	return &InventoryService{collectionName: collectionName}
}

// GetAllProducts returns all available products
func (s *InventoryService) GetAllProducts() ([]models.Product, error) {
	collection := mongodb.GetCollection(s.collectionName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Fatal("Failed to close cursor")
		}
	}(cursor, ctx)

	var products []models.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// GetProductByID fetches available product by its id
func (s *InventoryService) GetProductByID(id string) (*models.Product, error) {

	if !strings.HasPrefix(id, "P") {
		return nil, errors.New("invalid product ID")
	}

	collection := mongodb.GetCollection(s.collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product models.Product
	err := collection.FindOne(ctx, bson.M{"id": id}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}
