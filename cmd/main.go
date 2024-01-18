package main

import (
	"bs"
	"bs/internal/handler"
	"bs/internal/repository"
	"bs/internal/service"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {
	
	if err := initConfig(); err != nil {
		log.Printf("error init config: %s\n", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("Error with loading password %s\n", err.Error())
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})

	if err != nil {
		log.Printf("failed conection with BD %s\n", err.Error())
	}
	err = StartDataBase(db)
	if err != nil {
		log.Printf("failed inserting to BD %s\n", err.Error())
		panic(err)
	}

	err = testIputDB(db)
	if err != nil{
		log.Printf("failed inserting to BD %s\n", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(bs.Server)
	
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Print("Running...")

	ext := make(chan os.Signal, 1)
	signal.Notify(ext, syscall.SIGTERM, syscall.SIGINT)
	<-ext

	log.Print("Stopping...")

	if err = srv.Shutdown(context.Background()); err != nil {
		log.Printf("Server exit: %s\n", err.Error())
	}

	if err = db.Close(); err != nil {
		log.Printf("Error with DataBase: %s\n", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	
	return viper.ReadInConfig()
}

func StartDataBase(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Wallets(
		wallet_id bigint not null,
		currency char(20) default 'RUB' not null,
		value float default 0 not null,
		PRIMARY KEY (wallet_id, currency));
		
		CREATE INDEX ix_wallets_person_id ON Wallets (wallet_id, currency);`)
		
	if err != nil {
		log.Println(err)
	}	
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Transactions(
		id serial not null unique,
		wallet_id bigint not null,
		currency char(100) not null,
		sum float,
		status char(100) default 'Created')`)
	if err != nil {
		log.Println(err)
	}		
	// _, err = db.Exec(`INSERT INTO Wallets (wallet_id) VALUES (5435452135151251)`)
	// if err != nil {
	// 	log.Println(err)
	// }	
	// _, err = db.Exec(`INSERT INTO Wallets (wallet_id) VALUES (1251513451454616)`)
	// if err != nil {
	// 	log.Println(err)
	// }	
	return err
}

func testIputDB(db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO Wallets (wallet_id) VALUES (5435452135151251)`)
	if err != nil {
		log.Println(err)
	}	
	_, err = db.Exec(`INSERT INTO Wallets (wallet_id) VALUES (1251513451454616)`)
	if err != nil {
		log.Println(err)
	}	
	return err
}

