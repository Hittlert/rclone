package crypt

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
)

// Constants for custom obfuscation v6.2
const (
	// IV is always encoded as exactly 3 bytes in UTF-8
	ivByteLength = 3
)

// Global character sets for custom obfuscation v6.2
var (
	// ASCII character set
	asciiChars          []rune
	asciiCharToIndex    map[rune]int
	
	// CJK character set (unified, no longer separated)
	cjkChars            []rune
	cjkCharToIndex      map[rune]int
	
	initialized         bool = false
)

// initCustomCharSet initializes the character sets for custom obfuscation v6.2
func initCustomCharSet() {
	if initialized {
		return
	}

	// Initialize ASCII character set
	asciiSet := make(map[rune]bool)
	
	// Add lowercase letters a-z
	for i := 'a'; i <= 'z'; i++ {
		asciiSet[i] = true
	}

	// Add uppercase letters A-Z
	for i := 'A'; i <= 'Z'; i++ {
		asciiSet[i] = true
	}

	// Add digits 0-9
	for i := '0'; i <= '9'; i++ {
		asciiSet[i] = true
	}

	// Add safe ASCII punctuation
	safeASCIIPunctuation := []rune{'!', '#', '$', '%', '&', '\'', '(', ')', '+', ',', '-', '.', 
	                                ';', '=', '@', '[', ']', '^', '_', '`', '{', '}', '~'}
	for _, char := range safeASCIIPunctuation {
		asciiSet[char] = true
	}

	// Convert ASCII set to sorted slice
	asciiChars = make([]rune, 0, len(asciiSet))
	for char := range asciiSet {
		asciiChars = append(asciiChars, char)
	}
	sort.Slice(asciiChars, func(i, j int) bool { return asciiChars[i] < asciiChars[j] })

	// Create ASCII char to index mapping
	asciiCharToIndex = make(map[rune]int)
	for i, char := range asciiChars {
		asciiCharToIndex[char] = i
	}

	// Initialize unified CJK character set (no longer separated)
	cjkSet := make(map[rune]bool)

	// Add CJK Unified Ideographs (Chinese characters)
	for i := 0x4E00; i <= 0x9FFF; i++ {
		cjkSet[rune(i)] = true
	}

	// Add Hiragana
	for i := 0x3040; i <= 0x309F; i++ {
		cjkSet[rune(i)] = true
	}

	// Add Katakana
	for i := 0x30A0; i <= 0x30FF; i++ {
		cjkSet[rune(i)] = true
	}

	// Convert CJK set to sorted slice
	cjkChars = make([]rune, 0, len(cjkSet))
	for char := range cjkSet {
		cjkChars = append(cjkChars, char)
	}
	sort.Slice(cjkChars, func(i, j int) bool { return cjkChars[i] < cjkChars[j] })

	// Create CJK char to index mapping
	cjkCharToIndex = make(map[rune]int)
	for i, char := range cjkChars {
		cjkCharToIndex[char] = i
	}

	initialized = true
}

// getCharType determines the character type for v6.2 algorithm
func getCharType(char rune) string {
	if _, exists := asciiCharToIndex[char]; exists {
		return "ASCII"
	}
	if _, exists := cjkCharToIndex[char]; exists {
		return "CJK"
	}
	return "UNKNOWN"
}

// generateSmartIV generates a 3-byte smart IV based on v6.2 algorithm
func generateSmartIV(text string, key string) string {
	seedData := text + key + "IV_V6.2"
	hashValue := sha256.Sum256([]byte(seedData))
	
	// Use hash to determine IV type
	ivTypeSelector := binary.BigEndian.Uint16(hashValue[0:2]) % 2
	
	if ivTypeSelector == 0 && len(cjkChars) > 0 {
		// CJK IV: use 1 CJK character (3 bytes)
		ivIndex := binary.BigEndian.Uint32(hashValue[2:6]) % uint32(len(cjkChars))
		return string(cjkChars[ivIndex])
	} else {
		// ASCII IV: use 3 ASCII characters (3 bytes)
		ivChars := make([]rune, 3)
		seedVal := binary.BigEndian.Uint32(hashValue[2:6])
		for i := 0; i < 3; i++ {
			charIndex := (seedVal + uint32(i)*uint32(hashValue[i+6])) % uint32(len(asciiChars))
			ivChars[i] = asciiChars[charIndex]
		}
		return string(ivChars)
	}
}
// parseIVByBytes parses IV using the head-3-bytes method from v6.2
func parseIVByBytes(obfuscatedText string) (string, string, error) {
	if obfuscatedText == "" {
		return "", "", nil
	}
	
	// Convert to bytes and check if we have at least 3 bytes
	encodedBytes := []byte(obfuscatedText)
	if len(encodedBytes) < ivByteLength {
		return "", "", fmt.Errorf("encoded text is less than %d bytes", ivByteLength)
	}
	
	// Extract first 3 bytes as IV
	headerBytes := encodedBytes[:ivByteLength]
	ivString := string(headerBytes)
	
	// Calculate how many characters the IV string contains
	ivRunes := []rune(ivString)
	ivCharLen := len(ivRunes)
	
	// Extract the data part (remaining characters after IV)
	allRunes := []rune(obfuscatedText)
	if len(allRunes) < ivCharLen {
		return "", "", fmt.Errorf("obfuscated text too short")
	}
	
	dataString := string(allRunes[ivCharLen:])
	return ivString, dataString, nil
}

// generatePositionShuffle generates deterministic position shuffle based on v5.3 algorithm (Go native)
func generatePositionShuffle(key string, iv string, textLen int) []int {
	if textLen == 0 {
		return []int{}
	}
	
	shuffleSeedData := key + iv + "SHUFFLE" + fmt.Sprintf("%d", textLen)
	hashValue := sha256.Sum256([]byte(shuffleSeedData))
	
	// Use hash directly as seed for Go's native random generator
	seed := int64(binary.BigEndian.Uint64(hashValue[:8]))
	rng := rand.New(rand.NewSource(seed))
	
	// Create position array and shuffle (Fisher-Yates)
	positions := make([]int, textLen)
	for i := 0; i < textLen; i++ {
		positions[i] = i
	}
	
	// Fisher-Yates shuffle using Go's algorithm
	for i := textLen - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		positions[i], positions[j] = positions[j], positions[i]
	}
	
	return positions
}

// mapCharacter maps a single character based on v6.2 algorithm
func mapCharacter(char rune, newPos int, textLen int, key string, iv string) rune {
	charType := getCharType(char)
	mapSeedData := key + iv + fmt.Sprintf("%d", newPos) + fmt.Sprintf("%d", textLen) + charType
	hashValue := sha256.Sum256([]byte(mapSeedData))
	offset := binary.BigEndian.Uint64(hashValue[:8])

	switch charType {
	case "ASCII":
		if index, exists := asciiCharToIndex[char]; exists {
			newIndex := (uint64(index) + offset) % uint64(len(asciiChars))
			return asciiChars[newIndex]
		}
	case "CJK":
		if index, exists := cjkCharToIndex[char]; exists {
			newIndex := (uint64(index) + offset) % uint64(len(cjkChars))
			return cjkChars[newIndex]
		}
	}
	
	// Unknown character type, return as-is
	return char
}

// reverseMapCharacter reverses character mapping based on v6.2 algorithm
func reverseMapCharacter(mappedChar rune, newPos int, textLen int, key string, iv string) rune {
	charType := getCharType(mappedChar)
	mapSeedData := key + iv + fmt.Sprintf("%d", newPos) + fmt.Sprintf("%d", textLen) + charType
	hashValue := sha256.Sum256([]byte(mapSeedData))
	offset := binary.BigEndian.Uint64(hashValue[:8])

	switch charType {
	case "ASCII":
		if index, exists := asciiCharToIndex[mappedChar]; exists {
			charSetSize := uint64(len(asciiChars))
			offsetMod := offset % charSetSize
			originalIndex := (uint64(index) + charSetSize - offsetMod) % charSetSize
			return asciiChars[originalIndex]
		}
	case "CJK":
		if index, exists := cjkCharToIndex[mappedChar]; exists {
			charSetSize := uint64(len(cjkChars))
			offsetMod := offset % charSetSize
			originalIndex := (uint64(index) + charSetSize - offsetMod) % charSetSize
			return cjkChars[originalIndex]
		}
	}
	
	// Unknown character type, return as-is
	return mappedChar
}

// CustomObfuscateText implements the custom obfuscation algorithm v6.2
func CustomObfuscateText(text string, key string) string {
	// Handle empty string specially
	if text == "" {
		return ""
	}
	
	// Initialize character set if not done
	initCustomCharSet()
	
	// Step 1: Generate smart IV
	iv := generateSmartIV(text, key)
	
	// Step 2: Generate position shuffle
	textRunes := []rune(text)
	textLen := len(textRunes)
	shufflePositions := generatePositionShuffle(key, iv, textLen)
	
	// Step 3: Apply position shuffle and character mapping
	obfuscatedChars := make([]rune, textLen)
	for newPos, oldPos := range shufflePositions {
		char := textRunes[oldPos]
		mappedChar := mapCharacter(char, newPos, textLen, key, iv)
		obfuscatedChars[newPos] = mappedChar
	}
	
	return iv + string(obfuscatedChars)
}

// internalDeobfuscateText implements the internal deobfuscation logic without validation
func internalDeobfuscateText(obfuscatedText string, key string) (string, error) {
	// Initialize character set if not done
	initCustomCharSet()
	
	// Step 1: Parse IV using head-3-bytes method
	iv, dataPart, err := parseIVByBytes(obfuscatedText)
	if err != nil {
		return "", err
	}

	// Step 2: Generate same position shuffle
	dataRunes := []rune(dataPart)
	dataLen := len(dataRunes)
	shufflePositions := generatePositionShuffle(key, iv, dataLen)

	// Step 3: Reverse character mapping
	reversedMapChars := make([]rune, dataLen)
	for newPos, char := range dataRunes {
		reversedChar := reverseMapCharacter(char, newPos, dataLen, key, iv)
		reversedMapChars[newPos] = reversedChar
	}

	// Step 4: Reverse position shuffle
	originalChars := make([]rune, dataLen)
	for newPos, oldPos := range shufflePositions {
		originalChars[oldPos] = reversedMapChars[newPos]
	}

	return string(originalChars), nil
}

// CustomDeobfuscateText implements the custom deobfuscation algorithm v6.2 with integrity validation
func CustomDeobfuscateText(obfuscatedText string, key string) (string, error) {
	// Handle empty string specially
	if obfuscatedText == "" {
		return "", nil
	}
	
	// Step 1: Normal deobfuscation
	recoveredText, err := internalDeobfuscateText(obfuscatedText, key)
	if err != nil {
		return "", err // If initial deobfuscation fails, return error
	}

	// Step 2: Re-obfuscate the recovered text
	reObfuscatedText := CustomObfuscateText(recoveredText, key)

	// Step 3: Compare with original obfuscated text (zero-overhead integrity check)
	if reObfuscatedText == obfuscatedText {
		// If consistent, deobfuscation was successful
		return recoveredText, nil
	} else {
		// Inconsistent, indicates invalid input or wrong key
		return "", fmt.Errorf("integrity check failed: invalid input or wrong key")
	}
}