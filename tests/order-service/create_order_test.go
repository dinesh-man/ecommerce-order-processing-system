package order_service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dinesh-man/ecommerce-order-processing-system/order-service/service"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var (
	rc *redis.Client
	sk string
)

func TestCreateOrder(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	//launch miniredis for testing purposes
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("could not start miniredis: %v", err)
	}
	defer mr.Close()

	// Connect go-redis client to miniredis
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Mock inventory service
	mockInventory := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		product := models.Product{
			ID:    "P001",
			Name:  "Mock Product Name",
			Stock: 10,
			Price: 150,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	}))
	defer mockInventory.Close()

	mt.Run("order service test", func(mt *mtest.T) {
		// Simulate successful insert response from MongoDB
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Create order input
		order := models.Order{
			Items: []models.LineItem{
				{ProductID: "P001", Quantity: 2},
			},
		}

		// Patch mongodb.GetCollection to return mtest collection
		originalGetCollection := service.GetCollection
		service.GetCollection = func(name string) *mongo.Collection {
			return mt.Coll
		}
		defer func() { service.GetCollection = originalGetCollection }()

		orderService := service.NewOrderService("orders", mockInventory.URL, rc, "orders")
		createdOrder, err := orderService.CreateOrder(order)

		assert.NoError(t, err)
		assert.Equal(t, models.Pending, createdOrder.Status)
		assert.NotEqual(t, primitive.NilObjectID, createdOrder.ID)

		// Simulate FindOne response for GetOrderByID
		orderDoc := bson.D{
			{"_id", createdOrder.ID},
			{"customer_id", createdOrder.CustomerID},
			{"items", bson.A{
				bson.D{
					{"product_id", "P001"},
					{"quantity", 2},
				},
			}},
			{"status", string(createdOrder.Status)},
			{"created_at", createdOrder.CreatedAt},
			{"updated_at", createdOrder.UpdatedAt},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "orders.orders", mtest.FirstBatch, orderDoc))

		// Fetch order by ID
		fetchedOrder, err := orderService.GetOrderByID(createdOrder.ID.Hex())
		assert.NoError(t, err)
		assert.Equal(t, createdOrder.ID, fetchedOrder.ID)
		assert.Equal(t, createdOrder.CustomerID, fetchedOrder.CustomerID)
		assert.Equal(t, createdOrder.Status, fetchedOrder.Status)
		assert.Equal(t, createdOrder.Items, fetchedOrder.Items)
		assert.WithinDuration(t, createdOrder.CreatedAt, fetchedOrder.CreatedAt, time.Second)
		assert.WithinDuration(t, createdOrder.UpdatedAt, fetchedOrder.UpdatedAt, time.Second)
	})
}
