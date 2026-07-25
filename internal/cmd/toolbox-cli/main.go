package toolbox

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"horizon-core-engine/internal/crypto"
	"horizon-core-engine/internal/merkle"
	"horizon-core-engine/internal/network"
)

// RunCLI اجرای ابزار خط فرمان
func RunCLI() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "generate-key":
		generateKey()
	case "hash":
		if len(os.Args) < 3 {
			log.Println("Usage: toolbox hash <text>")
			return
		}
		hashText(os.Args[2])
	case "subnet":
		if len(os.Args) < 3 {
			log.Println("Usage: toolbox subnet <cidr>")
			return
		}
		subnetCalc(os.Args[2])
	case "mac":
		if len(os.Args) < 3 {
			log.Println("Usage: toolbox mac <mac>")
			return
		}
		macLookup(os.Args[2])
	case "encrypt":
		if len(os.Args) < 4 {
			log.Println("Usage: toolbox encrypt <key-hex> <text>")
			return
		}
		encryptText(os.Args[2], os.Args[3])
	case "decrypt":
		if len(os.Args) < 4 {
			log.Println("Usage: toolbox decrypt <key-hex> <ciphertext>")
			return
		}
		decryptText(os.Args[2], os.Args[3])
	case "generate-license":
		if len(os.Args) < 5 {
			log.Println("Usage: toolbox generate-license <product-id> <user-id> <duration-hours>")
			return
		}
		generateLicenseCLI(os.Args[2], os.Args[3], os.Args[4])
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Horizon Toolbox CLI
Usage: toolbox <command> [args]

Commands:
  generate-key          Generate ECDSA key pair
  hash <text>           Calculate SHA-256 hash
  subnet <cidr>         Calculate subnet info
  mac <mac>             Lookup MAC OUI vendor
  encrypt <key> <text>  AES-GCM encrypt
  decrypt <key> <text>  AES-GCM decrypt
  generate-license <product-id> <user-id> <duration-hours>  Generate license`)
}

func generateKey() {
	privateKey, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		log.Fatal("Error generating key:", err)
	}
	fmt.Printf("Private Key: %s\n", crypto.PrivateKeyToHex(privateKey))
	fmt.Printf("Public Key:  %x\n", publicKey)
}

func hashText(text string) {
	fmt.Printf("SHA-256: %s\n", crypto.SHA256HashString(text))
}

func subnetCalc(cidr string) {
	info, err := network.CalculateSubnet(cidr)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(data))
}

func macLookup(mac string) {
	fmt.Printf("Vendor: %s\n", network.LookupOUI(mac))
}

func encryptText(keyHex, text string) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		log.Fatal("Invalid key:", err)
	}
	ciphertext, err := crypto.EncryptAES(key, text)
	if err != nil {
		log.Fatal("Encrypt error:", err)
	}
	fmt.Printf("Ciphertext: %s\n", ciphertext)
}

func decryptText(keyHex, ciphertext string) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		log.Fatal("Invalid key:", err)
	}
	plaintext, err := crypto.DecryptAES(key, ciphertext)
	if err != nil {
		log.Fatal("Decrypt error:", err)
	}
	fmt.Printf("Plaintext: %s\n", plaintext)
}

func generateLicenseCLI(productID, userID, duration string) {
	// پیاده‌سازی تولید لایسنس در CLI
	fmt.Println("License generation via CLI coming soon...")
}
