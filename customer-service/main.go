package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "github.com/m-dehghani/customer-service/proto"
)

var jwtKey = []byte("your_secret_key")

type server struct {
	db *gorm.DB
}

type User struct {
	ID       uint32 `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		return nil, err
	}

	user := User{Username: req.Username, Password: string(hashedPassword)}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{Message: "registration successful", Customerid: user.ID}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: req.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{Token: tokenString, Message: "login successful"}, nil
}

var tokenBlacklist = make(map[string]bool)

func (s *server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	token := req.Token
	tokenBlacklist[token] = true
	return &pb.LogoutResponse{Message: "logout successful"}, nil
}

func (s *server) VerifyCustomerID(ctx context.Context, req *pb.VerifyCustomerIDRequest) (*pb.VerifyCustomerIDResponse, error) {
	var user User
	if err := s.db.Where("username = ? AND id = ?", req.Username, req.Customerid).First(&user).Error; err != nil {
		log.Println("in verifycustomer function customerId is: ", req.GetCustomerid(), "userName is: ", req.GetUsername())
		log.Println(err.Error())
		return &pb.VerifyCustomerIDResponse{Valid: false}, nil
	}
	return &pb.VerifyCustomerIDResponse{Valid: true}, nil
}

func main() {
	dsn := "host=" + os.Getenv("POSTGRES_HOST") + " user=" + os.Getenv("POSTGRES_USER") + " password=" + os.Getenv("POSTGRES_PASSWORD") + " dbname=" + os.Getenv("POSTGRES_DB") + " port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		return
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Println("error in grpc to connect to " + os.Getenv("CUSTOMER_SERVICE_PORT"))

		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterCustomerServiceServer(s, &server{db: db})

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
