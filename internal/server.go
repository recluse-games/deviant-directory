package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/go-redis/redis/v7"

	pb "github.com/recluse-games/deviant-protobuf/genproto/go/directory"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// DB is a pointer to our redis client struct
var DB *redis.Client

// NewDBClient Generate a new database client.
func NewDBClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error attempting to ping Redis cache")
	}

	return client
}

type server struct {
}

// Start Starts a new directory server
func Start() {
	// Create a new instance of our Redis client.
	DB = NewDBClient()

	listen, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Could not listen on port: %v", err)
	}

	// gRPC Server
	s := grpc.NewServer()
	pb.RegisterDirectoryServer(s, &server{})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	log.Printf("Hosting server on: %s", listen.Addr().String())
}

// GetPlayer reads an player using the ID in the database
// Returns the player and error (if any)
func (s *server) GetPlayer(ctx context.Context, em *pb.ID) (*pb.Player, error) {
	// If ID is null, return specific error
	if em.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is empty, please try again")
	}

	result, err := getPlayer(em)
	if err != nil {
		return result, err
	}

	return result, nil
}

// UpdatePlayer updates an player using a player object
// Returns the updated player and error (if any)
func (s *server) UpdatePlayer(ctx context.Context, em *pb.Player) (*pb.Player, error) {
	// If ID is null, return specific error
	if em.Id == nil {
		return nil, status.Error(codes.InvalidArgument, "ID is empty, please try again")
	}

	result, err := getPlayer(em.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Player with ID does not exist, please try again")
	}

	result, err = setPlayer(em)

	return result, err
}

// CreatePlayer creates a player from a player object
// Returns the new player and error (if any)
func (s *server) CreatePlayer(ctx context.Context, em *pb.Player) (*pb.Player, error) {
	// If Name is empty, return specific error
	if em.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Name is empty, please try again")
	}

	_, err := getPlayer(em.Id)
	if err != nil {
		result, err := setPlayer(em)

		return result, err
	}

	return nil, status.Error(codes.InvalidArgument, "Name is already taken, please try again")
}

// DeletePlayer deletes a player from a ID
// Returns the new player and error (if any)
func (s *server) DeletePlayer(ctx context.Context, em *pb.ID) (*pb.ID, error) {
	// If Name is empty, return specific error
	if em == nil {
		return nil, status.Error(codes.InvalidArgument, "ID is empty, please try again")
	}

	_, err := getPlayer(em)
	if err != nil {
		return nil, err
	}

	err = DB.Del(em.Id).Err()
	if err != nil {
		return nil, status.Error(codes.NotFound, "Unable to delete player from ID")
	}

	return em, nil
}

func getPlayer(id *pb.ID) (*pb.Player, error) {
	result := &pb.Player{}

	options := protojson.UnmarshalOptions{
		AllowPartial: true,
	}

	in, err := DB.Get(id.Id).Result()
	if err != nil {
		log.Printf("Error retrieving player with id: %s, error: %v", id.Id, err)
		return nil, err
	}

	err = protojson.UnmarshalOptions(options).Unmarshal([]byte(in), result)
	if err != nil {
		log.Printf("Error unmarshalling player json, error: %v", err)
		return nil, err
	}

	return result, nil
}

func setPlayer(player *pb.Player) (*pb.Player, error) {
	options := protojson.MarshalOptions{
		AllowPartial:    true,
		EmitUnpopulated: true,
	}

	result, err := protojson.MarshalOptions(options).Marshal(player)
	if err != nil {
		log.Printf("Error marshalling player to json: %v, error: %v", player, err)
		return nil, err
	}

	err = DB.Set(player.Id.Id, string(result), 0).Err()
	if err != nil {
		log.Printf("Error writing player update to database with id: %s, error: %v", player.Id.Id, err)
		return nil, err
	}

	return player, nil
}
