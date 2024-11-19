package utils

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var productCategoryMap map[string]string // Map[productID]productCategory

func ShuffleData(data []Review) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})
}

func LoadData(filePath string) []Review {
	filePath = "/app/" + filePath
	fmt.Printf("Attempting to load data from %s\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening data file %s: %v", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV data from %s: %v", filePath, err)
	}

	if len(records) == 0 {
		log.Fatalf("No records found in CSV file %s", filePath)
	}

	var reviews []Review
	productCategoryMap = make(map[string]string)

	// Assuming the first row is the header
	for _, record := range records[1:] {
		i_stars, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Error parsing stars for review ID %s: %v", record[0], err)
			continue // Skip this record if stars cannot be parsed
		}
		stars := float64(i_stars)

		review := Review{
			ReviewID:        record[1],
			ProductID:       record[2],
			ReviewerID:      record[3],
			Stars:           stars,
			ProductCategory: strings.ToLower(strings.TrimSpace(record[8])),
		}
		// log.Printf("Parsed review: %+v\n", review)
		reviews = append(reviews, review)
		productCategoryMap[review.ProductID] = review.ProductCategory
	}

	return reviews
}

func GetProductCategory(productID string) string {
	category := productCategoryMap[productID]
	fmt.Printf("Product ID %s has category '%s'\n", productID, category) // **Add this line**
	return category
}

func SplitData(data []Review, partitions int) [][]Review {
	// ShuffleData(data) // Shuffle data before splitting
	var result [][]Review
	partitionSize := (len(data) + partitions - 1) / partitions

	for i := 0; i < len(data); i += partitionSize {
		end := i + partitionSize
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}
	return result
}
