package utils

import (
	"fmt"
	"math"
)

func PerformComputation(partition []Review) ResultData {
	// Map of user IDs to their ratings
	userRatings := make(map[string]map[string]float64)

	// Build user ratings from the partition data
	for _, review := range partition {
		if _, exists := userRatings[review.ReviewerID]; !exists {
			userRatings[review.ReviewerID] = make(map[string]float64)
		}
		userRatings[review.ReviewerID][review.ProductID] = review.Stars
	}
	// Compute similarities between users
	similarities := computeUserSimilarities(userRatings)

	// Generate recommendations for each user
	recommendations := make(map[string][]string)
	for userID := range userRatings {
		recs := recommendProducts(userID, userRatings, similarities)
		recommendations[userID] = recs

		// **Add this line to log recommendations per user**
		// fmt.Printf("User %s recommendations: %v\n", userID, recs)
	}

	return ResultData{Recommendations: recommendations}
}

func computeUserSimilarities(
	userRatings map[string]map[string]float64,
) map[string]map[string]float64 {
	similarities := make(map[string]map[string]float64)

	users := make([]string, 0, len(userRatings))
	for user := range userRatings {
		users = append(users, user)
	}

	for i := 0; i < len(users); i++ {
		userA := users[i]
		ratingsA := userRatings[userA]

		for j := i + 1; j < len(users); j++ {
			userB := users[j]
			ratingsB := userRatings[userB]

			sim := calculateCosineSimilarity(ratingsA, ratingsB)
			if sim > 0 {
				if _, exists := similarities[userA]; !exists {
					similarities[userA] = make(map[string]float64)
				}
				if _, exists := similarities[userB]; !exists {
					similarities[userB] = make(map[string]float64)
				}
				similarities[userA][userB] = sim
				similarities[userB][userA] = sim

				// **Add this line to log similarities**
				fmt.Printf("Similarity between user %s and user %s: %f\n", userA, userB, sim)
			}
		}
	}
	return similarities
}

func recommendProducts(
	targetUserID string,
	userRatings map[string]map[string]float64,
	similarities map[string]map[string]float64,
) []string {
	scores := make(map[string]float64)
	totalSim := make(map[string]float64)

	for otherUserID, sim := range similarities[targetUserID] {
		for productID, rating := range userRatings[otherUserID] {
			if _, rated := userRatings[targetUserID][productID]; !rated {
				scores[productID] += sim * rating
				totalSim[productID] += sim
			}
		}
	}

	recommendations := make([]string, 0)
	for productID := range scores {
		// Compute the weighted average
		if totalSim[productID] != 0 {
			score := scores[productID] / totalSim[productID]
			if score >= 2.0 { // Threshold for recommendation
				recommendations = append(recommendations, productID)
			}
			fmt.Printf("User %s, Product %s, Score: %f\n", targetUserID, productID, score)
		}
	}

	return recommendations
}

func calculateCosineSimilarity(ratingsA, ratingsB map[string]float64) float64 {
	var sumProduct, sumASq, sumBSq float64

	for itemID, ratingA := range ratingsA {
		if ratingB, exists := ratingsB[itemID]; exists {
			sumProduct += ratingA * ratingB
			sumASq += ratingA * ratingA
			sumBSq += ratingB * ratingB
		}
	}

	denominator := math.Sqrt(sumASq) * math.Sqrt(sumBSq)
	if denominator == 0 {
		return 0
	}
	return sumProduct / denominator
}
