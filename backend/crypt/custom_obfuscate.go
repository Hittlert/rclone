package crypt

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
)

// Constants for custom obfuscation
const (
	ivCharLength = 2
)

// Global character set for custom obfuscation
var (
	charList      []rune
	charToIndex   map[rune]int
	indexToChar   map[int]rune
	charSetSize   int
	initialized   bool = false
)

// initCustomCharSet initializes the safe character set for custom obfuscation
func initCustomCharSet() {
	if initialized {
		return
	}

	var chars []rune

	// Add lowercase letters a-z
	for i := 'a'; i <= 'z'; i++ {
		chars = append(chars, i)
	}

	// Add uppercase letters A-Z
	for i := 'A'; i <= 'Z'; i++ {
		chars = append(chars, i)
	}

	// Add digits 0-9
	for i := '0'; i <= '9'; i++ {
		chars = append(chars, i)
	}

	// Add safe ASCII punctuation
	safeASCIIPunctuation := []rune{' ', '!', '#', '$', '%', '&', '(', ')', '+', ',', '-', '.', 
	                                ';', '=', '@', '[', ']', '^', '_', '`', '{', '}', '~'}
	chars = append(chars, safeASCIIPunctuation...)

	// Add CJK Unified Ideographs (Chinese characters)
	for i := 0x4E00; i <= 0x9FFF; i++ {
		chars = append(chars, rune(i))
	}

	// Add Hiragana
	for i := 0x3040; i <= 0x309F; i++ {
		chars = append(chars, rune(i))
	}

	// Add Katakana
	for i := 0x30A0; i <= 0x30FF; i++ {
		chars = append(chars, rune(i))
	}

	// Add Hangul Syllables (Korean characters)
	for i := 0xAC00; i <= 0xD7A3; i++ {
		chars = append(chars, rune(i))
	}

	// Remove duplicates and sort, also exclude control characters
	charSet := make(map[rune]bool)
	for _, char := range chars {
		// Exclude null character and other control characters below 0x20 except space
		if char == 0 || (char < 0x20 && char != ' ') {
			continue
		}
		charSet[char] = true
	}

	charList = make([]rune, 0, len(charSet))
	for char := range charSet {
		charList = append(charList, char)
	}
	sort.Slice(charList, func(i, j int) bool { return charList[i] < charList[j] })

	// Create mapping dictionaries
	charToIndex = make(map[rune]int)
	indexToChar = make(map[int]rune)
	for i, char := range charList {
		charToIndex[char] = i
		indexToChar[i] = char
	}

	charSetSize = len(charList)
	initialized = true
}

// getSeededRandomValue generates a deterministic random value based on key, IV, and position
func getSeededRandomValue(key string, ivNumeric int64, position int) int64 {
	ivBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(ivBytes, uint64(ivNumeric))
	
	combinedSeedMaterial := append([]byte(key), ivBytes...)
	seedHash := sha256.Sum256(combinedSeedMaterial)
	
	seed := int64(binary.BigEndian.Uint64(seedHash[:8])) + int64(position)
	rng := rand.New(rand.NewSource(seed))
	
	return rng.Int63()
}

// ivToChars converts numeric IV to character representation
func ivToChars(ivNumeric int64, length int) string {
	if charSetSize == 0 {
		panic("character set is empty, cannot generate IV characters")
	}

	maxIVValuePossible := int64(1)
	for i := 0; i < length; i++ {
		maxIVValuePossible *= int64(charSetSize)
	}
	maxIVValuePossible--

	if ivNumeric > maxIVValuePossible {
		ivNumeric = ivNumeric % (maxIVValuePossible + 1)
	}

	var ivIndices []int
	tempIV := ivNumeric
	for i := 0; i < length; i++ {
		ivIndices = append(ivIndices, int(tempIV%int64(charSetSize)))
		tempIV /= int64(charSetSize)
	}

	// Reverse the indices
	for i := 0; i < len(ivIndices)/2; i++ {
		j := len(ivIndices) - 1 - i
		ivIndices[i], ivIndices[j] = ivIndices[j], ivIndices[i]
	}

	var ivChars []rune
	for _, idx := range ivIndices {
		ivChars = append(ivChars, indexToChar[idx])
	}

	return string(ivChars)
}

// charsToIV converts character representation to numeric IV
func charsToIV(ivChars string) (int64, error) {
	if charSetSize == 0 {
		panic("character set is empty, cannot parse IV characters")
	}

	var ivNumeric int64 = 0
	for _, char := range ivChars {
		if index, exists := charToIndex[char]; exists {
			ivNumeric = ivNumeric*int64(charSetSize) + int64(index)
		} else {
			return 0, fmt.Errorf("IV character '%c' not in character set, cannot parse", char)
		}
	}

	return ivNumeric, nil
}

// getDeterministicPermutationMaps generates deterministic permutation maps for position shuffling
func getDeterministicPermutationMaps(key string, ivNumeric int64, length int) ([]int, []int) {
	if length == 0 {
		return []int{}, []int{}
	}

	// Create permutation seed
	ivBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(ivBytes, uint64(ivNumeric))
	
	permutationSeedMaterial := append([]byte(key), ivBytes...)
	permutationSeedMaterial = append(permutationSeedMaterial, []byte("POS_PERM_SALT")...)
	permutationSeed := sha256.Sum256(permutationSeedMaterial)
	
	seed := int64(binary.BigEndian.Uint64(permutationSeed[:8]))
	permRng := rand.New(rand.NewSource(seed))

	// Initialize original to new position map
	originalToNewPosMap := make([]int, length)
	for i := 0; i < length; i++ {
		originalToNewPosMap[i] = i
	}

	// Fisher-Yates shuffle
	for i := length - 1; i > 0; i-- {
		j := permRng.Intn(i + 1)
		originalToNewPosMap[i], originalToNewPosMap[j] = originalToNewPosMap[j], originalToNewPosMap[i]
	}

	// Build reverse mapping
	newToOriginalPosMap := make([]int, length)
	for originalIdx, newIdx := range originalToNewPosMap {
		newToOriginalPosMap[newIdx] = originalIdx
	}

	return originalToNewPosMap, newToOriginalPosMap
}

// customObfuscateText implements the custom obfuscation algorithm
func customObfuscateText(text string, key string) string {
	// Handle empty string specially
	if text == "" {
		return ""
	}
	
	// Initialize character set if not done
	initCustomCharSet()

	// Step 1: Generate deterministic IV based on key and original text content
	combinedSeedForIV := key + text
	ivHash := sha256.Sum256([]byte(combinedSeedForIV))
	// Use absolute value to ensure positive number
	ivNumeric := int64(binary.BigEndian.Uint64(ivHash[:8]) & 0x7FFFFFFFFFFFFFFF)

	// Ensure IV fits within character set bounds
	maxIVValuePossible := int64(1)
	for i := 0; i < ivCharLength; i++ {
		maxIVValuePossible *= int64(charSetSize)
	}
	maxIVValuePossible--

	if ivNumeric > maxIVValuePossible {
		ivNumeric = ivNumeric % (maxIVValuePossible + 1)
	}

	ivPrefixChars := ivToChars(ivNumeric, ivCharLength)

	// Step 2: Generate permutation mapping (original_pos -> new_pos)
	textRunes := []rune(text)
	originalToNewPosMap, _ := getDeterministicPermutationMaps(key, ivNumeric, len(textRunes))

	// Step 3: Apply position permutation: shuffle original string positions
	shuffledTextList := make([]rune, len(textRunes))
	for originalPos, newPos := range originalToNewPosMap {
		shuffledTextList[newPos] = textRunes[originalPos]
	}

	// Step 4: Character shifting: obfuscate each character in the shuffled string
	obfuscatedChars := make([]rune, len(shuffledTextList))
	for i, char := range shuffledTextList {
		if index, exists := charToIndex[char]; exists {
			// Character is in our character set, apply shifting
			effectiveShiftAmountBase := getSeededRandomValue(key, ivNumeric, i)
			effectiveShiftAmount := int(effectiveShiftAmountBase % int64(charSetSize))
			
			newCharSetIndex := (index + effectiveShiftAmount) % charSetSize
			obfuscatedChars[i] = indexToChar[newCharSetIndex]
		} else {
			// Character not in our character set, keep as-is
			obfuscatedChars[i] = char
		}
	}

	return ivPrefixChars + string(obfuscatedChars)
}

// customDeobfuscateText implements the custom deobfuscation algorithm
func customDeobfuscateText(obfuscatedText string, key string) (string, error) {
	// Handle empty string specially
	if obfuscatedText == "" {
		return "", nil
	}
	
	// Initialize character set if not done
	initCustomCharSet()

	if len([]rune(obfuscatedText)) < ivCharLength {
		return "", fmt.Errorf("obfuscated string too short, cannot extract %d IV characters", ivCharLength)
	}

	obfuscatedRunes := []rune(obfuscatedText)
	ivPrefixChars := string(obfuscatedRunes[:ivCharLength])
	actualObfuscatedData := obfuscatedRunes[ivCharLength:]

	ivNumeric, err := charsToIV(ivPrefixChars)
	if err != nil {
		return "", err
	}

	// Step 1: Character reverse shifting: restore character original values, 
	// but they are still in shuffled positions
	intermediateUnshuffledChars := make([]rune, len(actualObfuscatedData))
	for i, char := range actualObfuscatedData {
		if index, exists := charToIndex[char]; exists {
			// Character is in our character set, apply reverse shifting
			effectiveShiftAmountBase := getSeededRandomValue(key, ivNumeric, i)
			effectiveShiftAmount := int(effectiveShiftAmountBase % int64(charSetSize))
			
			originalIndexInCharSet := (index - effectiveShiftAmount + charSetSize) % charSetSize
			intermediateUnshuffledChars[i] = indexToChar[originalIndexInCharSet]
		} else {
			// Character not in our character set, keep as-is
			intermediateUnshuffledChars[i] = char
		}
	}

	// Step 2: Generate reverse permutation mapping (new_pos -> original_pos)
	_, newToOriginalPosMap := getDeterministicPermutationMaps(key, ivNumeric, len(actualObfuscatedData))

	// Step 3: Apply reverse permutation: restore characters to their original positions
	finalOriginalTextList := make([]rune, len(actualObfuscatedData))
	for newPos, originalPos := range newToOriginalPosMap {
		finalOriginalTextList[originalPos] = intermediateUnshuffledChars[newPos]
	}

	return string(finalOriginalTextList), nil
}