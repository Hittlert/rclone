package crypt

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomObfuscation tests the custom obfuscation functions
func TestCustomObfuscation(t *testing.T) {
	key := "CLTIhytye2gMPBwVsMZ3SEnQqohFe2LTCiWeB0XCyO5sQpte0jY5siKFbC1rpgOl"
	
	testCases := []struct {
		name     string
		input    string
		expected string  // We won't check exact output as it's deterministic but complex
	}{
		{
			name:  "simple filename",
			input: "test2.py",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "filename with spaces",
			input: "my test file.txt",
		},
		{
			name:  "filename with unicode",
			input: "测试文件.txt",
		},
		{
			name:  "complex filename",
			input: "Hello 🤔 World! Some weird char \x01\u01F914",
		},
		{
			name:  "long filename",
			input: "this_is_a_very_long_filename_with_many_characters_to_test_the_obfuscation_algorithm.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test obfuscation
			obfuscated := customObfuscateText(tc.input, key)
			
			// Test that obfuscation changes the input (except for empty string)
			if tc.input != "" {
				assert.NotEqual(t, tc.input, obfuscated, "Obfuscated text should differ from input")
			} else {
				assert.Equal(t, "", obfuscated, "Empty input should result in empty output")
			}
			
			// Test deobfuscation
			deobfuscated, err := customDeobfuscateText(obfuscated, key)
			require.NoError(t, err, "Deobfuscation should not error")
			assert.Equal(t, tc.input, deobfuscated, "Deobfuscated text should match original input")
		})
	}
}

// TestCustomObfuscationConsistency tests that the same input produces the same output
func TestCustomObfuscationConsistency(t *testing.T) {
	key := "CLTIhytye2gMPBwVsMZ3SEnQqohFe2LTCiWeB0XCyO5sQpte0jY5siKFbC1rpgOl"
	input := "test2.py"
	
	// Obfuscate multiple times
	obfuscated1 := customObfuscateText(input, key)
	obfuscated2 := customObfuscateText(input, key)
	
	assert.Equal(t, obfuscated1, obfuscated2, "Same input should produce same obfuscated output")
}

// TestCustomObfuscationWithDifferentKeys tests that different keys produce different outputs
func TestCustomObfuscationWithDifferentKeys(t *testing.T) {
	key1 := "key1"
	key2 := "key2"
	input := "test2.py"
	
	obfuscated1 := customObfuscateText(input, key1)
	obfuscated2 := customObfuscateText(input, key2)
	
	assert.NotEqual(t, obfuscated1, obfuscated2, "Different keys should produce different obfuscated outputs")
}

// TestCipherCustomObfuscation tests the cipher methods for custom obfuscation
func TestCipherCustomObfuscation(t *testing.T) {
	// Create a cipher with custom mode
	cipher, err := newCipher(NameEncryptionCustom, "password", "", true, nil)
	require.NoError(t, err)

	testCases := []string{
		"test.txt",
		"document.pdf",
		"folder",
		"file with spaces.doc",
		"测试文件.txt",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			// Test segment obfuscation
			obfuscated := cipher.customObfuscateSegment(testCase)
			if testCase != "" {
				assert.NotEqual(t, testCase, obfuscated, "Obfuscated segment should differ from input")
			}

			// Test segment deobfuscation
			deobfuscated, err := cipher.customDeobfuscateSegment(obfuscated)
			require.NoError(t, err, "Deobfuscation should not error")
			assert.Equal(t, testCase, deobfuscated, "Deobfuscated segment should match original")
		})
	}
}

// TestCipherCustomFilenameEncryption tests the full filename encryption/decryption
func TestCipherCustomFilenameEncryption(t *testing.T) {
	// Create a cipher with custom mode
	cipher, err := newCipher(NameEncryptionCustom, "password", "", true, nil)
	require.NoError(t, err)

	testCases := []string{
		"test.txt",
		"folder/file.txt",
		"deep/folder/structure/file.doc",
		"测试/文件夹/文件.txt",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			// Test filename encryption
			encrypted := cipher.EncryptFileName(testCase)
			assert.NotEqual(t, testCase, encrypted, "Encrypted filename should differ from input")

			// Test filename decryption
			decrypted, err := cipher.DecryptFileName(encrypted)
			require.NoError(t, err, "Decryption should not error")
			assert.Equal(t, testCase, decrypted, "Decrypted filename should match original")
		})
	}
}

// TestNameEncryptionModeCustom tests the NameEncryptionMode enumeration for custom mode
func TestNameEncryptionModeCustom(t *testing.T) {
	// Test string to mode conversion
	mode, err := NewNameEncryptionMode("custom")
	require.NoError(t, err)
	assert.Equal(t, NameEncryptionCustom, mode)

	// Test mode to string conversion
	modeStr := mode.String()
	assert.Equal(t, "custom", modeStr)
}