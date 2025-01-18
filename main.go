package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	// Menghubungkan ke MySQL (gunakan root dengan password kosong atau sesuaikan dengan kredensial Anda)
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Membuat database jika belum ada
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS userDB")
	if err != nil {
		log.Fatal("Error creating database: ", err)
	}

	// Menggunakan database yang baru dibuat
	_, err = db.Exec("USE userDB")
	if err != nil {
		log.Fatal("Error selecting database: ", err)
	}

	// Membuat tabel users jika belum ada
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL
		);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}
	addDefaultUsers()
	// Menyiapkan handler untuk halaman login dan register
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	defer db.Close()
}
func addDefaultUsers() {
	// Memeriksa apakah sudah ada data di tabel
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatal("Error checking users count: ", err)
	}

	// Jika tabel kosong, tambahkan pengguna default
	if count == 0 {
		users := []struct {
			username string
			password string
		}{
			{"admin", "admin123"},
			{"123", "asd"},
			{"asd", "123"},
		}

		for _, user := range users {
			_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.username, user.password)
			if err != nil {
				log.Printf("Error inserting user %s: %v", user.username, err)
			}
		}

		fmt.Println("Default users added to the database")
	}
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Menampilkan form login ketika metode GET
		http.ServeFile(w, r, "login.html")
	} else if r.Method == http.MethodPost {
		// Mengambil data dari form
		username := r.FormValue("username")
		password := r.FormValue("password")

		fmt.Printf("Username: %s, Password: %s\n", username, password) // Debugging

		var dbPassword string
		// Query untuk mendapatkan password dari database
		err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				// Jika tidak ditemukan username di database
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			} else {
				// Menangani kesalahan lain (misalnya kesalahan query)
				http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			}
			return
		}

		fmt.Printf("Password from DB: %s\n", dbPassword) // Debugging

		// Membandingkan password yang diinput dengan password di database
		if password != dbPassword {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Jika login berhasil
		fmt.Fprintf(w, "Welcome, %s!", username)
	} else {
		// Menangani metode selain GET atau POST
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Menampilkan halaman registrasi ketika metode GET
		http.ServeFile(w, r, "register.html")
	} else if r.Method == http.MethodPost {
		// Mengambil data dari form
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		// Validasi bahwa password dan konfirmasi password cocok
		if password != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		// Cek apakah username sudah ada di database
		var existingUser string
		err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUser)
		if err == nil {
			// Username sudah terdaftar
			http.Error(w, "Username already exists", http.StatusBadRequest)
			return
		}

		// Menyimpan pengguna baru ke database
		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error registering user: %v", err), http.StatusInternalServerError)
			return
		}

		// Menampilkan pesan sukses atau mengarahkan ke halaman login
		fmt.Fprintf(w, "Registration successful! You can now <a href='/login'>login</a>.")

		http.HandleFunc("/login", loginHandler)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}
