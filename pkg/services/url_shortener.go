// url_shortener.go

package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"url_shortener/api/pb"

	"github.com/go-redis/redis/v8"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type URLShortenerServer struct {
	redisClient *redis.Client
	counter     uint64
	pb.UnimplementedURLShortenerServer
	mu sync.RWMutex
}

func (s *URLShortenerServer) mustEmbedUnimplementedURLShortenerServer() {

}

func NewURLShortenerServer(redisClient *redis.Client) *URLShortenerServer {
	return &URLShortenerServer{
		redisClient: redisClient,
	}
}

func (s *URLShortenerServer) ShortenURL(ctx context.Context, req *pb.URLRequest) (*pb.URLResponse, error) {
	shortKey := generateShortKey(s)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Store the short URL in Redis
	if err := s.redisClient.Set(ctx, shortKey, req.OriginalUrl, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to store short URL in Redis: %v", err)
	}

	shortenedURL := fmt.Sprintf("http://yourdomain/%s", shortKey)

	return &pb.URLResponse{
		ShortUrl:    shortenedURL,
		OriginalUrl: req.OriginalUrl,
	}, nil
}

func (s *URLShortenerServer) ExpandURL(ctx context.Context, req *pb.URLRequest) (*pb.URLResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Retrieve the short URL from Redis
	shortKey := getShortKey(req.OriginalUrl)
	originalURL, err := s.redisClient.Get(ctx, shortKey).Result()
	if err != nil {
		return nil, fmt.Errorf("URL not found in Redis: %v", err)
	}

	return &pb.URLResponse{
		ShortUrl:    fmt.Sprintf("http://yourdomain/%s", shortKey),
		OriginalUrl: originalURL,
	}, nil
}

func generateShortKey(s *URLShortenerServer) string {
	// Increment counter for each URL
	counter := atomic.AddUint64(&s.counter, 1)
	return Base62Encode(counter)
}

func getShortKey(originalURL string) string {
	// Implement your logic to retrieve short key from original URL
	// This function will perform a reverse operation of Base62Encode
	var number uint64
	length := uint64(len(alphabet))
	for _, char := range originalURL {
		index := strings.IndexRune(alphabet, char)
		if index == -1 {
			// Invalid character in the original URL
			return "invalid_short_key"
		}
		number = number*length + uint64(index)
	}

	return Base62Encode(number)
}

func Base62Encode(number uint64) string {
	length := uint64(len(alphabet))
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)

	for ; number > 0; number = number / length {
		encodedBuilder.WriteByte(alphabet[(number % length)])
	}

	return encodedBuilder.String()
}
