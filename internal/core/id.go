package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const idCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const idLength = 10

// GenerateID generates a random alphanumeric ID of length 10.
func GenerateID() string {
	b := make([]byte, idLength)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(idCharset))))
		if err != nil {
			// Fallback to less secure random if crypto/rand fails (unlikely)
			return generateFallbackID()
		}
		b[i] = idCharset[num.Int64()]
	}
	return string(b)
}

func generateFallbackID() string {
	// Simple fallback, though in practice crypto/rand shouldn't fail
	return "fallback" + fmt.Sprintf("%d", time.Now().UnixNano())[:2]
}

// Agent name letters for sequential assignment.
var agentLetters = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
	"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T",
	"U", "V", "W", "X", "Y", "Z",
}

// GenerateMaskedName generates a consistent masked name for an agent.
// The name is deterministic based on the agent's ID, ensuring the same
// agent always gets the same masked name within a session.
func GenerateMaskedName(agentID string) string {
	if agentID == "" {
		return "Agent ?"
	}
	// Use the agent ID to deterministically select a letter
	hash := 0
	for _, c := range agentID {
		hash = hash*31 + int(c)
	}
	index := (hash & 0x7FFFFFFF) % len(agentLetters)
	return fmt.Sprintf("Agent %s", agentLetters[index])
}

// GenerateMaskedNamesForList generates unique masked names for a list of agents.
// This is used when creating a debate/council to ensure all agents get unique names.
func GenerateMaskedNamesForList(agentIDs []string) map[string]string {
	names := make(map[string]string)
	usedLetters := make(map[string]bool)

	// First, try to assign deterministically based on ID
	for _, id := range agentIDs {
		hash := 0
		for _, c := range id {
			hash = hash*31 + int(c)
		}
		index := (hash & 0x7FFFFFFF) % len(agentLetters)
		letter := agentLetters[index]

		// If collision, find next available
		for usedLetters[letter] {
			index = (index + 1) % len(agentLetters)
			letter = agentLetters[index]
		}

		usedLetters[letter] = true
		names[id] = fmt.Sprintf("Agent %s", letter)
	}

	return names
}
