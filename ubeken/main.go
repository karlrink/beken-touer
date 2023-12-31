package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/fernet/fernet-go"
	_ "github.com/mattn/go-sqlite3"

	"ubeken/kes"
)

var version = "1.0.0.🎃-2023-10-02"

func usage() {

	fmt.Println(`Usage: ` + os.Args[0] + ` </path/db> [port]

  --help|-help|help           Display this help message
  --version|-version|version  Display version

<db> 9480 # Default `)
}

func main() {

	//DEBUG=1 go run main.go
	PrintDebug("Debug mode enabled")

	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "--help", "-help", "help":
		usage()
		return
	case "--version", "-version", "version":
		fmt.Println("Version: " + version)
		sqlite3version, err := getSqlite3Version(os.Args[1])
		if err != nil {
			log.Fatal("Failed to get SQLite version: %v\n", err)
		}
		fmt.Println("Sqlite3: " + sqlite3version)
		return

	}

	dbFile := os.Args[1]

	var defaultPort = "9480"

	if len(os.Args) > 2 {
		defaultPort = os.Args[2]
	}

	// Define the address to listen on
	address := ":" + defaultPort // ":9480"

	// Configure the log package to write to standard output (os.Stdout).
	log.SetOutput(os.Stdout)

	// Resolve the UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal("Error resolving UDP address:", err)
	}

	// Create a UDP connection
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("Error creating UDP connection:", err)
	}
	defer conn.Close()

	// Create a buffer to hold incoming data
	buffer := make([]byte, 1024)

	// Open or create the SQLite3 database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	create := createTables(db)
	if create != nil {
		log.Fatal("Error creating tables:", create)
	}

	log.Println("UDP server listening on " + address)

	for {

		// Read data from the connection
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from UDP connection: %v\n", err)
			continue
		}

		receivedData := string(buffer[:n])

		// Process each packet in a Goroutine
		go func(data string, clientAddr *net.UDPAddr) {

			// Convert clientAddr to a string
			clientAddrStr := clientAddr.String()

			// Remove the port from the IP address
			host, _, err := net.SplitHostPort(clientAddrStr) // works for IPv4 and IPv6 addresses
			if err != nil {
				log.Printf("Failed to split host and port: %v\n", err)
			}
			log.Println("UDP connection " + host)

			str := strings.Split(data, " ")

			fmt.Println(str)

			//PrintDebug("Length " + string(len(str))) // empty
			PrintDebug("Length " + strconv.Itoa(len(str)))

			strLength := len(str)

			if strLength == 1 {
				//PrintDebug("one1")
				switch str[0] {
				case "PUBLIC_KEY", "RSA_PUBLIC_KEY":
					log.Println("PUBLIC_KEY request " + host)
					return
				}
				return
			}

			if strLength == 2 {
				log.Println("Insufficient element split " + host + " " + data)
				return
			}
			//log.Println("Not found " + host + " " + data)
			//return

			if strLength >= 3 {

				//PrintDebug("hello.hello.hello")

				field1 := str[0] //name
				field2 := str[1] //code
				//field3 := str[2] //cypher

				exists, decrypted := existsAndDecrypts(db, data)

				if !exists {
					PrintDebug("Not ExistAndDecrypts")
					return
				}

				PrintDebug(data)
				PrintDebug("decrypted: " + decrypted)

				switch field2 {
				case "0": //kes

					// Save the IP to the database
					_, err := db.Exec("INSERT INTO ips (Name, Data) VALUES (?, ?)", host, field1)
					if err != nil {
						log.Printf("Failed to save IP to database: %v\n", err)
					} else {
						log.Printf("Isert IP " + host)
					}

					// Send a response back to the client
					response := host

					_, err_response := conn.WriteToUDP([]byte(response), clientAddr)
					if err_response != nil {
						log.Printf("Error sending response to client: %v\n", err_response)
					}
					log.Println("Sent response " + host)

				case "1": //rsa

					// Save the IP to the database
					_, err := db.Exec("INSERT INTO ips (Name, Data) VALUES (?, ?)", host, field1)
					if err != nil {
						log.Printf("Failed to save IP to database: %v\n", err)
					} else {
						log.Printf("Isert IP " + host)
					}

					// Send a response back to the client
					response := host

					_, err_response := conn.WriteToUDP([]byte(response), clientAddr)
					if err_response != nil {
						log.Printf("Error sending response to client: %v\n", err_response)
					}
					log.Println("Sent response " + host)

				case "2": //fernet

					// Save the IP to the database
					_, err := db.Exec("INSERT INTO ips (Name, Data) VALUES (?, ?)", host, field1)
					if err != nil {
						log.Printf("Failed to save IP to database: %v\n", err)
					} else {
						log.Println("Isert IP " + host)
					}
					// Send a response back to the client
					response := host

					_, err_response := conn.WriteToUDP([]byte(response), clientAddr)
					if err_response != nil {
						log.Printf("Error sending response to client: %v\n", err_response)
					}
					log.Println("Sent response " + host)

				case "3": //aes

					// Save the IP to the database
					_, err := db.Exec("INSERT INTO ips (Name, Data) VALUES (?, ?)", host, field1)
					if err != nil {
						log.Printf("Failed to save IP to database: %v\n", err)
					} else {
						log.Println("Isert IP " + host)
					}
					// Send a response back to the client
					response := host

					_, err_response := conn.WriteToUDP([]byte(response), clientAddr)
					if err_response != nil {
						log.Printf("Error sending response to client: %v\n", err_response)
					}
					log.Println("Sent response " + host)

				case "X": //XOR

					// Save the IP to the database
					_, err := db.Exec("INSERT INTO ips (Name, Data) VALUES (?, ?)", host, field1)
					if err != nil {
						log.Printf("Failed to save IP to database: %v\n", err)
					} else {
						log.Println("Isert IP " + host)
					}

					// Send a response back to the client
					response := host

					_, err_response := conn.WriteToUDP([]byte(response), clientAddr)
					if err_response != nil {
						log.Printf("Error sending response to client: %v\n", err_response)
					}
					log.Println("Sent response " + host)

				case "A1": //aes128

					// Verify decrypted
					pattern := `^Beken`

					// Compile the regular expression pattern
					re := regexp.MustCompile(pattern)

					// Use FindString to check if the string starts with "Beken"
					if re.FindString(decrypted) != "Beken" {
						log.Println("Failed decrypt Beken " + host)
						return
					}

					var inDB string
					// Save the IP to the database
					_, err := db.Exec("INSERT INTO ips (Name, Data) VALUES (?, ?)", host, field1)
					if err != nil {
						inDB = "True"
						PrintDebugf("Failed to save IP to database: %v\n", err)
					} else {
						inDB = "New"
						PrintDebug("Isert IP " + host)
					}

					// Send a response back to the client
					response := host + " " + inDB

					_, err_response := conn.WriteToUDP([]byte(response), clientAddr)
					if err_response != nil {
						log.Printf("Error sending response to client: %v\n", err_response)
					}
					log.Println("Sent response " + host)

				default:
					log.Println("code not found: " + field2 + " " + host)

				}

			}

		}(receivedData, addr)
	}
}

func PrintDebug(message string) {
	debugMode := os.Getenv("DEBUG")
	if debugMode != "" {
		fmt.Println(message)
	}
}

func PrintDebugf(format string, err error) {
	debugMode := os.Getenv("DEBUG")
	if debugMode != "" {
		fmt.Printf(format, err)
	}
}

func existsAndDecrypts(db *sql.DB, dataStr string) (bool, string) {

	var exists bool
	var key string

	str := strings.Split(dataStr, " ")

	field1 := str[0] //name
	field2 := str[1] //code
	field3 := str[2] //cypher

	switch field2 {

	case "0": //kes
		err_query := db.QueryRow("SELECT EXISTS (SELECT 1 FROM kes_keys WHERE Name = ?), Data FROM kes_keys WHERE Name = ?", field1, field1).Scan(&exists, &key)
		if err_query != nil {
			log.Println("Error QueryRow database:", err_query)
			return false, ""
		}
		PrintDebug("Exists in db: " + field1)

		decrypted := kes.DecryptKES(field3, key)
		if decrypted == "" {
			log.Println("Error decrypt xor: empty")
			return false, ""
		}

		return true, decrypted

	case "1": //rsa
		err_query := db.QueryRow("SELECT EXISTS (SELECT 1 FROM private_keys WHERE Name = ?), Data FROM private_keys WHERE Name = ?", field1, field1).Scan(&exists, &key)
		if err_query != nil {
			log.Println("Error QueryRow database:", err_query)
			return false, ""
		}
		PrintDebug("Exists in db: " + field1)

		decrypted, err := decryptRSA(field3, key)
		if err != nil {
			log.Println("Error decrypt rsa:", err)
			return false, ""
		}

		return true, decrypted

	case "2": //fernet
		err_query := db.QueryRow("SELECT EXISTS (SELECT 1 FROM fernet_keys WHERE Name = ?), Data FROM fernet_keys WHERE Name = ?", field1, field1).Scan(&exists, &key)
		if err_query != nil {
			log.Println("Error QueryRow database:", err_query)
			return false, ""
		}
		PrintDebug("Exists in db: " + field1)

		decrypted, err := decryptFernet(field3, key)
		if err != nil {
			log.Println("Error decrypt fernet:", err)
			return false, ""
		}

		return true, decrypted

	case "3": //aes gcm
		//field1 := str[0] //name
		//field2 := str[1] //code
		//field3 := str[2] //cypher
		field4 := str[3] //nonce
		field5 := str[4] //tag

		err_query := db.QueryRow("SELECT EXISTS (SELECT 1 FROM aes_keys WHERE Name = ?), Data FROM aes_keys WHERE Name = ?", field1, field1).Scan(&exists, &key)
		if err_query != nil {
			log.Println("Error QueryRow database:", err_query)
			return false, ""
		}
		PrintDebug("Exists in db: " + field1)

		decrypted, err := decryptAESGCM(field3, field4, field5, key)
		if err != nil {
			log.Println("Error decrypt aes:", err)
			return false, ""
		}

		return exists, decrypted

	case "X": //XOR
		//field1 := str[0] //name
		//field2 := str[1] //code
		//field3 := str[2] //cypher

		err_query := db.QueryRow("SELECT EXISTS (SELECT 1 FROM ubeken_keys WHERE Name = ?), Data FROM ubeken_keys WHERE Name = ?", field1, field1).Scan(&exists, &key)
		if err_query != nil {
			log.Println("Error QueryRow database:", err_query)
			return false, ""
		}
		//PrintDebug("Exists in db: " + field1)

		decrypted := kes.XorDecrypt(field3, key)
		if decrypted == "" {
			log.Println("Error decrypt xor: empty")
			return false, ""
		}

		return exists, decrypted

	case "A1": //aes128
		//field1 := str[0] //name
		//field2 := str[1] //code
		//field3 := str[2] //cypher //AES/CBC/PKCS7Padding
		//field4 := str[3] //iv     //fixed IV (all zero bytes)

		err_query := db.QueryRow("SELECT EXISTS (SELECT 1 FROM aes_keys WHERE Name = ?), Data FROM aes_keys WHERE Name = ?", field1, field1).Scan(&exists, &key)
		if err_query != nil {
			log.Println("Error QueryRow database:", err_query)
			return false, ""
		}
		PrintDebug("Exists in db: " + field1)

		decrypted, err := decryptAES128(field3, key)
		if err != nil {
			log.Printf("Error decrypt aes128: %v\n", err)
			return false, ""
		}

		return exists, decrypted

	}
	return false, ""
}

func createTables(db *sql.DB) error {

	// Create tables in the database

	sql := `CREATE TABLE IF NOT EXISTS private_keys (
		"Name" TEXT PRIMARY KEY NOT NULL,
		"Data" TEXT,
		"Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS public_keys (
        "Name" TEXT PRIMARY KEY NOT NULL,
        "Data" TEXT,
        "Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS fernet_keys (
        "Name" TEXT PRIMARY KEY NOT NULL,
        "Data" TEXT,
        "Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS aes_keys (
        "Name" TEXT PRIMARY KEY NOT NULL,
        "Data" TEXT,
        "Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS kes_keys (
        "Name" TEXT PRIMARY KEY NOT NULL,
        "Data" TEXT,
        "Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS ubeken_keys (
        "Name" TEXT PRIMARY KEY NOT NULL,
        "Data" TEXT,
        "Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS ips (
        "Name" TEXT PRIMARY KEY NOT NULL,
        "Data" TEXT,
        "Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS procs (
        "Name" TEXT,
        "Data" TEXT,
        "Timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP);`
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func getSqlite3Version(dbFile string) (string, error) {

	// Open the SQLite database from the given path
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return "", fmt.Errorf("Error opening database: %v", err)
	}
	defer db.Close()

	version, err := sqlite3Version(db)
	if err != nil {
		return "", fmt.Errorf("Error getting SQLite version: %v", err)
	}

	return version, nil
}

func sqlite3Version(db *sql.DB) (string, error) {
	var version string
	err := db.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)
	if err != nil {
		return "", err
	}
	return version, nil
}

/*






 */

func decryptRSA(base64Cipher, keyStr string) (string, error) {

	// Decode base64 strings to byte
	ciphertext, err := base64.StdEncoding.DecodeString(base64Cipher)
	if err != nil {
		return "", err
	}

	privateKeyPEM := []byte(keyStr)

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		fmt.Println("Error decoding private key")
		return "", err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("Error parsing private key:", err)
		return "", err
	}

	// Decrypt the data using the private key
	decryptedData, err := rsa.DecryptPKCS1v15(nil, privateKey, ciphertext)
	if err != nil {
		//fmt.Println("Error decrypting rsa:", err)
		return "", err
	}

	return string(decryptedData), nil
}

func decryptFernet(base64Cipher, keyStr string) (string, error) {

	//keyStr := "12345678901234567890123456789012"

	// Encode the key as a base64 string
	base64Key := base64.StdEncoding.EncodeToString([]byte(keyStr))
	//fmt.Println("base64 key: " + base64Key)
	//k := fernet.MustDecodeKeys("cw_0x689RpI-jtRR7oE8h_eQsKImvJapLeSbXpwF4e4=")
	k := fernet.MustDecodeKeys(base64Key)

	//base64tok := os.Args[1]

	// Decode the base64 string
	decodedBytes, err := base64.StdEncoding.DecodeString(base64Cipher)
	if err != nil {
		fmt.Println("Error decoding base64:", err)
		return "", err
	}

	//tokStr := string(tok)
	//fmt.Println("Encrypted: " + tok)

	msg := fernet.VerifyAndDecrypt([]byte(decodedBytes), 60*time.Second, k)

	//fmt.Println(string(msg))

	return string(msg), nil
}

func decryptAESGCM(base64Cipher, base64Nonce, base64Tag, keyStr string) (string, error) {

	// Decode the Base64 strings to []byte
	cipherText, err := base64.StdEncoding.DecodeString(base64Cipher)
	if err != nil {
		log.Println("Error decoding ciphertext:", err)
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(base64Nonce)
	if err != nil {
		log.Println("Error decoding nonce:", err)
		return "", err
	}

	tag, err := base64.StdEncoding.DecodeString(base64Tag)
	if err != nil {
		log.Println("Error decoding tag:", err)
		return "", err
	}

	//key := []byte("YOUR_AES_KEY_HERE")
	key := []byte(keyStr)
	log.Println("keyStr: " + keyStr)

	// Create a new AES block cipher with your key
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("Error creating AES cipher:", err)
		return "", err
	}

	// Create a GCM cipher with the block cipher and the nonce
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Println("Error creating GCM cipher:", err)
		return "", err
	}

	// Decrypt the ciphertext
	plainText, err := aesGCM.Open(nil, nonce, cipherText, tag)
	if err != nil {
		//log.Println("Error decrypting aes:", err)
		return "", err
	}

	// Convert the plaintext to a string and print it
	//log.Println("Decrypted Text:", string(plainText))

	return string(plainText), nil
}

// Decrypts an AES-128 encrypted message using the provided base64 encoded ciphertext and key.
func decryptAES128(base64Cipher, keyStr string) (string, error) {
	// Decode the Base64 string to []byte
	cipherText, err := base64.StdEncoding.DecodeString(base64Cipher)
	if err != nil {
		log.Println("Error decoding ciphertext:", err)
		return "", err
	}

	key := []byte(keyStr)

	// Check if the key size is valid for AES-128 (16 bytes)
	if len(key) != 16 {
		return "", errors.New("Invalid key size. Key size must be 16 bytes.")
	}

	// Create a new AES-128 block cipher with key
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("Error creating AES cipher:", err)
		return "", err
	}

	// Use a fixed IV (all zero bytes) for decryption
	iv := make([]byte, aes.BlockSize)

	// Initialize the decryption mode with the IV
	decrypter := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the ciphertext
	decryptedText := make([]byte, len(cipherText))
	decrypter.CryptBlocks(decryptedText, cipherText)

	// Remove PKCS7 padding (if used during encryption)
	padding := int(decryptedText[len(decryptedText)-1])
	if padding > 0 && padding <= aes.BlockSize {
		decryptedText = decryptedText[:len(decryptedText)-padding]
	}

	// Convert the plaintext to a string and return it
	decryptedStr := string(decryptedText)
	//log.Println("Decrypted Text:", decryptedStr)

	return decryptedStr, nil
}

/*

 */
