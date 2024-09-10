package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func fooHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	fmt.Fprintf(w, `{"hello":true}`)
}

type Animal struct {
	Id   int16
	Name string
}

type AnimalDTO struct {
	Name string
}

var db *sql.DB

func main() {
	dotenvErr := godotenv.Load()

	if dotenvErr != nil {
		log.Fatal(dotenvErr)
	}

	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   os.Getenv("127.0.0.1:3306"),
		DBName: "goapp",
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("DB Connected")

	Setup()

	http.HandleFunc("/animals", PostAnimals)
	http.HandleFunc("/foo", fooHandler)
	http.HandleFunc("/animals/{id}", animalHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func animalHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	animalId, err := strconv.ParseInt(id, 10, 16)

	if err != nil {
		log.Fatal(err)
	}

	var animal Animal

	rows, err := db.Query("SELECT id, name FROM animals WHERE id = ?", animalId)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var id int16
		var name string
		err := rows.Scan(&id, &name)

		if err != nil {
			log.Fatal(err)
		}

		animal = Animal{Id: id, Name: name}
	}

	resp, err := json.Marshal(animal)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, "%s", resp)
}

func PostAnimals(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)

		if err != nil {
			log.Fatal(err)
		}

		var animalDto AnimalDTO

		jsonErr := json.Unmarshal(body, &animalDto)

		if jsonErr != nil {
			log.Fatal(err)
		}

		stmt, prepErr := db.Prepare("INSERT INTO animals (name) VALUES (?)")

		if prepErr != nil {
			log.Fatal(prepErr)
		}

		_, reqErr := stmt.Exec(animalDto.Name)

		if reqErr != nil {
			log.Fatal(reqErr)
		}

		fmt.Fprintf(w, `{"message": "Animal created"}`)
	} else if r.Method == "GET" {
		rows, err := db.Query("SELECT id, name FROM animals")
		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()
		var animals []Animal

		for rows.Next() {
			var id int16
			var name string
			err := rows.Scan(&id, &name)

			if err != nil {
				log.Fatal(err)
			}

			animals = append(animals, Animal{Id: id, Name: name})
		}

		err = rows.Err()

		if err != nil {
			log.Fatal(err)
		}

		responseString, err := json.Marshal(animals)

		if err != nil {
			log.Fatal(err)
		}

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, "%s", responseString)
	} else {
		r.Header.Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"message": "The POST method is required for this URI"}`)
	}
}

func Setup() {
	_, err := db.Exec("CREATE TABLE animals (id MEDIUMINT NOT NULL AUTO_INCREMENT,name CHAR(30) NOT NULL,PRIMARY KEY (id));")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Table miaou created")
}
