package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type Customer struct {
	ID        int    `json:"customerId"`
	Name      string `json:"customerName"`
	HandPower int    `json:"handPower"`
	LegPower  int    `json:"legPower"`
	Iq        int    `json:"iq"`
}

type CustomerUpdate struct {
	CustomerID int    `json:"customerId"`
	UpdatedParameter string `json:"updatedParameter"`
	Class string `json::"@class"`
}


func ConnectToDatabase() (*sql.DB, error) {
	dsn := "host=localhost port=5434 user=postgres password=postgres dbname=training_center_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InitCustomersTable(db *sql.DB) error {
    const createTableQuery = `
    CREATE TABLE IF NOT EXISTS customers (
        customer_id INT PRIMARY KEY,
        hand_power INT,
        iq INT,
        leg_power INT,
        name VARCHAR(255)
    );`
    _, err := db.Exec(createTableQuery)
    return err
}


func SendToKafka(customers []Customer) error {
	topic := "training-updates"
	partition := 0
	
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:29092", topic, partition)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	defer conn.Close()

	for _, customer := range customers {
		sendCustomer := CustomerUpdate{
			CustomerID: customer.ID,
			UpdatedParameter: "HandPower",
			Class: "com.trainingcenter.kafka.CustomerUpdate",
		}
		data, err := json.Marshal(sendCustomer)
		if err != nil {
			return fmt.Errorf("failed to marshal customer data: %w", err)
		}

		_, err = conn.WriteMessages(
			kafka.Message{
				Value: data,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to write message to Kafka: %w", err)
		}
	}

	return nil
}

func GetFromKafka() ([]Customer, error) {
	topic := "customers"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:29092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	batch := conn.ReadBatch(10e3, 1e6)

	b := make([]byte, 10e3)
	var customers []Customer
	for {
		n, err := batch.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}

		var customer Customer
		if err := json.Unmarshal(b[:n], &customer); err != nil {
			log.Printf("failed to unmarshal json: %v", err)
			continue
		}
		customers = append(customers, customer)
	}

	batch.Close()

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close connection:", err)
	}

	return customers, nil
}

func main() {
	
	customers, err := GetFromKafka()
	if err != nil {
		log.Fatal("failed to load customers from Kafka:", err)
	}

	fmt.Println(customers)

	db, err := ConnectToDatabase()
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	if err := InitCustomersTable(db); err != nil {
		log.Fatal("failed to initialize customers table:", err)
	}

	for _, customer := range customers {
		const insertQuery = `
			INSERT INTO customers (customer_id, hand_power, iq, leg_power, name)	
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (customer_id) DO 
			UPDATE SET 
			hand_power = EXCLUDED.hand_power,
			iq = EXCLUDED.iq,
			leg_power = EXCLUDED.leg_power,
			name = EXCLUDED.name;
			`
		_, err := db.Exec(insertQuery, customer.ID, customer.HandPower, customer.Iq, customer.LegPower, customer.Name)
		if err != nil {
			log.Printf("failed to insert customer %d: %v", customer.ID, err)
			continue
		}
	}

	if err := SendToKafka(customers); err != nil {
		log.Fatal("failed to send to Kafka:", err)
	}





	fmt.Println("Done!")
}
